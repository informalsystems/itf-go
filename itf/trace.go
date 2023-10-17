package itf

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
)

// ITF spec: https://apalache.informal.systems/docs/adr/015adr-trace.html#the-itf-format
type Trace struct {
	Meta   *meta            `json:"#meta,omitempty"`
	Params map[string]*Expr `json:"params"`
	Vars   []string         `json:"vars"`
	States []*State         `json:"states"`
	Loop   int              `json:"loop,omitempty"`
}

type meta struct {
	Description string `json:"description,omitempty"`
	Source      string `json:"source,omitempty"`
}

type State struct {
	Meta      map[string]any   `json:"#meta,omitempty"`
	VarValues map[string]*Expr `json:",omitempty"`
}

func (s *State) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}
	var objmap map[string]json.RawMessage
	if err := json.Unmarshal(data, &objmap); err != nil {
		return err
	}

	var meta map[string]any
	values := make(map[string]*Expr)
	for k, v := range objmap {
		if k == "#meta" {
			if err := json.Unmarshal(v, &meta); err != nil {
				return err
			}
		} else {
			var expr Expr
			if err := json.Unmarshal(v, &expr); err != nil {
				return err
			}
			values[k] = &expr
		}
	}

	*s = State{Meta: meta, VarValues: values}

	return nil
}

type Expr struct {
	Value any
}

type (
	ListExprType = []Expr
	MapExprType  = map[string]Expr
)

func toExpr(parsed any) (Expr, error) {
	switch val := parsed.(type) {

	case nil:
		return Expr{}, fmt.Errorf("null value not allowed")

	case bool, string, float64:
		return Expr{val}, nil

	case []any:
		var list ListExprType
		for _, v := range val {
			if exp, err := toExpr(v); err != nil {
				return Expr{}, err
			} else {
				list = append(list, exp)
			}
		}
		return Expr{list}, nil

	case map[string]any:
		mapContent, is_map := val["#map"].([]any)
		if len(val) == 1 && is_map {
			map_ := make(MapExprType)
			for _, p := range mapContent {
				pair, ok := p.([]any)
				if !ok {
					return Expr{}, fmt.Errorf("map entry is not a key/value array: %v", p)
				}
				if valExpr, err := toExpr(pair[1]); err != nil {
					return Expr{}, err
				} else {
					key := toString(pair[0])
					map_[key] = valExpr
				}
			}
			return Expr{map_}, nil
		}

		mapContent, is_tuple := val["#tup"].([]any)
		if len(val) == 1 && is_tuple {
			var elements ListExprType
			for _, elem := range mapContent {
				if exp, err := toExpr(elem); err != nil {
					return Expr{}, err
				} else {
					elements = append(elements, exp)
				}
			}
			return Expr{elements}, nil
		}

		mapContent, is_set := val["#set"].([]any)
		if len(val) == 1 && is_set {
			var elements ListExprType
			for _, elem := range mapContent {
				if exp, err := toExpr(elem); err != nil {
					return Expr{}, err
				} else {
					elements = append(elements, exp)
				}
			}
			return Expr{elements}, nil
		}

		bigintContent, is_bigint := val["#bigint"].(string)
		if is_bigint {
			bigint, err := strconv.ParseInt(bigintContent, 10, 64)
			if err != nil {
				panic(err)
			}
			return Expr{bigint}, nil
		}

		// Otherwise, it's a record.
		// "Field names should not start with # and hence should not pose any collision with other constructs."
		record := make(MapExprType)
		for key, val := range val {
			if valExpr, err := toExpr(val); err != nil {
				return Expr{}, err
			} else {
				record[toString(key)] = valExpr
			}
		}
		return Expr{record}, nil

	default:
		return Expr{}, fmt.Errorf("type %T unexpected", parsed)
	}
}

func toString(x any) string {
	switch x := x.(type) {
	case nil:
		return ""
	case string:
		return x
	case map[string]string:
		return mapToString(x)
	case map[string]any:
		return mapToString(x)
	default:
		return x.(string)
	}
}

// TODO: fix collisions
func mapToString[T any](x map[string]T) string {
	s := ""
	keys := make([]string, len(x))
	for k := range x {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		s += k + toString(x[k])
	}
	return s
}

func (e *Expr) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	var parsed any
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}

	if expr, err := toExpr(parsed); err != nil {
		return err
	} else {
		*e = expr
	}

	return nil
}

func (trace *Trace) LoadFromFile(filePath string) error {
	// If file doesn't exist, do nothing.
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return err
	}

	// Load file
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Unmarshal content
	if err := json.NewDecoder(file).Decode(trace); err != nil {
		return err
	}

	return nil
}
