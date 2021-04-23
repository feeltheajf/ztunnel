package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
)

const (
	headerAPIToken    = "X-Api-Token"
	headerContentType = "Content-Type"

	contentTypeJSON = "application/json; charset=utf-8"
)

type RequestOption func(*http.Request) error

func WithHeader(key, value string) RequestOption {
	return func(req *http.Request) error {
		req.Header.Set(key, value)
		return nil
	}
}

func WithAPIToken(token string) RequestOption {
	return WithHeader(headerAPIToken, token)
}

func Post(url string, reqBody interface{}, resBody interface{}, opts ...RequestOption) (*http.Response, error) {
	return do(http.MethodPost, url, reqBody, resBody, opts...)
}

func Get(url string, resBody interface{}, opts ...RequestOption) (*http.Response, error) {
	return do(http.MethodGet, url, nil, resBody, opts...)
}

func do(method, url string, reqBody interface{}, resBody interface{}, opts ...RequestOption) (*http.Response, error) {
	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set(headerContentType, contentTypeJSON)

	for _, opt := range opts {
		if err := opt(req); err != nil {
			return nil, err
		}
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != expected(method) {
		return nil, fmt.Errorf("unexpected status code: %d\n\n%s", res.StatusCode, dump(res))
	}

	ct := res.Header.Get(headerContentType)
	if ct != contentTypeJSON {
		return nil, fmt.Errorf("unexpected content type: '%s'\n\n%s", ct, dump(res))
	}

	if err := json.NewDecoder(res.Body).Decode(resBody); err != nil {
		return nil, err
	}

	return res, nil
}

func dump(r *http.Response) string {
	b, err := httputil.DumpResponse(r, true)
	if err != nil {
		return fmt.Sprintf("failed to load response: %s", err)
	}
	return string(b)
}

func expected(method string) int {
	return map[string]int{
		http.MethodPost: http.StatusCreated,
		http.MethodGet:  http.StatusOK,
	}[method]
}
