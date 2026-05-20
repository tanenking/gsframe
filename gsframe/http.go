package gsframe

import (
	"net/http"

	"github.com/tanenking/gsframe/internal/helper"
)

func SetHttpTransport(t *http.Transport) {
	helper.SetHttpTransport(t)
}

func HttpPost(Domain string, headers map[string]string, params map[string]string, data string) (resp *http.Response, response []byte, err error) {
	return helper.DoPost(Domain, headers, params, data)
}

func HttpGet(Domain string, headers map[string]string, params map[string]string) (resp *http.Response, response []byte, err error) {
	return helper.DoGet(Domain, headers, params)
}

func HttpGetDirect(Domain string, headers map[string]string, query string) (resp *http.Response, response []byte, err error) {
	return helper.DoGetDirect(Domain, headers, query)
}
