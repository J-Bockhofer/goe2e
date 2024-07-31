package goe2e

import (
	"encoding/json"
	"fmt"
)

// SpecOption is used to initialize optional fields of a Spec via the constructor NewSpec.
// Use it to define your own Spec modifications or use one of the predefined "With" functions.
type SpecOption func(*Spec) error

// WithBody sets the request's body.
func WithBody(body []byte) SpecOption {
	return func(rs *Spec) error {
		rs.Body = body
		return nil
	}
}

// WithUrl sets the request's url.
func WithUrl(url string) SpecOption {
	return func(rs *Spec) error {
		rs.Url = url
		return nil
	}
}

// WithBaseURLFromEnv attempts to build the request's url from a domain name and route.
// The domain name literal is stored in a map called env, where the urlKey is the key under which it is stored.
func WithBaseURLFromEnv(env H, urlKey string, route string) SpecOption {
	return func(rs *Spec) error {
		baseUrl := ValueInMapByKey(urlKey, env)
		if baseUrl == nil {
			return fmt.Errorf("baseURL not found: key %s not found in env", urlKey)
		}
		switch t := baseUrl.(type) {
		case string:
			rs.Url = JoinAsRoute(t, route)
		default:
			return fmt.Errorf("baseURL in env not of type string: %v", baseUrl)
		}
		return nil
	}
}

// WithMethod sets the http method for the request.
// Best used with the constants from the net/http package, http.MethodGet etc.
// Does no validity checking on the passed string.
func WithMethod(method string) SpecOption {
	return func(rs *Spec) error {
		rs.Method = method
		return nil
	}
}

// WithJSON tries to marshal the given payload and sets it as the request body.
func WithJSON(payload interface{}) SpecOption {
	return func(rs *Spec) error {
		b, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("spec option WithJSON failed - json.Marshal: %s", err.Error())
		}
		rs.Body = b
		return nil
	}
}

// WithSetFromEnv tries to unmarshal the existing request body into a map and then set fields of the map to values from the passed env map.
// Optionally takes a keymap to allow for different field naming in env and json body.
func WithSetFromEnv(env H, keymap D) SpecOption {
	return func(rs *Spec) error {
		var body H
		err := json.Unmarshal(rs.Body, &body)
		if err != nil {
			return fmt.Errorf("spec option WithSetFromEnv failed - json.Unmarshal: %s", err.Error())
		}
		for k, v := range env {
			keyInBody := k
			if keymap != nil {
				var ok bool
				keyInBody, ok = keymap[k]
				if !ok {
					keyInBody = k
				}
			}
			ValueToMapByKey(keyInBody, v, body)
		}
		err = WithJSON(body)(rs)
		if err != nil {
			return fmt.Errorf("spec option WithSetFromEnv failed - Setting body: %s", err.Error())
		}
		return nil
	}
}

// WithRouteFromQueryMap takes a route and assembles the query from a map.
// The route should include the domain.
func WithRouteFromQueryMap(route string, queryMap D) SpecOption {
	return func(rs *Spec) error {
		rs.Url = AssembleQuery(route, queryMap)
		return nil
	}
}

// AddQueryFromMap assembles the query string from the given parameter map and takes the Spec.Url as the route.
func AddQueryFromMap(queryMap D) SpecOption {
	return func(rs *Spec) error {
		rs.Url = AssembleQuery(rs.Url, queryMap)
		return nil
	}
}
