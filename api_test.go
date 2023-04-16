package llb

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"reflect"
	"testing"
)

type (
	mockHttpClient struct {
		do func(r *http.Request) (*http.Response, error)
	}
)

var (
	_ = httpClient(mockHttpClient{})
)

func (client mockHttpClient) Do(r *http.Request) (*http.Response, error) {
	return client.do(r)
}

func valid202Response() *http.Response {
	return &http.Response{Status: "202", StatusCode: http.StatusAccepted, Body: io.NopCloser(bytes.NewBufferString("data"))}
}

func valid400Response() *http.Response {
	return &http.Response{Status: "400", StatusCode: http.StatusBadRequest, Body: io.NopCloser(bytes.NewBufferString("data"))}
}

func valid403Response() *http.Response {
	return &http.Response{Status: "403", StatusCode: http.StatusForbidden, Body: io.NopCloser(bytes.NewBufferString("data"))}
}

func valid500Response() *http.Response {
	return &http.Response{Status: "500", StatusCode: http.StatusInternalServerError, Body: io.NopCloser(bytes.NewBufferString("data"))}
}

func valid507Response() *http.Response {
	return &http.Response{Status: "507", StatusCode: http.StatusInsufficientStorage, Body: io.NopCloser(bytes.NewBufferString("data"))}
}

func Test_newDefaultAPI(t *testing.T) {
	domain := "domain"
	os.Setenv(envRuntimeDomain, domain)

	got := newDefaultAPI(nil)

	want := defaultAPI{
		domain:              domain,
		invocationUrlPrefix: "http://domain/2018-06-01/runtime/invocation/",
		nextUrl:             "http://domain/2018-06-01/runtime/invocation/next",
		initErrorUrl:        "http://domain/2018-06-01/runtime/init/error",
		client:              nil,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Test_newDefaultAPI() = %v, want %v", got, want)
	}
}

func Test_defaultAPI_getRuntimeInvocationNext(t *testing.T) {
	tests := []struct {
		name    string
		api     defaultAPI
		want    *http.Response
		wantErr bool
	}{
		{
			name: "Success",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) { return nil, nil },
			}),
			want:    nil,
			wantErr: false,
		},
		{
			name: "Error",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) { return nil, errors.New("error") },
			}),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.api.getRuntimeInvocationNext()
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultAPI.getRuntimeInvocationNext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("defaultAPI.getRuntimeInvocationNext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_defaultAPI_postRuntimeInitError(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name    string
		api     defaultAPI
		args    args
		want    *http.Response
		wantErr bool
	}{
		{
			name: "Response Error",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) {
					return nil, errors.New("error")
				},
			}),
			args: args{
				err: errors.New("error"),
			},
			wantErr: true,
		},
		{
			name: "Custom Error",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) {
					return nil, errors.New("error")
				},
			}),
			args: args{
				err: NewError(errors.New("error"), "header", "typ"),
			},
			wantErr: true,
		},
		{
			name: "Error 202",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) {
					return valid202Response(), nil
				},
			}),
			args: args{
				err: errors.New("error"),
			},
			wantErr: true,
		},
		{
			name: "Error 403",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) {
					return valid403Response(), nil
				},
			}),
			args: args{
				err: errors.New("error"),
			},
			wantErr: true,
		},
		{
			name: "Error 500",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) {
					return valid500Response(), nil
				},
			}),
			args: args{
				err: errors.New("error"),
			},
			wantErr: true,
		},
		{
			name: "Error 507",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) {
					return valid507Response(), nil
				},
			}),
			args: args{
				err: errors.New("error"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.api.postRuntimeInitError(tt.args.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultAPI.postRuntimeInitError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_defaultAPI_postRuntimeInvocationError(t *testing.T) {
	type args struct {
		requestId string
		err       error
	}
	tests := []struct {
		name    string
		api     defaultAPI
		args    args
		wantErr bool
	}{
		{
			name: "Response Error",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) {
					return nil, errors.New("error")
				},
			}),
			args: args{
				err: errors.New("error"),
			},
			wantErr: true,
		},
		{
			name: "Custom Error",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) {
					return nil, errors.New("error")
				},
			}),
			args: args{
				err: NewError(errors.New("error"), "header", "typ"),
			},
			wantErr: true,
		},
		{
			name: "Error 202",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) {
					return valid202Response(), nil
				},
			}),
			args: args{
				err: errors.New("error"),
			},
			wantErr: true,
		},
		{
			name: "Error 400",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) {
					return valid400Response(), nil
				},
			}),
			args: args{
				err: errors.New("error"),
			},
			wantErr: true,
		},
		{
			name: "Error 403",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) {
					return valid403Response(), nil
				},
			}),
			args: args{
				err: errors.New("error"),
			},
			wantErr: true,
		},
		{
			name: "Error 500",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) {
					return valid500Response(), nil
				},
			}),
			args: args{
				err: errors.New("error"),
			},
			wantErr: true,
		},
		{
			name: "Error 507",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) {
					return valid507Response(), nil
				},
			}),
			args: args{
				err: errors.New("error"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.api.postRuntimeInvocationError(tt.args.requestId, tt.args.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultAPI.postRuntimeInvocationError() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_defaultAPI_postRuntimeInvocationResponse(t *testing.T) {
	type args struct {
		requestId string
		response  io.Reader
	}
	tests := []struct {
		name    string
		api     defaultAPI
		args    args
		wantErr bool
	}{
		{
			name: "Success",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) {
					return nil, nil
				},
			}),
			args: args{
				requestId: "request",
				response:  nil,
			},
			wantErr: false,
		},
		{
			name: "Custom Success",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) {
					return nil, nil
				},
			}),
			args: args{
				requestId: "request",
				response:  NewResponse(nil, "content"),
			},
			wantErr: false,
		},
		{
			name: "Error",
			api: newDefaultAPI(mockHttpClient{
				do: func(r *http.Request) (*http.Response, error) {
					return nil, errors.New("error")
				},
			}),
			args: args{
				requestId: "request",
				response:  nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.api.postRuntimeInvocationResponse(tt.args.requestId, tt.args.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultAPI.postRuntimeInvocationResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
