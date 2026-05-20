package helper

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tanenking/gsframe/internal/constants"
	"github.com/tanenking/gsframe/internal/logx"
)

const (
	maxConcurrency = 500 // 最大并发数（关键：控制goroutine数量）
)

var client *http.Client
var httpTransport *http.Transport

// 控制并发量
var sem = make(chan struct{}, maxConcurrency)

func init() {
	httpTransport = &http.Transport{
		MaxIdleConns:          1000,
		MaxIdleConnsPerHost:   200,
		IdleConnTimeout:       30 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		// 设置超时，避免请求永久阻塞
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second, // 连接超时
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client = &http.Client{Timeout: 30 * time.Second}
	if httpTransport.MaxIdleConns > 1000 {
		httpTransport.MaxIdleConns = 1000
	}
	client.Transport = httpTransport
}

func SetHttpTransport(t *http.Transport) {
	httpTransport = t
	if httpTransport.MaxIdleConns > 1000 {
		httpTransport.MaxIdleConns = 1000
	}
	client.Transport = httpTransport
}

func DoPost(Domain string, headers map[string]string, params map[string]string, data string) (resp *http.Response, response []byte, err error) {
	sem <- struct{}{}
	wait := make(chan struct{})

	constants.Go(func() {
		defer func() {
			wait <- struct{}{}
			<-sem
		}()
		Url, err := url.Parse(Domain)
		if err != nil {
			return
		}
		querys := url.Values{}
		for k, v := range params {
			querys.Set(k, v)
		}
		if len(Url.RawQuery) > 0 && len(querys) > 0 {
			Url.RawQuery += "&"
		}
		Url.RawQuery += querys.Encode()
		urlPath := Url.String()

		req, err := http.NewRequest("POST", urlPath, strings.NewReader(data))
		if err != nil {
			logx.ErrorF("DoPost NewRequest err = %v", err)
			return
		}
		if len(headers) > 0 {
			for k, v := range headers {
				req.Header.Set(k, v)
			}
		}
		// logx.DebugF("DoPost Domain = %s", Domain)
		// logx.DebugF("DoPost data = %s", data)

		resp, err = client.Do(req)
		if err != nil {
			logx.ErrorF("DoPost client.Do err = %v", err)
			return
		}
		defer resp.Body.Close()

		response, err = ioutil.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			logx.ErrorF("DoPost StatusCode = %d, status = %s, resp msg = %+v", resp.StatusCode, resp.Status, string(response))
		}
	})
	<-wait
	return
}
func DoGet(Domain string, headers map[string]string, params map[string]string) (resp *http.Response, response []byte, err error) {
	sem <- struct{}{}
	wait := make(chan struct{})

	constants.Go(func() {
		defer func() {
			wait <- struct{}{}
			<-sem
		}()
		Url, err := url.Parse(Domain)
		if err != nil {
			return
		}
		querys := url.Values{}
		for k, v := range params {
			querys.Set(k, v)
		}
		if len(Url.RawQuery) > 0 && len(querys) > 0 {
			Url.RawQuery += "&"
		}
		Url.RawQuery += querys.Encode()
		urlPath := Url.String()
		// logx.DebugF("urlPath = %s", urlPath)

		req, err := http.NewRequest("GET", urlPath, nil)
		if err != nil {
			logx.ErrorF("DoGet NewRequest err = %v", err)
			return
		}
		if len(headers) > 0 {
			for k, v := range headers {
				req.Header.Set(k, v)
			}
		}
		resp, err = client.Do(req)
		if err != nil {
			logx.ErrorF("DoGet client.Do err = %v", err)
			return
		}
		defer resp.Body.Close()

		response, err = ioutil.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			logx.ErrorF("DoGet StatusCode = %d, status = %s, resp msg = %+v", resp.StatusCode, resp.Status, string(response))
		}
	})
	<-wait
	return
}

func DoGetDirect(Domain string, headers map[string]string, query string) (resp *http.Response, response []byte, err error) {
	sem <- struct{}{}
	wait := make(chan struct{})

	constants.Go(func() {
		defer func() {
			wait <- struct{}{}
			<-sem
		}()
		Url, err := url.Parse(Domain)
		if err != nil {
			return
		}
		if len(Url.RawQuery) > 0 && len(query) > 0 {
			Url.RawQuery += "&"
		}
		Url.RawQuery += query
		urlPath := Url.String()

		req, err := http.NewRequest("GET", urlPath, nil)
		if err != nil {
			logx.ErrorF("DoGetDirect NewRequest err = %v", err)
			return
		}
		if len(headers) > 0 {
			for k, v := range headers {
				req.Header.Set(k, v)
			}
		}
		resp, err = client.Do(req)
		if err != nil {
			logx.ErrorF("DoGetDirect client.Do err = %v", err)
			return
		}
		response, err = ioutil.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusOK {
			logx.ErrorF("DoGetDirect StatusCode = %d, status = %s, resp msg = %+v", resp.StatusCode, resp.Status, string(response))
		}
	})
	<-wait
	return
}
