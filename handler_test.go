package llb

import (
	"errors"
	"testing"
)

func TestNewResponse(t *testing.T) {
	resp := NewResponse(nil, "content")
	if resp.ContentType() != "content" {
		t.Fatal("NewResponse did not preserve the content type passed in")
	}
}

func TestNewError(t *testing.T) {
	err := NewError(nil, "header", "typ")
	if err.Header() != "header" {
		t.Fatal("NewError did not preserve the header passed in")
	}
	if err.Type() != "typ" {
		t.Fatal("NewError did not preserve the type passed in")
	}
}

func TestDefaultErrorHandler(t *testing.T) {
	_, err := DefaultErrorHandler(errors.New("test"))
	if err.Error() != "test" {
		t.Fatal("DefaultErrorHandler did not preserve error message")
	}
}
