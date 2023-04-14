package llb

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
)

type (
	defaultReponse struct {
		io.Reader
		contentType string
	}
	Response interface {
		io.Reader
		ContentType() string
	}

	defaultError struct {
		error
		header, typ string
	}
	Error interface {
		error
		Header() string
		Type() string
	}

	/*
		If a llb.Response is returned from handler, the extra response information will be passed on to the response endpoint, to create a conforming Response use NewResponse

		If a llb.Error is returned from handler, the extra error information will be passed on to the error endpoint, to create a conforming Error use NewError
	*/
	Handler func(ctx context.Context, r io.Reader) (io.Reader, error)

	TypedHandler[In, Out any] func(ctx context.Context, in In) (Out, error)
	ErrorHandler              func(err error) (io.Reader, error)
)

var (
	_ = Response(defaultReponse{})
	_ = Error(defaultError{})
	_ = ErrorHandler(DefaultErrorHandler)
	_ = ErrorHandler(JsonErrorHandler)
)

func (cr defaultReponse) ContentType() string { return cr.contentType }

func NewResponse(r io.Reader, contentType string) Response {
	return defaultReponse{
		Reader:      r,
		contentType: contentType,
	}
}

func (ce defaultError) Header() string { return ce.header }
func (ce defaultError) Type() string   { return ce.typ }

func NewError(err error, header, typ string) Error {
	return defaultError{
		error:  err,
		header: header,
		typ:    typ,
	}
}

// WrapHandler creates a Handler from a TypedHandler and an optional ErrorHandler, if no ErrorHandler is provided DefaultErrHandler is used instead
func WrapHandler[In, Out any](handler TypedHandler[In, Out], errHandler ErrorHandler) Handler {
	if errHandler == nil {
		errHandler = DefaultErrorHandler
	}

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

		data, err = json.Marshal(out)
		if err != nil {
			return errHandler(err)
		}

		return bytes.NewBuffer(data), nil
	}
}

func DefaultErrorHandler(err error) (io.Reader, error) { return nil, err }

func JsonErrorHandler(err error) (io.Reader, error) {
	data, _ := json.Marshal(struct {
		Err string `json:"error"`
	}{
		Err: err.Error(),
	})

	return bytes.NewBuffer(data), nil
}
