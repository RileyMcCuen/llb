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
	lambdaRuntime struct {
		api     lambdaRuntimeAPI
		handler Handler
		meta    RequestMeta
	}
)

const (
	envRuntimeDomain = "AWS_LAMBDA_RUNTIME_API"
	envTraceId       = "_X_AMZN_TRACE_ID"

	headerRequestId       = "Lambda-Runtime-Aws-Request-Id"
	headerDeadline        = "Lambda-Runtime-Deadline-Ms"
	headerLambdaArn       = "Lambda-Runtime-Invoked-Function-Arn"
	headerTraceId         = "Lambda-Runtime-Trace-Id"
	headerClientContext   = "Lambda-Runtime-Client-Context"
	headerCognitoIdentity = "Lambda-Runtime-Cognito-Identity"
)

func Start(handler Handler) {
	lrt := newLambdaRuntime(handler)

	defer lrt.recover()

	for {
		if err := lrt.next(); err != nil {
			log.Fatal(err)
		}

		lrt.reset()
	}
}

func newLambdaRuntime(handler Handler) *lambdaRuntime {
	domain := os.Getenv(envRuntimeDomain)

	return &lambdaRuntime{
		api: defaultAPI{
			domain:              domain,
			invocationUrlPrefix: "http://" + domain + "/2018-06-01/runtime/invocation/",
			nextUrl:             "http://" + domain + "/2018-06-01/runtime/invocation/next",
			initErrorUrl:        "http://" + domain + "/2018-06-01/runtime/init/error",
		},
		handler: handler,
	}
}

func (lrt *lambdaRuntime) recover() {
	if err := recover(); err != nil {
		if err, ok := err.(error); ok {
			if lrt.meta.RequestId == "" {
				_, err := lrt.api.postRuntimeInitError(err)
				log.Fatal(err)
			} else {
				_, err := lrt.api.postRuntimeInvocationError(lrt.meta.RequestId, err)
				log.Fatal(err)
			}
		}
	}
}

func (lrt *lambdaRuntime) next() error {
	resp, err := lrt.api.getRuntimeInvocationNext()

	if err != nil {
		_, err = lrt.api.postRuntimeInitError(err)
		return err
	}

	if err := lrt.updateMeta(resp); err != nil {
		_, err = lrt.api.postRuntimeInitError(err)
		return err
	}

	ctx := context.WithValue(context.Background(), contextKey, lrt.meta)

	handlerResponse, err := lrt.handler(ctx, resp.Body)

	if err != nil {
		_, err = lrt.api.postRuntimeInvocationError(lrt.meta.RequestId, err)
		return err
	} else {
		resp, err := lrt.api.postRuntimeInvocationResponse(lrt.meta.RequestId, handlerResponse)
		if err != nil {
			return err
		}

		resp.Body.Close()
		return nil
	}
}

func (lrt *lambdaRuntime) reset() {
	lrt.meta = RequestMeta{}
}

func (lrt *lambdaRuntime) updateMeta(resp *http.Response) error {
	var err error
	headers := resp.Header

	lrt.meta.TraceId, err = validateTraceId(headers)
	if err != nil {
		return fmt.Errorf("%w; lambdaRuntime.updateMeta", err)
	}

	lrt.meta.RequestId, err = validateHeader(headers, headerRequestId)
	if err != nil {
		return fmt.Errorf("%w; lambdaRuntime.updateMeta", err)
	}

	lrt.meta.Deadline, err = validateDeadline(headers)
	if err != nil {
		return fmt.Errorf("%w; lambdaRuntime.updateMeta", err)
	}

	lrt.meta.LambdaArn, err = validateHeader(headers, headerLambdaArn)
	if err != nil {
		return fmt.Errorf("%w; lambdaRuntime.updateMeta", err)
	}

	lrt.meta.ClientContext = validateHeaderNoError(headers, headerClientContext)

	lrt.meta.CognitoIdentity = validateHeaderNoError(headers, headerCognitoIdentity)
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
