package goe2e

import (
	"net/http"
)

// ResponseModifier is used to modify the http.Response before parsing the body.
type ResponseModifier func(*http.Response) error

// ResponseBodyModifier is used to modify the response body before running tests.
// It is separate to the ResponseModifier because once the io.ReadCloser of the http.Response is read, there is no putting the body back in.
type ResponseBodyModifier func([]byte) ([]byte, error)

// ResponseJSONToEnv parses the ResponseBody (JSON) into a map[string]interface{}.
// Optionally takes a keymap (or nil) that holds a mapping from the key literal to a key literal in the json encoded body.
// If the env key is not found in the keymap, the env key itself is used instead, as if not passing a keymap at all.
func ResponseJSONToEnv(env H, keymap D) ResponseBodyModifier {
	return func(body []byte) ([]byte, error) {
		bodyMap, err := bodyJSONToMap(body)
		if err != nil {
			return nil, err
		}
		for k := range env {
			keyInBody := k
			if keymap != nil {
				var ok bool
				keyInBody, ok = keymap[k]
				if !ok {
					keyInBody = k
				}
			}
			valInBody := ValueInMapByKey(keyInBody, bodyMap)
			if valInBody == nil {
				continue
			}
			env[k] = valInBody
		}
		return body, nil
	}
}
