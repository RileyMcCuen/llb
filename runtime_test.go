package llb

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func newValidNextResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     strconv.FormatInt(http.StatusOK, 10),
		Header:     http.Header{headerTraceId: []string{"trace"}, headerRequestId: []string{"req"}, headerDeadline: []string{"100"}, headerLambdaArn: []string{"arn"}},
		Body:       io.NopCloser(bytes.NewBufferString(`{"datakey":"dataval"}`)),
	}
}

type (
	errorReadCloser struct{}
	mockAPI         struct {
		_getRuntimeInvocationNext      func() (resp *http.Response, err error)
		_postRuntimeInitError          func(err error) (*http.Response, error)
		_postRuntimeInvocationError    func(requestId string, err error) (*http.Response, error)
		_postRuntimeInvocationResponse func(requestId string, response io.Reader) (*http.Response, error)
	}
)

var (
	_ = io.ReadCloser(errorReadCloser{})
	_ = api(mockAPI{})
)

func (errorReadCloser) Read([]byte) (int, error) { return 0, nil }
func (errorReadCloser) Close() error             { return errors.New("error") }

func (api mockAPI) getRuntimeInvocationNext() (*http.Response, error) {
	return api._getRuntimeInvocationNext()
}

func (api mockAPI) postRuntimeInitError(err error) (*http.Response, error) {
	return api._postRuntimeInitError(err)
}

func (api mockAPI) postRuntimeInvocationError(requestId string, err error) (*http.Response, error) {
	return api._postRuntimeInvocationError(requestId, err)
}

func (api mockAPI) postRuntimeInvocationResponse(requestId string, response io.Reader) (*http.Response, error) {
	return api._postRuntimeInvocationResponse(requestId, response)
}

