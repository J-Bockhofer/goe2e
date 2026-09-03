package goe2e

import (
	"encoding/json"
	"fmt"
)

// ResponseJSONPointer returns the response value selected by an RFC 6901 JSON Pointer.
func (rh *RequestHandler) ResponseJSONPointer(pointer string) (any, error) {
	if rh == nil {
		return nil, fmt.Errorf("response handler must not be nil")
	}
	var document any
	if err := json.Unmarshal(rh.ResponseBody, &document); err != nil {
		return nil, fmt.Errorf("parse response JSON: %w", err)
	}
	return jsonPointerValue(document, pointer)
}

// ResponseJSONPointerString returns a string response value selected by an RFC 6901 JSON Pointer.
func (rh *RequestHandler) ResponseJSONPointerString(pointer string) (string, error) {
	value, err := rh.ResponseJSONPointer(pointer)
	if err != nil {
		return "", err
	}
	stringValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("response JSON pointer %q resolved to %T, not string", pointer, value)
	}
	return stringValue, nil
}
