package llb

import (
	"context"
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

	ErrorHandler func(err error) (io.Reader, error)
)

var (
	_ = Response(defaultReponse{})
	_ = Error(defaultError{})
	_ = ErrorHandler(DefaultErrorHandler)
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

func DefaultErrorHandler(err error) (io.Reader, error) { return nil, err }
