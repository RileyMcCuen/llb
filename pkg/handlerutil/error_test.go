package handlerutil

import (
	"errors"
	"io"
	"testing"
)

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
