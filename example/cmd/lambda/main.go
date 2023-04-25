package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/RileyMcCuen/llb/pkg/handlerutil"

	"github.com/RileyMcCuen/llb"

	"github.com/aws/aws-lambda-go/events"
)

func main() {
	// llb.Start(Handler)
	llb.Start(handlerutil.APIGatewayHandler(Handler, handlerutil.JsonErrorHandler))
}

var (
	// _ = llb.Handler(Handler)
	_ = handlerutil.InOutTypedHandler[events.APIGatewayProxyRequest, events.APIGatewayProxyResponse](Handler)
)

// func Handler(ctx context.Context, r io.Reader) (io.Reader, error) {
// 	data, err := io.ReadAll(r)

// 	log.Println(os.Getenv("test"), err, len(data))
// 	req := &events.APIGatewayProxyRequest{}
// 	if err := json.Unmarshal(data, r); err != nil {
// 		log.Println(err)
// 	}
// 	log.Println("Encoded", req.IsBase64Encoded)
// 	out, _ := json.Marshal(events.APIGatewayProxyResponse{
// 		StatusCode: 200,
// 		Body:       fmt.Sprintf(`{"size":%d}`, len(data)),
// 	})

// 	return bytes.NewBuffer(out), nil
// }

func Handler(ctx context.Context, r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println(os.Getenv("test"), len(r.Body))

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf(`{"size":%d}`, len(r.Body)),
	}, nil
}
