package itf

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshallBasicTypes(t *testing.T) {
	var expr Expr
	var err error

	err = json.Unmarshal([]byte(`false`), &expr)
	assert.Nil(t, err)
	assert.Equal(t, false, expr.Value)

	err = json.Unmarshal([]byte(`"abc"`), &expr)
	assert.Nil(t, err)
	assert.Equal(t, "abc", expr.Value)

	err = json.Unmarshal([]byte(`123`), &expr)
	assert.Nil(t, err)
	assert.Equal(t, float64(123), expr.Value)

	// TODO: bigint
}

func TestUnmarshallList(t *testing.T) {
	var expr Expr
	err := json.Unmarshal([]byte(`[123, "abc"]`), &expr)
	assert.Nil(t, err)

	elems := expr.Value.(ListExprType)
	assert.NotNil(t, elems)
	assert.Equal(t, float64(123), elems[0].Value)
	assert.Equal(t, "abc", elems[1].Value)
}

func TestUnmarshallMap(t *testing.T) {
	var expr Expr
	err := json.Unmarshal([]byte(`{ "#map": [["foo", "abc"], ["bar", { "#map": [["baz", 123]]}]] }`), &expr)
	assert.Nil(t, err)

	map1, ok := expr.Value.(MapExprType)
	assert.True(t, ok)
	assert.Equal(t, "abc", map1["foo"].Value)

	map2 := map1["bar"].Value.(MapExprType)
	assert.NotNil(t, map2)
	assert.Equal(t, float64(123), map2["baz"].Value)
}

func TestToString(t *testing.T) {
	keyMap := map[string]string{"a": "b", "c": "d"}
	key := toString(keyMap)
	assert.Equal(t, "abcd", key)
}

func TestUnmarshallMapWithExprKey(t *testing.T) {
	var expr Expr
	err := json.Unmarshal([]byte(`{"#map": [ [ { "a": "b", "c": "d" }, { "#tup": [1, "foo"] } ] ]}`), &expr)
	assert.Nil(t, err)

	map1, ok := expr.Value.(MapExprType)
	assert.True(t, ok)
	assert.Equal(t, 1, len(map1))

	key := toString(map[string]string{"a": "b", "c": "d"})
	val := map1[key].Value.(ListExprType)
	assert.Equal(t, float64(1), val[0].Value)
	assert.Equal(t, "foo", val[1].Value)
}

func TestUnmarshallTuple(t *testing.T) {
	var expr Expr
	err := json.Unmarshal([]byte(`{ "#tup": ["foo", true] }`), &expr)
	assert.Nil(t, err)

	elems := expr.Value.(ListExprType)
	assert.NotNil(t, elems)
	assert.Equal(t, "foo", elems[0].Value)
	assert.Equal(t, true, elems[1].Value)
}

func TestUnmarshallSet(t *testing.T) {
	var expr Expr
	err := json.Unmarshal([]byte(`{ "#set": ["foo", "bar", "baz"] }`), &expr)
	assert.Nil(t, err)

	elems := expr.Value.(ListExprType)
	assert.NotNil(t, elems)
	assert.Equal(t, 3, len(elems))
	assert.Equal(t, "foo", elems[0].Value)
	assert.Equal(t, "bar", elems[1].Value)
	assert.Equal(t, "baz", elems[2].Value)
}

func TestUnmarshallBigInt(t *testing.T) {
	var expr Expr
	err := json.Unmarshal([]byte(`{ "#bigint": "-1234567890" }`), &expr)
	assert.Nil(t, err)

	bigint := expr.Value.(int64)
	assert.NotNil(t, bigint)
	assert.Equal(t, int64(-1234567890), bigint)
}

func TestUnmarshallRecord(t *testing.T) {
	var expr Expr
	err := json.Unmarshal([]byte(`{ "x": true, "y": { "#map": [["z", "abc"]] } }`), &expr)
	assert.Nil(t, err)

	elems := expr.Value.(MapExprType)
	assert.NotNil(t, elems)
	assert.Equal(t, true, elems["x"].Value)

	map_ := elems["y"].Value.(MapExprType)
	assert.NotNil(t, map_)
	assert.Equal(t, "abc", map_["z"].Value)
}

func TestUnmarshallState(t *testing.T) {
	data := []byte(`{"#meta": { "index": 0 }, "var1": "abc", "var2": true}`)

	var state State
	err := json.Unmarshal(data, &state)
	assert.Nil(t, err)
	assert.Equal(t, "abc", state.VarValues["var1"].Value)
	assert.Equal(t, true, state.VarValues["var2"].Value)
}

func TestUnmarshallRecordAsMapKey(t *testing.T) {
	data := []byte(`
	{
			"currentState": {
				"#map": [
					[
						{
							"validatorSet": {
								"#map": [
									[
										"node1",
										{
											"#bigint": "100"
										}
									]
								]
							}
						},
						{
							"#bigint": "1209600"
						}
					]
				]
			}
	}`)

	var expr Expr
	err := json.Unmarshal(data, &expr)
	assert.Nil(t, err)

	states := expr.Value.(MapExprType)
	assert.NotNil(t, states)

	state := states["currentState"].Value.(MapExprType)
	assert.NotNil(t, state)
}

func TestUnmarshallRecordWithIntsAsMapKey(t *testing.T) {
	data := []byte(`
	{
	"mapping": {
		"#map": [
			[
				{
					"a": {
						"#bigint": "1"
					},
					"b": {
						"#bigint": "2"
					}
				},
				{
					"#bigint": "3"
				}
			]
		]
	}
	}`)

	var expr Expr
	err := json.Unmarshal(data, &expr)
	assert.Nil(t, err)

	mapping := expr.Value.(MapExprType)
	assert.NotNil(t, mapping)

	map_ := mapping["mapping"].Value.(MapExprType)
	assert.NotNil(t, map_)

	for key := range map_ {
		assert.NotEqual(t, key, "a#bigint1b#bigint2")
	}
}

func TestUnmarshallTrace(t *testing.T) {
	data := []byte(`{
		"#meta": {
			"format": "ITF",
			"format-description": "https://apalache.informal.systems/docs/adr/015adr-trace.html",
			"source": "spec.qnt",
			"status": "ok",
			"description": "Created by Quint on ...",
			"timestamp": 1686147869875
		},
		"vars": [ "var1", "var2", "var3" ],
		"states": [
			{
				"#meta": { "index": 0 },
				"var1": "foo",
				"var2": { "#map": [
					["n1", { "#map": [] }],
					["n2", { "#map": [ ["foo", "bar"] ] }]
				]},
				"var3": { "#set": [1, 2, 3] }
			}
		]
	}`)

	var trace Trace
	err := json.Unmarshal(data, &trace)
	assert.Nil(t, err)
}
