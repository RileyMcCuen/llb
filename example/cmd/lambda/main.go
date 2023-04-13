package main

import (
	"bytes"
	"context"
	"io"
	"llb"
	"log"
)

func main() {
	// llb.Start(Handler)
	llb.Start(llb.WrapHandler(TypedHandler, llb.JsonErrorHandler))
}

var (
	_ = llb.Handler(Handler)
	_ = llb.TypedHandler[Request, Response](TypedHandler)
)

func Handler(ctx context.Context, r io.Reader) (io.Reader, error) {
	data, err := io.ReadAll(r)

	log.Println(err, string(data))

	return bytes.NewBufferString(`{"status":"success"}`), nil
}

type (
	Request struct {
		Val int `json:"val"`
	}
	Response struct {
		Msg string `json:"message"`
	}
)

func TypedHandler(ctx context.Context, r Request) (Response, error) {
	log.Println("Value", r.Val)

	return Response{
		Msg: "success, printed the value",
	}, nil
}
