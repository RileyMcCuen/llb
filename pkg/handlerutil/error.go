package handlerutil

import (
	"bytes"
	"encoding/json"
	"io"
	"llb"
)

var (
	_ = llb.ErrorHandler(JsonErrorHandler)
)

func JsonErrorHandler(err error) (io.Reader, error) {
	data, _ := json.Marshal(struct {
		Err string `json:"error"`
	}{
		Err: err.Error(),
	})

	return bytes.NewBuffer(data), nil
}
