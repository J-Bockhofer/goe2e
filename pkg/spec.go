package goe2e

import (
	"bytes"
	"net/http"
)

// Spec specifies the parameters for a http.Request.
// It generates the request at the end of the constructor NewSpec.
// Once the request is generated you cant modify it easily.
type Spec struct {
	Method  string
	Url     string
	Body    []byte
	Request *http.Request
}

// NewSpec constructs a new Spec from SpecOption functions.
// Will also construct the http.Request. To further modify the http.Request use a RequestModifier.
// To modify the request body use a RequestSpecModifier, which will regenerate the http.Request.
func NewSpec(opts ...SpecOption) (*Spec, error) {
	r := &Spec{
		Method:  http.MethodGet,
		Url:     "https://example.com",
		Body:    make([]byte, 0),
		Request: nil,
	}
	for _, opt := range opts {
		err := opt(r)
		if err != nil {
			return nil, err
		}
	}
	err := r.generateRequest()
	if err != nil {
		return nil, err
	}
	return r, nil
}

// generateRequest creates the http.Request.
func (rs *Spec) generateRequest() error {
	req, err := http.NewRequest(rs.Method, rs.Url, bytes.NewReader(rs.Body))
	if err != nil {
		return err
	}
	rs.Request = req
	return nil
}
