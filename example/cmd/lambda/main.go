package main

import (
	"bytes"
	"context"
	"io"
	"llb"
	"log"

	"github.com/aws/aws-lambda-go/events"
)

func main() {
	// llb.Start(Handler)
	llb.Start(llb.WrapHandler(TypedHandler, llb.JsonErrorHandler))
}

var (
	_ = llb.Handler(Handler)
	_ = llb.TypedHandler[events.APIGatewayProxyRequest, events.APIGatewayProxyResponse](TypedHandler)
)

func Handler(ctx context.Context, r io.Reader) (io.Reader, error) {
	data, err := io.ReadAll(r)

	log.Println(err, string(data))

	return bytes.NewBufferString(`{"status":"success"}`), nil
}

type ()

func TypedHandler(ctx context.Context, r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("Value", r.Body)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       `{"status":"success"}`,
	}, nil
}
