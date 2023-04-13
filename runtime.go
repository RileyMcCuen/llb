package llb

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type (
	runtime struct {
		api     api
		handler Handler
		meta    RequestMeta
		fatal   func(error)
	}
)

const (
	envTraceId = "_X_AMZN_TRACE_ID"

	headerRequestId       = "Lambda-Runtime-Aws-Request-Id"
	headerDeadline        = "Lambda-Runtime-Deadline-Ms"
	headerLambdaArn       = "Lambda-Runtime-Invoked-Function-Arn"
	headerTraceId         = "Lambda-Runtime-Trace-Id"
	headerClientContext   = "Lambda-Runtime-Client-Context"
	headerCognitoIdentity = "Lambda-Runtime-Cognito-Identity"
)

func Start(handler Handler) {
	newruntime(handler, newDefaultAPI()).start()
}

func newruntime(handler Handler, api api) *runtime {
	return &runtime{
		api:     api,
		handler: handler,
		meta:    RequestMeta{},
		fatal:   func(err error) { log.Fatal(err.Error()) },
	}
}

func (rt *runtime) start() {
	defer rt.recover()

	for {
		if err := rt.next(); err != nil {
			rt.fatal(err)
		}

		rt.reset()
	}
}

func (rt *runtime) recover() {
	if err := recover(); err != nil {
		if err, ok := err.(error); ok {
			if rt.meta.RequestId == "" {
				_, err := rt.api.postRuntimeInitError(err)
				rt.fatal(err)
			} else {
				_, err := rt.api.postRuntimeInvocationError(rt.meta.RequestId, err)
				rt.fatal(err)
			}
		}
	}
}

func (rt *runtime) next() error {
	resp, err := rt.api.getRuntimeInvocationNext()

	if err != nil {
		_, err = rt.api.postRuntimeInitError(err)
		return err
	}

	if err := rt.updateMeta(resp); err != nil {
		_, err = rt.api.postRuntimeInitError(err)
		return err
	}

	ctx := context.WithValue(context.Background(), contextKey, rt.meta)

	handlerResponse, err := rt.handler(ctx, resp.Body)

	if err != nil {
		_, err = rt.api.postRuntimeInvocationError(rt.meta.RequestId, err)
		return err
	} else {
		resp, err := rt.api.postRuntimeInvocationResponse(rt.meta.RequestId, handlerResponse)
		if err != nil {
			return err
		}

		resp.Body.Close()
		return nil
	}
}

func (rt *runtime) reset() {
	rt.meta = RequestMeta{}
}

func (rt *runtime) updateMeta(resp *http.Response) error {
	var err error
	headers := resp.Header

	rt.meta.TraceId, err = validateTraceId(headers)
	if err != nil {
		return fmt.Errorf("%w; lambdaRuntime.updateMeta", err)
	}

	rt.meta.RequestId, err = validateHeader(headers, headerRequestId)
	if err != nil {
		return fmt.Errorf("%w; lambdaRuntime.updateMeta", err)
	}

	rt.meta.Deadline, err = validateDeadline(headers)
	if err != nil {
		return fmt.Errorf("%w; lambdaRuntime.updateMeta", err)
	}

	rt.meta.LambdaArn, err = validateHeader(headers, headerLambdaArn)
	if err != nil {
		return fmt.Errorf("%w; lambdaRuntime.updateMeta", err)
	}

	rt.meta.ClientContext = validateHeaderNoError(headers, headerClientContext)

	rt.meta.CognitoIdentity = validateHeaderNoError(headers, headerCognitoIdentity)
	return nil
}

func validateHeaderNoError(headers http.Header, key string) string {
	vals := headers[key]

	if len(vals) < 1 {
		return ""
	}

	return vals[0]
}

func validateHeader(headers http.Header, key string) (string, error) {
	vals := headers[key]

	if len(vals) < 1 {
		return "", fmt.Errorf("validateHeader failed for key %s: no header values found for that key", key)
	}

	return vals[0], nil
}

func validateTraceId(headers http.Header) (string, error) {
	traceId, err := validateHeader(headers, headerTraceId)
	if err != nil {
		return "", err
	}
	os.Setenv(envTraceId, traceId)

	return traceId, nil
}

func validateDeadline(headers http.Header) (time.Time, error) {
	deadlineMs, err := validateHeader(headers, headerDeadline)
	if err != nil {
		return time.Time{}, err
	}

	deadlineMsInt, err := strconv.ParseInt(deadlineMs, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w; deadline header is not a valid unix ms string", err)
	}

	return time.UnixMilli(deadlineMsInt), nil
}
