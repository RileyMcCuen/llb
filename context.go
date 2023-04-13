package llb

import (
	"context"
	"time"
)

type (
	requestMetaContextKey struct{}
	RequestMeta           struct {
		TraceId         string
		RequestId       string
		Deadline        time.Time
		LambdaArn       string
		ClientContext   string
		CognitoIdentity string
	}
)

var (
	contextKey = requestMetaContextKey{}
)

func GetRequestMeta(ctx context.Context) (RequestMeta, bool) {
	raw := ctx.Value(contextKey)
	if raw == nil {
		return RequestMeta{}, false
	}

	return raw.(RequestMeta), true
}

func MustRequestMeta(ctx context.Context) RequestMeta {
	return ctx.Value(contextKey).(RequestMeta)
}
