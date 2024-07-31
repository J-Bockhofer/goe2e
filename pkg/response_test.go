package goe2e_test

import (
	"testing"

	goe2e "github.com/J-Bockhofer/goe2e/pkg"

	"github.com/stretchr/testify/assert"
)

func TestResponseJSONToEnv(t *testing.T) {
	r, err := goe2e.NewRequestHandler()
	if err != nil {
		t.Errorf("failed to make new request %s", err.Error())
	}
	env := map[string]interface{}{
		"name": "",
		"type": 0,
	}
	keymap := map[string]string{
		"name": "pname",
	}
	testCases := []struct {
		name     string
		body     []byte
		keymap   goe2e.D
		keyInEnv string
		expected string
	}{
		{"Direct env", []byte(`{"name":"john","type":2}`), nil, "name", "john"},
		{"Mapped env", []byte(`{"pname":"jamie","type":2}`), keymap, "name", "jamie"},
		{"Nested Response to Mapped Env", []byte(`{"type":"person","data":{"pname":"jamie","age":24}}`), keymap, "name", "jamie"},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			r.ResponseBody = tt.body
			mod := goe2e.ResponseJSONToEnv(env, tt.keymap)
			err = r.ModifyResponseBody(mod)
			if err != nil {
				t.Errorf("failed to modify responseBody %s", err.Error())
			}
			assert.Equal(t, tt.expected, env[tt.keyInEnv])
		})
	}
}
