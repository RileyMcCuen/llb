package handlerutil

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"llb"

	"github.com/aws/aws-lambda-go/events"
)

type (
	InTypedHandler[In any]         func(ctx context.Context, in In) error
	InOutTypedHandler[In, Out any] func(ctx context.Context, in In) (Out, error)
	nothing                        struct{}
)

var (
	Nothing = nothing{}
)

// InTypeHandler creates a Handler from an InTypedHandler and an optional errHandler, if no errHandler is provided DefaultErrHandler is used instead
func InTypeHandler[In any](handler func(ctx context.Context, in In) error, errHandler llb.ErrorHandler) llb.Handler {
	return InOutTypeHandler(func(ctx context.Context, in In) (nothing, error) {
		return Nothing, handler(ctx, in)
	}, errHandler)
}

// InOutTypeHandler creates a Handler from an InOutTypedHandler and an optional errHandler, if no errHandler is provided DefaultErrHandler is used instead
func InOutTypeHandler[In any, Out any](handler InOutTypedHandler[In, Out], errHandler llb.ErrorHandler) llb.Handler {
	if errHandler == nil {
		errHandler = llb.DefaultErrorHandler
	}

	_, nilOut := any(*new(Out)).(nothing)

	return func(ctx context.Context, r io.Reader) (io.Reader, error) {
		in := new(In)
		data, _ := io.ReadAll(r)

		if err := json.Unmarshal(data, in); err != nil {
			return errHandler(err)
		}

		out, err := handler(ctx, *in)
		if err != nil {
			return errHandler(err)
		}

		if nilOut {
			return nil, nil
		}

		data, err = json.Marshal(out)
		if err != nil {
			return errHandler(err)
		}

		return bytes.NewBuffer(data), nil
	}
}

func SNSHandler(handler func(ctx context.Context, in events.SNSEvent) error, errHandler llb.ErrorHandler) llb.Handler {
	return InOutTypeHandler(func(ctx context.Context, in events.SNSEvent) (nothing, error) {
		return Nothing, handler(ctx, in)
	}, errHandler)
}

func SQSHandler(handler func(ctx context.Context, in events.SQSEvent) error, errHandler llb.ErrorHandler) llb.Handler {
	return InOutTypeHandler(func(ctx context.Context, in events.SQSEvent) (nothing, error) {
		return Nothing, handler(ctx, in)
	}, errHandler)
}

func CWEHandler(handler func(ctx context.Context, in events.CloudWatchEvent) error, errHandler llb.ErrorHandler) llb.Handler {
	return InOutTypeHandler(func(ctx context.Context, in events.CloudWatchEvent) (nothing, error) {
		return Nothing, handler(ctx, in)
	}, errHandler)
}

func APIGatewayHandler(handler func(ctx context.Context, in events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error), errHandler llb.ErrorHandler) llb.Handler {
	return InOutTypeHandler(handler, errHandler)
}
