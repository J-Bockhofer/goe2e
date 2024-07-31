package goe2e_test

import (
	"encoding/json"
	"testing"

	goe2e "github.com/J-Bockhofer/goe2e/pkg"

	"github.com/stretchr/testify/assert"
)

func TestValueInBodyByKey(t *testing.T) {
	key := "pname"
	name := "jamie"

	bodyBytes := []byte(`{"type":"person","data":{"pname":"jamie","age":24}}`)
	var bodyMap goe2e.H
	err := json.Unmarshal(bodyBytes, &bodyMap)
	if err != nil {
		t.Errorf("failed to unmarshal: %s", err.Error())
	}

	testCases := []struct {
		description string
		body        goe2e.H
		key         string
		expected    interface{}
	}{
		{
			"From bytes",
			bodyMap,
			key,
			name,
		},
		{
			"Direct map",
			goe2e.H{
				"type": "person",
				"data": goe2e.H{
					key: name, "age": 24,
				},
			},
			key,
			name,
		},
		{
			"Nested map",
			goe2e.H{
				"type": "person",
				"data": goe2e.H{
					"gname": "gamie", "age": 32,
				},
				"data2": goe2e.H{
					key: name, "age": 24,
				},
			},
			key,
			name,
		},
		{
			"Nested map not found",
			goe2e.H{
				"type": "person",
				"data": goe2e.H{
					"gname": "gamie", "age": 32,
				},
				"data2": goe2e.H{
					"hname": "jamie", "age": 24,
				},
			},
			key,
			nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.description, func(t *testing.T) {
			name := goe2e.ValueInMapByKey(tt.key, tt.body)
			assert.Equal(t, tt.expected, name)
		})
	}
}

func TestValueToBodyByKey(t *testing.T) {
	key := "pname"
	name := "jamie"

	bodyBytes := []byte(`{"type":"person","data":{"pname":"gnark","age":24}}`)
	var bodyMap goe2e.H
	err := json.Unmarshal(bodyBytes, &bodyMap)
	if err != nil {
		t.Errorf("failed to unmarshal: %s", err.Error())
	}

	testCases := []struct {
		description string
		body        goe2e.H
		key         string
		value       string
		shouldSet   bool
	}{
		{
			"From bytes",
			bodyMap,
			key,
			name,
			true,
		},
		{
			"Direct map",
			goe2e.H{
				"type": "person",
				"data": goe2e.H{
					key: "name", "age": 24,
				},
			},
			key,
			name,
			true,
		},
		{
			"Nested map",
			goe2e.H{
				"type": "person",
				"data": goe2e.H{
					"gname": "gamie", "age": 32,
				},
				"data2": goe2e.H{
					key: "name", "age": 24,
				},
			},
			key,
			name,
			true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.description, func(t *testing.T) {
			wasSet := goe2e.ValueToMapByKey(tt.key, tt.value, tt.body)
			assert.Equal(t, tt.shouldSet, wasSet)
			setVal := goe2e.ValueInMapByKey(tt.key, tt.body)
			assert.Equal(t, tt.value, setVal)
		})
	}
}

func TestAssembleQuery(t *testing.T) {
	testCases := []struct {
		description   string
		route         string
		queryMap      goe2e.D
		expectedRoute string
	}{
		{
			"simple",
			"route/persons",
			goe2e.D{
				"val1": "1",
				"val2": "2",
			},
			"route/persons?val1=1&val2=2",
		},
		{
			"with invalid char",
			"route/persons",
			goe2e.D{
				"val1": "&",
				"val2": "2",
			},
			"route/persons?val1=%26&val2=2",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.description, func(t *testing.T) {
			actual := goe2e.AssembleQuery(tt.route, tt.queryMap)
			assert.Equal(t, tt.expectedRoute, actual)
		})
	}

}
