package llb

import (
	"errors"
	"io"
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

func TestJsonErrorHandler(t *testing.T) {
	r, err := JsonErrorHandler(errors.New("test"))
	if err != nil {
		t.Fatal("JsonErrorHandler returned an error but should not have")
	}
	data, _ := io.ReadAll(r)
	if string(data) != `{"error":"test"}` {
		t.Fatal("JsonErrorHandler did not preserve error message")
	}
}
