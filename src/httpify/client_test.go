package httpify

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/Shopify/themekit/src/release"
	"github.com/stretchr/testify/assert"
)

func TestClient_do(t *testing.T) {
	body := map[string]interface{}{"key": "main.js", "value": "alert('this is javascript');"}

	var client *HTTPClient
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Header.Get("X-Shopify-Access-Token"), client.password)
		assert.Equal(t, r.Header.Get("Content-Type"), "application/json")
		assert.Equal(t, r.Header.Get("Accept"), "application/json")
		assert.Equal(t, r.Header.Get("User-Agent"), fmt.Sprintf("go/themekit (%s; %s; %s)", runtime.GOOS, runtime.GOARCH, release.ThemeKitVersion.String()))

		reqBody, err := ioutil.ReadAll(r.Body)
		assert.Nil(t, err)

		asset := struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{}
		assert.Nil(t, json.Unmarshal(reqBody, &asset))
		assert.Equal(t, asset.Key, body["key"])
		assert.Equal(t, asset.Value, body["value"])
	}))
	defer server.Close()

	client, err := NewClient(Params{
		Domain:   server.URL,
		Password: "secret_password",
		APILimit: time.Nanosecond,
	})
	client.baseURL.Scheme = "http"

	assert.NotNil(t, client)
	assert.Nil(t, err)

	resp, err := client.do("POST", "/assets.json", body)
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func TestGenerateClientTransport(t *testing.T) {
	testcases := []struct {
		proxyURL, err string
		expectNil     bool
	}{
		{proxyURL: "", expectNil: true},
		{proxyURL: "http//localhost:3000", expectNil: true, err: "invalid proxy URI"},
		{proxyURL: "http://127.0.0.1:8080", expectNil: false},
	}

	for _, testcase := range testcases {
		transport, err := generateClientTransport(testcase.proxyURL)
		assert.Equal(t, transport == nil, testcase.expectNil)
		if testcase.err == "" {
			assert.Nil(t, err)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}

func TestParseBaseUrl(t *testing.T) {
	testcases := []struct {
		domain, expected, err string
	}{
		{domain: "test.myshopify.com", expected: "https://test.myshopify.com"},
		{domain: "$%@#.myshopify.com", expected: "", err: "invalid domain"},
	}

	for _, testcase := range testcases {
		actual, err := parseBaseURL(testcase.domain)
		if testcase.err == "" && assert.Nil(t, err) {
			assert.Equal(t, actual.String(), testcase.expected)
		} else if assert.NotNil(t, err) {
			assert.Contains(t, err.Error(), testcase.err)
		}
	}
}