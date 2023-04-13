package llb

import (
	"io"
	"net/http"
)

type (
	mockAPI struct {
		_getRuntimeInvocationNext      func() (resp *http.Response, err error)
		_postRuntimeInitError          func(err error) (*http.Response, error)
		_postRuntimeInvocationError    func(requestId string, err error) (*http.Response, error)
		_postRuntimeInvocationResponse func(requestId string, response io.Reader) (*http.Response, error)
	}
)

var (
	_ = api(mockAPI{})
)

func (api mockAPI) getRuntimeInvocationNext() (*http.Response, error) {
	return api._getRuntimeInvocationNext()
}

func (api mockAPI) postRuntimeInitError(err error) (*http.Response, error) {
	return api._postRuntimeInitError(err)
}

func (api mockAPI) postRuntimeInvocationError(requestId string, err error) (*http.Response, error) {
	return api._postRuntimeInvocationError(requestId, err)
}

func (api mockAPI) postRuntimeInvocationResponse(requestId string, response io.Reader) (*http.Response, error) {
	return api._postRuntimeInvocationResponse(requestId, response)
}