func Test_newRuntime(t *testing.T) {
	type args struct {
		handler Handler
		api     api
		fatal   func(error)
	}
	tests := []struct {
		name string
		args args
		want *runtime
	}{
		{"Success", args{handler: nil, api: nil, fatal: nil}, &runtime{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newRuntime(tt.args.handler, tt.args.api, tt.args.fatal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newRuntime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_runtime_start(t *testing.T) {
	type fields struct {
		api     api
		handler Handler
		meta    RequestMeta
		fatal   func(error)
	}
	firstRunForSuccessTest := true
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "Success - then panic to end test",
			fields: fields{
				api: mockAPI{
					_getRuntimeInvocationNext: func() (resp *http.Response, err error) {
						if firstRunForSuccessTest {
							firstRunForSuccessTest = false
							return newValidNextResponse(), nil
						} else {
							return nil, errors.New("error")
						}
					},
					_postRuntimeInitError: func(err error) (*http.Response, error) {
						return nil, err
					},
					_postRuntimeInvocationError: func(requestId string, err error) (*http.Response, error) {
						return nil, err
					},
					_postRuntimeInvocationResponse: func(requestId string, response io.Reader) (*http.Response, error) {
						return &http.Response{Body: io.NopCloser(bytes.NewBufferString(""))}, nil
					},
				},
				handler: func(ctx context.Context, r io.Reader) (io.Reader, error) { return bytes.NewBufferString(""), nil },
				fatal:   func(err error) { panic(errors.New("fatal")) },
			},
		},
		{
			name: "Fail With Request Id",
			fields: fields{
				api: mockAPI{
					_getRuntimeInvocationNext: func() (resp *http.Response, err error) {
						return newValidNextResponse(), nil
					},
					_postRuntimeInitError: func(err error) (*http.Response, error) {
						return nil, err
					},
					_postRuntimeInvocationError: func(requestId string, err error) (*http.Response, error) {
						return nil, err
					},
					_postRuntimeInvocationResponse: func(requestId string, response io.Reader) (*http.Response, error) {
						return &http.Response{Body: io.NopCloser(bytes.NewBufferString(""))}, nil
					},
				},
				handler: func(ctx context.Context, r io.Reader) (io.Reader, error) { return nil, errors.New("error") },
				fatal:   func(err error) { panic(errors.New("fatal")) },
			},
		},
		{
			name: "Fail Without Request Id",
			fields: fields{
				api: mockAPI{
					_getRuntimeInvocationNext: func() (resp *http.Response, err error) {
						return nil, errors.New("error")
					},
					_postRuntimeInitError: func(err error) (*http.Response, error) {
						return nil, err
					},
					_postRuntimeInvocationError: func(requestId string, err error) (*http.Response, error) {
						return nil, err
					},
					_postRuntimeInvocationResponse: func(requestId string, response io.Reader) (*http.Response, error) {
						return &http.Response{Body: io.NopCloser(bytes.NewBufferString(""))}, nil
					},
				},
				handler: func(ctx context.Context, r io.Reader) (io.Reader, error) { return nil, errors.New("error") },
				fatal:   func(err error) { panic(errors.New("fatal")) },
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := &runtime{
				api:     tt.fields.api,
				handler: tt.fields.handler,
				meta:    tt.fields.meta,
				fatal:   tt.fields.fatal,
			}
			rt.start()
		})
	}
}

func Test_runtime_next(t *testing.T) {
	type fields struct {
		api     api
		handler Handler
		meta    RequestMeta
		fatal   func(error)
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Success",
			fields: fields{
				api: mockAPI{
					_getRuntimeInvocationNext: func() (resp *http.Response, err error) {
						return newValidNextResponse(), nil
					},
					_postRuntimeInitError: func(err error) (*http.Response, error) {
						return nil, err
					},
					_postRuntimeInvocationError: func(requestId string, err error) (*http.Response, error) {
						return nil, err
					},
					_postRuntimeInvocationResponse: func(requestId string, response io.Reader) (*http.Response, error) {
						return &http.Response{Body: io.NopCloser(bytes.NewBufferString(""))}, nil
					},
				},
				handler: func(ctx context.Context, r io.Reader) (io.Reader, error) { return nil, nil },
			},
			wantErr: false,
		},
		{
			name: "Response Error",
			fields: fields{
				api: mockAPI{
					_getRuntimeInvocationNext: func() (resp *http.Response, err error) {
						return newValidNextResponse(), nil
					},
					_postRuntimeInitError: func(err error) (*http.Response, error) {
						return nil, err
					},
					_postRuntimeInvocationError: func(requestId string, err error) (*http.Response, error) {
						return nil, err
					},
					_postRuntimeInvocationResponse: func(requestId string, response io.Reader) (*http.Response, error) {
						return &http.Response{Body: io.NopCloser(bytes.NewBufferString(""))}, errors.New("error")
					},
				},
				handler: func(ctx context.Context, r io.Reader) (io.Reader, error) { return nil, nil },
			},
			wantErr: true,
		},
		{
			name: "Next Error",
			fields: fields{
				api: mockAPI{
					_getRuntimeInvocationNext: func() (resp *http.Response, err error) {
						return nil, errors.New("error")
					},
					_postRuntimeInitError: func(err error) (*http.Response, error) {
						return nil, err
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Update Meta",
			fields: fields{
				api: mockAPI{
					_getRuntimeInvocationNext: func() (resp *http.Response, err error) {
						return &http.Response{Header: http.Header{}}, nil
					},
					_postRuntimeInitError: func(err error) (*http.Response, error) {
						return nil, err
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Handler Error",
			fields: fields{
				api: mockAPI{
					_getRuntimeInvocationNext: func() (resp *http.Response, err error) {
						return newValidNextResponse(), nil
					},
					_postRuntimeInitError: func(err error) (*http.Response, error) {
						return nil, err
					},
					_postRuntimeInvocationError: func(requestId string, err error) (*http.Response, error) {
						return nil, err
					},
				},
				handler: func(ctx context.Context, r io.Reader) (io.Reader, error) { return nil, errors.New("error") },
			},
			wantErr: true,
		},
		{
			name: "Next Close Body Error",
			fields: fields{
				api: mockAPI{
					_getRuntimeInvocationNext: func() (resp *http.Response, err error) {
						resp = newValidNextResponse()
						resp.Body = errorReadCloser{}
						return resp, nil
					},
					_postRuntimeInitError: func(err error) (*http.Response, error) {
						return nil, err
					},
					_postRuntimeInvocationError: func(requestId string, err error) (*http.Response, error) {
						return nil, err
					},
				},
				handler: func(ctx context.Context, r io.Reader) (io.Reader, error) { return nil, errors.New("error") },
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := &runtime{
				api:     tt.fields.api,
				handler: tt.fields.handler,
				meta:    tt.fields.meta,
				fatal:   tt.fields.fatal,
			}
			if err := rt.next(); (err != nil) != tt.wantErr {
				t.Errorf("runtime.next() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_runtime_reset(t *testing.T) {
	type fields struct {
		api     api
		handler Handler
		meta    RequestMeta
		fatal   func(error)
	}
	fs := fields{}

	tests := []struct {
		name   string
		fields fields
		want   RequestMeta
	}{
		{name: "Success", fields: fs},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := &runtime{
				api:     tt.fields.api,
				handler: tt.fields.handler,
				meta:    tt.fields.meta,
				fatal:   tt.fields.fatal,
			}
			rt.reset()
			if !reflect.DeepEqual(rt.meta, tt.want) {
				t.Errorf("rt.meta = %v, want %v", rt.meta, tt.want)
			}
		})
	}
}

func Test_runtime_updateMeta(t *testing.T) {
	type fields struct {
		api     api
		handler Handler
		meta    RequestMeta
		fatal   func(error)
	}
	fs := fields{}

	type args struct {
		resp *http.Response
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "Success", fields: fs, args: args{resp: newValidNextResponse()}, wantErr: false},
		{name: "Missing Trace Id", fields: fs, args: args{resp: &http.Response{Header: http.Header{headerRequestId: []string{"req"}, headerDeadline: []string{"100"}, headerLambdaArn: []string{"arn"}}}}, wantErr: true},
		{name: "Missing Request Id", fields: fs, args: args{resp: &http.Response{Header: http.Header{headerTraceId: []string{"trace"}, headerDeadline: []string{"100"}, headerLambdaArn: []string{"arn"}}}}, wantErr: true},
		{name: "Missing Deadline", fields: fs, args: args{resp: &http.Response{Header: http.Header{headerTraceId: []string{"trace"}, headerRequestId: []string{"req"}, headerLambdaArn: []string{"arn"}}}}, wantErr: true},
		{name: "Missing Lambda ARN", fields: fs, args: args{resp: &http.Response{Header: http.Header{headerTraceId: []string{"trace"}, headerRequestId: []string{"req"}, headerDeadline: []string{"100"}}}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt := &runtime{
				api:     tt.fields.api,
				handler: tt.fields.handler,
				meta:    tt.fields.meta,
				fatal:   tt.fields.fatal,
			}
			if err := rt.updateMeta(tt.args.resp); (err != nil) != tt.wantErr {
				t.Errorf("runtime.updateMeta() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateHeaderNoError(t *testing.T) {
	type args struct {
		headers http.Header
		key     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Success", args: args{key: "h", headers: http.Header{"h": []string{"abc"}}}, want: "abc"},
		{name: "No Header", args: args{key: "h", headers: http.Header{}}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateHeaderNoError(tt.args.headers, tt.args.key); got != tt.want {
				t.Errorf("validateHeaderNoError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateHeader(t *testing.T) {
	type args struct {
		headers http.Header
		key     string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "Success", args: args{key: "h", headers: http.Header{"h": []string{"abc"}}}, want: "abc", wantErr: false},
		{name: "No Header", args: args{key: "h", headers: http.Header{}}, want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateHeader(tt.args.headers, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("validateHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateTraceId(t *testing.T) {
	type args struct {
		headers http.Header
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "Success", args: args{headers: http.Header{headerTraceId: []string{"abc"}}}, want: "abc", wantErr: false},
		{name: "No Header", args: args{headers: http.Header{}}, want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateTraceId(tt.args.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTraceId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("validateTraceId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateDeadline(t *testing.T) {
	type args struct {
		headers http.Header
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{name: "Success", args: args{headers: http.Header{headerDeadline: []string{"100"}}}, want: time.UnixMilli(100), wantErr: false},
		{name: "No Header", args: args{headers: http.Header{}}, want: time.Time{}, wantErr: true},
		{name: "Invalid Header", args: args{headers: http.Header{headerDeadline: []string{"abc"}}}, want: time.Time{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateDeadline(tt.args.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDeadline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validateDeadline() = %v, want %v", got, tt.want)
			}
		})
	}
}
