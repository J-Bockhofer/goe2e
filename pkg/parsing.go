package goe2e

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ValueInMapByKey takes a key and attempts to recursively find the corresponding value in a potentially nested map.
// Returns nil if nothing is found.
// Searches depth first.
func ValueInMapByKey(key string, body H) interface{} {
	valInBody, ok := body[key]
	if ok {
		return valInBody
	}
	// check for nested map
	for _, v := range body {
		switch t := v.(type) {
		case H:
			pv := ValueInMapByKey(key, t)
			if pv != nil {
				return pv
			}
		default:
			continue
		}
	}
	return nil
}

// ValueToMapByKey takes a key and attempts to recursively find the corresponding entry in a potentially nested map and set its value.
// Returns false if nothing was set.
// Searches depth first.
func ValueToMapByKey(key string, val interface{}, body H) bool {
	_, ok := body[key]
	if ok {
		body[key] = val
		return true
	}
	// check for nested map
	for _, v := range body {
		switch t := v.(type) {
		case H:
			found := ValueToMapByKey(key, val, t)
			if found {
				return true
			}
		default:
			continue
		}
	}
	return false
}

// AssembleQuery helps in building complex query strings.
// It returns the route if the passed map is nil or empty.
func AssembleQuery(route string, queryMap D) string {
	if len(queryMap) == 0 {
		return route
	}
	if !strings.HasSuffix(route, "?") {
		route += "?"
	}
	for k, v := range queryMap {
		k = queryEncodeCharsInString(k)
		v = queryEncodeCharsInString(v)
		route += k + "=" + v + "&"
	}
	route = strings.TrimSuffix(route, "&")
	return route
}

// https://stackoverflow.com/questions/2366260/whats-valid-and-whats-not-in-a-uri-query
var queryCharsInvalidSet = []rune{
	'%',
	'"',
	' ',
	'\\',
	'^',
	'`',
	'{',
	'|',
	'}',
	'#',
	'~',
	'&',
	'=',
}

func encodeURIComponent(r rune) string {
	if isInSet(r, queryCharsInvalidSet) {
		return fmt.Sprintf("%%%X", r)
	}
	return string(r)
}

func isInSet(r rune, set []rune) bool {
	for _, v := range set {
		if r == v {
			return true
		}
	}
	return false
}

func queryEncodeCharsInString(str string) string {
	var builder strings.Builder
	for _, char := range str {
		s := encodeURIComponent(char)
		builder.WriteString(s)
	}
	return builder.String()
}

func bodyJSONToMap(b []byte) (H, error) {
	var bodyMap H
	err := json.Unmarshal(b, &bodyMap)
	if err != nil {
		return nil, err
	}
	return bodyMap, nil
}
