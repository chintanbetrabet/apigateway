package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"bitbucket.org/qubole/gateway/internal/instrument"
	"bitbucket.org/qubole/gateway/internal/logger"
	"github.com/hashicorp/go-retryablehttp"
)

var glogger = logger.Create(os.Stdout)

// Response ...
type Response struct {
	Body    []byte
	Headers http.Header
}

// RetryOptions ...
type RetryOptions struct {
	Max        int
	WaitMin    float64
	WaitMax    float64
	CheckRetry retryablehttp.CheckRetry
}

//Client ...
func (ro *RetryOptions) Client() *http.Client {

	rc := retryablehttp.NewClient()

	if ro.Max > 0 {
		rc.RetryMax = ro.Max
	}

	if ro.WaitMin > 0.0 {
		rc.RetryWaitMin = time.Duration(ro.WaitMin) * time.Second
	}

	if ro.WaitMax > 0.0 {
		rc.RetryWaitMax = time.Duration(ro.WaitMax) * time.Second
	}

	if ro.CheckRetry != nil {
		rc.CheckRetry = ro.CheckRetry
	}

	// dont use default logger of retryablehttp
	rc.Logger = nil

	// Hook is executed before each retry
	rc.RequestLogHook = func(_ retryablehttp.Logger, req *http.Request, retry int) {

		path := req.URL.Path
		method := req.Method

		if retry > 0 {
			instrument.Increment(strings.ReplaceAll(fmt.Sprintf("infra.gateway.%v.retried", path), "/", "."))
			glogger.Log("retryablehttp", fmt.Sprintf("Retrying %v %v, Attempt: %v", method, path, retry))
		}

		if retry == rc.RetryMax {
			// eg: infra.gateway.tugboat.authenticate.retries.exhausted
			metric := strings.ReplaceAll(fmt.Sprintf("infra.gateway.%v.retries.exhausted", path), "/", ".")
			instrument.Increment(metric)
			glogger.Log("retryablehttp", fmt.Sprintf("Retries Exhausted for %v %v", method, path))
		}
	}

	return rc.StandardClient()
}

// Client ...
func Client(req *http.Request, ro *RetryOptions) *http.Client {
	// if retry option are provided create
	// http client with retry from retryablehttp
	if ro != nil {
		return ro.Client()
	}

	// if no retry option are provided
	// use normal http client
	return &http.Client{}
}

// Post return (statusCode, body, error)
func Post(ctx context.Context, url string, body interface{}, headers ...map[string]string) (*Response, int, error) {
	js, err := json.Marshal(body)
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(js))
	if err != nil {
		return nil, 0, err
	}

	return perform(ctx, nil, req, headers...)
}

// GetWithRetries return (statusCode, body, error)
// This method does a retry on failure
func GetWithRetries(ctx context.Context, ro *RetryOptions, url string, query map[string]interface{}, headers ...map[string]string) (*Response, int, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, 0, err
	}

	if len(query) > 0 {
		q := req.URL.Query()
		for k, v := range query {
			q.Add(k, fmt.Sprintf("%v", v))
		}
		req.URL.RawQuery = q.Encode()
	}

	return perform(ctx, ro, req, headers...)
}

// Get return (statusCode, body, error)
// Get without any retires, retryOption explicitly passed nil
func Get(ctx context.Context, url string, query map[string]interface{}, headers ...map[string]string) (*Response, int, error) {
	return GetWithRetries(ctx, nil, url, query, headers...)
}

func perform(ctx context.Context, ro *RetryOptions, req *http.Request, headers ...map[string]string) (*Response, int, error) {
	req = req.WithContext(ctx)

	// Set converts key to canonicalMIMEkey X-Api-Token
	// so no need to transform key
	if len(headers) > 0 {
		for k, v := range headers[0] {
			req.Header.Set(k, v)
		}
	}

	resp, err := Client(req, ro).Do(req)

	// An error is returned if caused by client policy (such as CheckRedirect),
	// or failure to speak HTTP (such as a network connectivity problem).
	// A non-2xx status code doesn't cause an error.
	// On error, any Response can be ignored. A non-nil Response with a non-nil
	// error only occurs when CheckRedirect fails, and even then the returned
	// Response.Body is already closed.
	if err != nil {
		if resp != nil {
			return &Response{Headers: resp.Header}, resp.StatusCode, err
		}
		return nil, http.StatusInternalServerError, err
	}

	// We will close only when no error as response is always closed in error case
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return &Response{Headers: resp.Header}, http.StatusNoContent, err
	}

	return &Response{Body: b, Headers: resp.Header}, resp.StatusCode, nil
}
