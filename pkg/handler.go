package goe2e

import (
	"fmt"
	"io"
	"net/http"
)

// RequestHandler handles construction, modification and execution of a http.Request.
// After running a request the response is immediately stored in the ResponseBody field for easy modification and assertions.
type RequestHandler struct {
	spec         *Spec
	Client       *http.Client
	Response     *http.Response
	ResponseBody []byte
}

type RequestHandlerOption func(*RequestHandler) error

// NewRequestHandler is used to initialize a RequestHandler, with optional methods that modify the constructor.
// It wraps the initializer of a Spec, which will, in turn, generate the http.Request.
// You can call (RequestHandler).Exec right after.
func NewRequestHandler(opts ...RequestHandlerOption) (*RequestHandler, error) {
	spec, err := NewSpec()
	if err != nil {
		return nil, err
	}
	rh := &RequestHandler{
		spec:         spec,
		Client:       nil,
		Response:     nil,
		ResponseBody: nil,
	}
	for _, opt := range opts {
		err := opt(rh)
		if err != nil {
			return nil, err
		}
	}
	return rh, nil
}

func WithSpec(rs *Spec) RequestHandlerOption {
	return func(rh *RequestHandler) error {
		rh.spec = rs
		return nil
	}
}

// WithSpecOpts constructs a RequestHandler from SpecOptions, the handler will then be ready to perform the request.
func WithSpecOpts(specOpts ...SpecOption) RequestHandlerOption {
	return func(rh *RequestHandler) error {
		rs, err := NewSpec(specOpts...)
		if err != nil {
			return err
		}
		rh.spec = rs
		return nil
	}
}

// WithClient passes a re-usable client to run the request.
func WithClient(client *http.Client) RequestHandlerOption {
	return func(rh *RequestHandler) error {
		rh.Client = client
		return nil
	}
}

// RunRequest will execute the http.Request and write the response to the RequestHandler.ResponseBody.
func (rh *RequestHandler) RunRequest() error {
	if rh.spec == nil {
		return fmt.Errorf("no request specifications initialized before executing")
	}
	if rh.Client == nil {
		rh.Client = &http.Client{}
	}
	resp, err := rh.Client.Do(rh.spec.Request)
	if err != nil {
		return err
	}
	// didnt get a body, which is fine
	if resp.Body == nil {
		rh.Response = resp
		return nil
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	rh.ResponseBody = b
	rh.Response = resp
	return nil
}

// Close closes the response body. "You should never need it, but it is here" - Justin Case.
func (rh *RequestHandler) Close() error {
	if rh.Response != nil {
		return rh.Response.Body.Close()
	}
	return nil
}

// GetRequest returns the http.Request for inspection during assertions.
// It should not be used to try and modify the request body, as this should be left to the Spec constructor.
func (rh *RequestHandler) GetRequest() *http.Request {
	return rh.spec.Request
}

func (rh *RequestHandler) ModifyRequest(respBodyOpts ...RequestModifier) error {
	if rh.spec.Request == nil {
		return fmt.Errorf("no http.Request generated before modifying it")
	}
	for _, opt := range respBodyOpts {
		err := opt(rh.spec.Request)
		if err != nil {
			return fmt.Errorf("modifying http.Request failed: %w", err)
		}
	}
	return nil
}

// ModifyResponseBody allows for passing a set of functions that sequentially modify the ResponseBody.
func (rh *RequestHandler) ModifyResponseBody(respBodyOpts ...ResponseBodyModifier) error {
	if rh.ResponseBody == nil {
		return fmt.Errorf("no ResponseBody set before modifying it")
	}
	for _, opt := range respBodyOpts {
		b, err := opt(rh.ResponseBody)
		if err != nil {
			return fmt.Errorf("modifying response body failed: %w", err)
		}
		rh.ResponseBody = b
	}
	return nil
}

// ModifyResponse allows for passing a set of functions that change the http.Response directly.
// Should not be used for modifying the response body, use (RequestHandler).ModifyResponseBody instead.
func (rh *RequestHandler) ModifyResponse(respOpts ...ResponseModifier) error {
	if rh.Response == nil {
		return fmt.Errorf("no response gathered before modifying it")
	}
	for _, opt := range respOpts {
		err := opt(rh.Response)
		if err != nil {
			return fmt.Errorf("modifying response failed: %w", err)
		}
	}
	return nil
}

type RequestHandlerModifier interface {
	Apply(*RequestHandler) error
}

type RequestHandlerModFunc func(*RequestHandler) error

func (hmf RequestHandlerModFunc) Apply(rh *RequestHandler) error {
	return hmf(rh)
}

func RequestHandlerNoop() RequestHandlerModFunc {
	return func(rh *RequestHandler) error {
		return nil
	}
}
