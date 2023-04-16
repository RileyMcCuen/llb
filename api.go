package llb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type (
	defaultAPI struct {
		domain              string
		invocationUrlPrefix string
		nextUrl             string
		initErrorUrl        string
		client              *http.Client
	}
	api interface {
		getRuntimeInvocationNext() (resp *http.Response, err error)
		postRuntimeInitError(err error) (*http.Response, error)
		postRuntimeInvocationError(requestId string, err error) (*http.Response, error)
		postRuntimeInvocationResponse(requestId string, response io.Reader) (*http.Response, error)
	}
)

const (
	envRuntimeDomain = "AWS_LAMBDA_RUNTIME_API"

	headerContentType  = "Content-Type"
	defaultContentType = "application/json"

	headerErrorType          = "Lambda-Runtime-Function-Error-Type"
	defaultInitErrorHeader   = "Runtime.InitError"
	defaultInvokeErrorHeader = "Runtime.InvokeError"
)

var (
	_ = api(&defaultAPI{})
)

func newDefaultAPI(client *http.Client) defaultAPI {
	domain := os.Getenv(envRuntimeDomain)

	return defaultAPI{
		domain:              domain,
		invocationUrlPrefix: "http://" + domain + "/2018-06-01/runtime/invocation/",
		nextUrl:             "http://" + domain + "/2018-06-01/runtime/invocation/next",
		initErrorUrl:        "http://" + domain + "/2018-06-01/runtime/init/error",
		client:              client,
	}
}

func (api defaultAPI) getRuntimeInvocationNext() (*http.Response, error) {
	resp, err := api.client.Get(api.nextUrl)
	if err != nil {
		return resp, fmt.Errorf("%w; defaultAPI.getRuntimeInvocationNext", err)
	}

	return resp, nil
}

func (api defaultAPI) postRuntimeInitError(err error) (*http.Response, error) {
	log.Println("defaultAPI.postRuntimeInitError", err)

	header := defaultInitErrorHeader
	payload := struct {
		Message    string   `json:"errorMessage"`
		Type       string   `json:"errorType"`
		StackTrace []string `json:"stackTrace"`
	}{
		Message:    err.Error(),
		Type:       defaultInitErrorHeader,
		StackTrace: []string{},
	}

	if err, ok := err.(Error); ok {
		header = err.Header()
		payload.Type = err.Type()
	}

	body, _ := json.Marshal(payload)

	request, _ := http.NewRequest(
		http.MethodPost,
		api.initErrorUrl,
		bytes.NewBuffer(body),
	)
	request.Header.Add(headerErrorType, header)

	resp, err := api.client.Do(request)
	if err != nil {
		return resp, fmt.Errorf("%w; error submitting defaultAPI.postRuntimeInitError request", err)
	} else {
		msg, _ := io.ReadAll(resp.Body)

		switch resp.StatusCode {
		case http.StatusAccepted:
			return resp, fmt.Errorf("accepted status code (202) defaultAPI.postRuntimeInitError\n%s", string(msg))
		case http.StatusForbidden:
			return resp, fmt.Errorf("forbidden status code (403) defaultAPI.postRuntimeInitError\n%s", string(msg))
		case http.StatusInternalServerError:
			return resp, fmt.Errorf("container Error status code (500) defaultAPI.postRuntimeInitError\n%s", string(msg))
		default:
			return resp, fmt.Errorf("invalid status code (%d) defaultAPI.postRuntimeInitError\n%s", resp.StatusCode, string(msg))
		}
	}
}

func (api defaultAPI) postRuntimeInvocationError(requestId string, err error) (*http.Response, error) {
	log.Println("defaultAPI.postRuntimeInvocationError", requestId, err)

	header := defaultInvokeErrorHeader
	payload := struct {
		Message    string   `json:"errorMessage"`
		Type       string   `json:"errorType"`
		StackTrace []string `json:"stackTrace"`
	}{
		Message:    err.Error(),
		Type:       defaultInvokeErrorHeader,
		StackTrace: []string{},
	}

	if err, ok := err.(Error); ok {
		header = err.Header()
		payload.Type = err.Type()
	}

	body, _ := json.Marshal(payload)

	request, _ := http.NewRequest(
		http.MethodPost,
		api.invocationUrlPrefix+requestId+"/error",
		bytes.NewBuffer(body),
	)
	request.Header.Add(headerErrorType, header)

	resp, err := api.client.Do(request)
	if err != nil {
		return resp, fmt.Errorf("%w; error submitting defaultAPI.postRuntimeInvocationError request", err)
	} else {
		msg, _ := io.ReadAll(resp.Body)

		switch resp.StatusCode {
		case http.StatusAccepted:
			return resp, fmt.Errorf("accepted status code (202) defaultAPI.postRuntimeInvocationError for request: %s\n%s", requestId, string(msg))
		case http.StatusBadRequest:
			return resp, fmt.Errorf("bad Request status code (400) defaultAPI.postRuntimeInvocationError for request: %s\n%s", requestId, string(msg))
		case http.StatusForbidden:
			return resp, fmt.Errorf("forbidden status code (403) defaultAPI.postRuntimeInvocationError for request: %s\n%s", requestId, string(msg))
		case http.StatusInternalServerError:
			return resp, fmt.Errorf("container Error status code (500) defaultAPI.postRuntimeInvocationError for request: %s\n%s", requestId, string(msg))
		default:
			return resp, fmt.Errorf("invalid status code (%d) defaultAPI.postRuntimeInvocationError for request: %s\n%s", resp.StatusCode, requestId, string(msg))
		}
	}
}

func (api defaultAPI) postRuntimeInvocationResponse(requestId string, response io.Reader) (*http.Response, error) {
	req, _ := http.NewRequest(
		http.MethodPost,
		api.invocationUrlPrefix+requestId+"/response",
		response,
	)

	if response, ok := response.(Response); ok {
		req.Header.Add(headerContentType, response.ContentType())
	} else {
		req.Header.Add(headerContentType, defaultContentType)
	}

	resp, err := api.client.Do(req)
	if err != nil {
		err = fmt.Errorf("%w; defaultAPI.postRuntimeInvocationResponse for request: %s", err, requestId)
		return api.postRuntimeInvocationError(requestId, err)
	}

	return resp, nil
}
