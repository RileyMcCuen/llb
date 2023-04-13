package llb

import (
	"context"
	"testing"
)

func TestGetRequestMetaSuccess(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextKey, RequestMeta{TraceId: "trace"})
	meta, ok := GetRequestMeta(ctx)
	if !ok {
		t.Fatal("GetRequestMeta retuned not ok")
	}

	if meta.TraceId != "trace" {
		t.Fatal("GetRequestMeta retuned wrong RequestMeta")
	}
}

func TestGetRequestMetaFailure(t *testing.T) {
	ctx := context.Background()
	_, ok := GetRequestMeta(ctx)
	if ok {
		t.Fatal("GetRequestMeta reported ok when there was no RequestMeta")
	}
}

func TestMustRequestMetaSuccess(t *testing.T) {
	ctx := context.WithValue(context.Background(), contextKey, RequestMeta{TraceId: "trace"})
	meta := MustRequestMeta(ctx)

	if meta.TraceId != "trace" {
		t.Fatal("GetRequestMeta retuned wrong RequestMeta")
	}
}

func TestMustRequestMetaFailure(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Fatal("no panic occurred on bad MustRequestMeta call")
		}
	}()

	ctx := context.Background()
	_ = MustRequestMeta(ctx)
}
