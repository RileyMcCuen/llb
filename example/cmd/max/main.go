package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func body(size int) io.Reader {
	prefix, suffix := `{"data":"`, `"}`
	data := make([]byte, size-len(suffix), size)
	for i := range data {
		if i < len(prefix) {
			data[i] = prefix[i]
		} else {
			data[i] = byte(i%26) + 'a'
		}
	}

	data = append(data, suffix...)
	return bytes.NewBuffer(data)
}

const (
	max           = 6291456
	maxAddtnlData = 2380
)

func main() {
	log.Println(time.Now())
	for i := 0; i < 1000; i++ {
		reqData := body(max - maxAddtnlData)
		req, _ := http.NewRequest(http.MethodPost, "https://iw88b0fidh.execute-api.us-east-1.amazonaws.com/dev/api/llb", reqData)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		if resp.StatusCode != 200 {
			fmt.Println(resp.Status)
		}
		data, _ := io.ReadAll(resp.Body)
		fmt.Println(string(data))
	}
	log.Println(time.Now())
}
