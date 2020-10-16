package httpclient_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"bitbucket.org/qubole/gateway/internal/httpclient"
	"github.com/hashicorp/go-retryablehttp"
)

func TestPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.URL.String() {
		case "/pass":
			body, _ := ioutil.ReadAll(req.Body)
			rw.Header().Set("TestHeader", "TestHeaderValue")
			rw.Write(body)
			return
		case "/fail":
			body, _ := ioutil.ReadAll(req.Body)
			rw.Header().Set("TestHeader", "TestHeaderValue")
			rw.WriteHeader(http.StatusNotFound)
			rw.Write(body)
			return
		default:
			rw.Header().Set("TestHeader", "TestHeaderValue")
			rw.WriteHeader(http.StatusInternalServerError)
		}
	}))
	// Close the server when test finishes
	defer server.Close()

	testHeaders := http.Header{}
	testHeaders.Set("TestHeader", "TestHeaderValue")

	type args struct {
		ctx     context.Context
		url     string
		query   map[string]interface{}
		headers []map[string]string
		body    interface{}
	}
	tests := []struct {
		name       string
		args       args
		want       *httpclient.Response
		wantStatus int
		wantErr    bool
	}{
		{
			name: "Success",
			args: args{
				url:  "/pass",
				body: map[string]string{"msg": "hello"},
			},
			want: &httpclient.Response{
				Body:    []byte(`{"msg":"hello"}`),
				Headers: testHeaders,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "Failure_404",
			args: args{
				url:  "/fail",
				body: map[string]string{"msg": "Failure_404"},
			},
			want: &httpclient.Response{
				Body:    []byte(`{"msg":"Failure_404"}`),
				Headers: testHeaders,
			},
			wantStatus: http.StatusNotFound,
			wantErr:    false,
		},
		{
			name: "Failure_500",
			args: args{
				url: "/fail_500",
			},
			want: &httpclient.Response{
				Headers: testHeaders,
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			url := server.URL + tt.args.url // eg. 127.0.0.1:9001 + /pass

			got, status, err := httpclient.Post(ctx, url, tt.args.body, tt.args.headers...)
			if err != nil {
				t.Errorf("Post() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantStatus != status {
				t.Errorf("Status got = %v, want %v", status, tt.wantStatus)
			}

			if string(got.Body) != string(tt.want.Body) {
				t.Errorf("Body got = %v, want %v", string(got.Body), string(tt.want.Body))
			}

			// Custom test header should be present
			if !reflect.DeepEqual(got.Headers.Get("TestHeader"), tt.want.Headers.Get("TestHeader")) {
				t.Errorf("Headers got = %v, want %v", got.Headers.Get("TestHeader"), tt.want.Headers.Get("TestHeader"))
			}
		})
	}
}

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		switch req.URL.String() {
		case "/pass":
			rw.Header().Set("TestHeader", "TestHeaderValue")
			rw.Write([]byte(`OK`))
			return
		case "/fail":
			rw.Header().Set("TestHeader", "TestHeaderValue")
			rw.WriteHeader(http.StatusNotFound)
			return
		default:
			rw.Header().Set("TestHeader", "TestHeaderValue")
			rw.WriteHeader(http.StatusInternalServerError)
		}
	}))
	// Close the server when test finishes
	defer server.Close()

	testHeaders := http.Header{}
	testHeaders.Set("TestHeader", "TestHeaderValue")

	type args struct {
		ctx     context.Context
		url     string
		query   map[string]interface{}
		headers []map[string]string
	}
	tests := []struct {
		name       string
		args       args
		want       *httpclient.Response
		wantStatus int
		wantErr    bool
	}{
		{
			name: "Success",
			args: args{
				url: "/pass",
			},
			want: &httpclient.Response{
				Body:    []byte(`OK`),
				Headers: testHeaders,
			},
			wantStatus: http.StatusOK,
			wantErr:    false,
		},
		{
			name: "Failure_404",
			args: args{
				url: "/fail",
			},
			want: &httpclient.Response{
				Headers: testHeaders,
			},
			wantStatus: http.StatusNotFound,
			wantErr:    false,
		},
		{
			name: "Failure_500",
			args: args{
				url: "/fail_500",
			},
			want: &httpclient.Response{
				Headers: testHeaders,
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			url := server.URL + tt.args.url // eg. 127.0.0.1:9001 + /pass

			got, status, err := httpclient.Get(ctx, url, tt.args.query, tt.args.headers...)
			if err != nil {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantStatus != status {
				t.Errorf("Status got = %v, want %v", status, tt.wantStatus)
			}

			if string(got.Body) != string(tt.want.Body) {
				t.Errorf("Body got = %v, want %v", string(got.Body), string(tt.want.Body))
			}

			// Custom test header should be present
			if !reflect.DeepEqual(got.Headers.Get("TestHeader"), tt.want.Headers.Get("TestHeader")) {
				t.Errorf("Headers got = %v, want %v", got.Headers.Get("TestHeader"), tt.want.Headers.Get("TestHeader"))
			}
		})
	}
}

