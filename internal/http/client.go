package http

import (
	"fmt"
	"golang.org/x/net/http2"
	"net"
	"net/http"
	"sync"
	"time"
)

func NewHTTPClient(httpSettings Settings) (*http.Client, error) {
	var client http.Client
	tr := &http.Transport{
		ResponseHeaderTimeout: httpSettings.ResponseHeader,
		Proxy:                 http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			KeepAlive: httpSettings.ConnKeepAlive,
			Timeout:   httpSettings.Connect,
		}).DialContext,
		MaxIdleConns:          httpSettings.MaxAllIdleConns,
		IdleConnTimeout:       httpSettings.IdleConn,
		TLSHandshakeTimeout:   httpSettings.TLSHandshake,
		MaxIdleConnsPerHost:   httpSettings.MaxHostIdleConns,
		ExpectContinueTimeout: httpSettings.ExpectContinue,
	}

	err := http2.ConfigureTransport(tr)
	if err != nil {
		return &client, err
	}

	client = http.Client{
		Transport: tr,
	}

	return &client, nil
}

func StartTLSContext(region string, client *http.Client) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go dummyTLSInvocation(fmt.Sprintf("https://dynamodb.%s.amazonaws.com", region), client, wg)
	go dummyTLSInvocation(fmt.Sprintf("https://s3.%s.amazonaws.com", region), client, wg)
	wg.Wait()
}

func dummyTLSInvocation(endpoint string, client *http.Client, wg *sync.WaitGroup) {
	_, _ = client.Head(endpoint)
	wg.Done()
}

func NewDefaultHTTPClient() *http.Client {
	httpClient, err := NewHTTPClient(Settings{
		Connect:          5 * time.Second,
		ExpectContinue:   1 * time.Second,
		IdleConn:         90 * time.Second,
		ConnKeepAlive:    60 * time.Second,
		MaxAllIdleConns:  100,
		MaxHostIdleConns: 10,
		ResponseHeader:   5 * time.Second,
		TLSHandshake:     5 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	return httpClient
}
