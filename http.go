// Package http is a clone of net/http for GopherJS. It replicates
// most of the net/http.Client API, using XHR in the background, and
// making use of actual types from the net/http package, such as
// net/http.Request and net/http.Response.
//
// At this moment, the future of this package is not yet certain.
// Better solutions might present themselves, in which case this
// package will be removed and replaced with something better. Use at
// your own risk.
package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/gopherjs/gopherjs/js"

	"honnef.co/go/js/xhr"
)

type Client struct {
	// The amount of time a request can take before it will be
	// terminated. Millisecond precision is supported.
	Timeout time.Duration
}

var DefaultClient = Client{}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var err error
	x := xhr.NewRequest(req.Method, req.URL.String())
	x.ResponseType = xhr.ArrayBuffer
	if c.Timeout > 0 {
		x.Timeout = int(c.Timeout / time.Millisecond)
	}
	for k, v := range req.Header {
		for _, vv := range v {
			x.SetRequestHeader(k, vv)
		}
	}
	var data []byte
	if req.Body != nil {
		defer req.Body.Close()
		data, err = ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
	}

	// FIXME req.Host

	err = x.Send(data)
	if err != nil {
		return nil, err
	}

	r := strings.NewReader(x.ResponseHeaders())
	headers, err := textproto.NewReader(bufio.NewReader(r)).ReadMIMEHeader()
	if err != nil && err != io.EOF {
		return nil, err
	}

	b := js.Global.Get("Uint8Array").New(x.Response).Interface().([]byte)
	res := &http.Response{
		Status:        fmt.Sprintf("%d %s", x.Status, x.StatusText),
		StatusCode:    x.Status,
		Header:        http.Header(headers),
		Body:          ioutil.NopCloser(bytes.NewReader(b)),
		ContentLength: int64(len(b)),
		// TODO other stuff
	}
	return res, nil
}

func (c *Client) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) Head(url string) (*http.Response, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) Post(url string, bodyType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", bodyType)
	return c.Do(req)
}

func (c *Client) PostForm(url string, data url.Values) (*http.Response, error) {
	return c.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func Get(url string) (*http.Response, error) {
	return DefaultClient.Get(url)
}

func Head(url string) (*http.Response, error) {
	return DefaultClient.Head(url)
}

func Post(url string, bodyType string, body io.Reader) (*http.Response, error) {
	return DefaultClient.Post(url, bodyType, body)
}