func TestRetryOptions_client(t *testing.T) {
	type fields struct {
		Max        int
		WaitMin    float64
		WaitMax    float64
		CheckRetry retryablehttp.CheckRetry
	}
	tests := []struct {
		name   string
		fields fields
		want   *http.Client
	}{
		{
			name: "TestRetryOptionClient",
			fields: fields{
				Max:     5,
				WaitMin: 0.5,
				WaitMax: 1,
				CheckRetry: func(context context.Context, resp *http.Response, err error) (bool, error) {
					return retryablehttp.DefaultRetryPolicy(context, resp, err)
				},
			},
			want: retryablehttp.NewClient().StandardClient(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ro := &httpclient.RetryOptions{
				Max:        tt.fields.Max,
				WaitMin:    tt.fields.WaitMin,
				WaitMax:    tt.fields.WaitMax,
				CheckRetry: tt.fields.CheckRetry,
			}
			got := ro.Client()
			if fmt.Sprintf("%T", got) != fmt.Sprintf("%T", tt.want) {
				t.Errorf("RetryOptions.client() = %v, want %v", fmt.Sprintf("%T", got), fmt.Sprintf("%T", tt.want))
			}

			if fmt.Sprintf("%T", got.Transport) == fmt.Sprintf("%T", http.Client{}.Transport) {
				t.Errorf("RetryOptions.client().Transport should have a custom transport")
			}
		})
	}
}

func TestClient(t *testing.T) {
	type args struct {
		req *http.Request
		ro  *httpclient.RetryOptions
	}
	tests := []struct {
		name string
		args args
		want *http.Client
	}{
		{
			name: "TestClientNormal",
			args: args{
				ro: nil,
			},
			want: &http.Client{},
		},

		{
			name: "TestClientWithRetry",
			args: args{
				ro: &httpclient.RetryOptions{
					Max:     5,
					WaitMin: 0.5,
					WaitMax: 1,
					CheckRetry: func(context context.Context, resp *http.Response, err error) (bool, error) {
						return retryablehttp.DefaultRetryPolicy(context, resp, err)
					},
				},
			},
			want: retryablehttp.NewClient().StandardClient(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := httpclient.Client(tt.args.req, tt.args.ro)

			if fmt.Sprintf("%T", got.Transport) != fmt.Sprintf("%T", tt.want.Transport) {
				t.Errorf("Client() = %v, want %v", fmt.Sprintf("%T", got.Transport), fmt.Sprintf("%T", tt.want.Transport))
			}

		})
	}
}

func TestGetWithRetries(t *testing.T) {
	type args struct {
		ctx        context.Context
		ro         *httpclient.RetryOptions
		query      map[string]interface{}
		headers    []map[string]string
		successTry int
	}
	tests := []struct {
		name         string
		args         args
		want         *http.Response
		retryCounter int
		got1         int
	}{
		{
			name: "TestGetWithRetriesShouldRetryonFailure",
			got1: 200,
			args: args{
				ro: &httpclient.RetryOptions{
					Max:     3,
					WaitMax: 0.2,
					WaitMin: 0.1,
				},
				ctx:        context.Background(),
				successTry: 2,
			},
		},

		{
			name: "TestGetWithRetriesMaxRetries",
			got1: 200,
			args: args{
				ro: &httpclient.RetryOptions{
					Max:     3,
					WaitMax: 0.2,
					WaitMin: 0.1,
				},
				ctx:        context.Background(),
				successTry: 4, // retries + original
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				tt.retryCounter = tt.retryCounter + 1
				if tt.retryCounter >= tt.args.successTry {
					w.WriteHeader(200)
				} else {
					w.WriteHeader(502)
				}

			}))
			resp, got1, err := httpclient.GetWithRetries(tt.args.ctx, tt.args.ro, testServer.URL, tt.args.query, tt.args.headers...)

			if resp == nil {
				t.Errorf("got empty response")
			}

			if err != nil {
				t.Errorf("Error should be nil:%v", err)
			}

			if got1 != tt.got1 {
				t.Errorf("Response status should have been %v, got:%v", tt.got1, got1)
			}

			if tt.retryCounter != tt.args.successTry {
				t.Errorf("should have tried %v before sucess", tt.retryCounter)
			}

		})
	}

}

func TestGetWithRetriesWithoutRetries(t *testing.T) {
	type args struct {
		ctx        context.Context
		ro         *httpclient.RetryOptions
		query      map[string]interface{}
		headers    []map[string]string
		successTry int
	}
	tests := []struct {
		name         string
		args         args
		want         *http.Response
		retryCounter int
		got1         int
	}{
		{
			name: "NoRetriesWhenRetryOptionisNotProvided",
			got1: 502,
			args: args{
				ro:         nil,
				ctx:        context.Background(),
				successTry: 2,
			},
		},
		{
			name: "NoRetriesWhenCheckRetryDoesntAllow",
			got1: 502,
			args: args{
				ro: &httpclient.RetryOptions{
					Max:     3,
					WaitMax: 0.2,
					WaitMin: 0.1,
					CheckRetry: func(context context.Context, resp *http.Response, err error) (bool, error) {
						return false, nil
					},
				},
				ctx:        context.Background(),
				successTry: 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				tt.retryCounter = tt.retryCounter + 1
				w.WriteHeader(502)

			}))

			resp, got1, err := httpclient.GetWithRetries(tt.args.ctx, tt.args.ro, testServer.URL, tt.args.query, tt.args.headers...)

			if resp == nil {
				t.Errorf("got empty response")
			}

			if err != nil {
				t.Errorf("didnt get error")
			}

			if got1 != tt.got1 {
				t.Errorf("Response status should have been %v, got:%v", tt.got1, got1)
			}

			if tt.retryCounter != 1 {
				t.Errorf("should not have tried %v before sucess", tt.retryCounter)
			}

		})
	}

}
