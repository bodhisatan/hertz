/*
 * Copyright 2022 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * The MIT License (MIT)
 *
 * Copyright (c) 2015-present Aliaksandr Valialkin, VertaMedia, Kirill Danshin, Erik Dubbelboer, FastHTTP Authors
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 *
 * This file may have been modified by CloudWeGo authors. All CloudWeGo
 * Modifications are Copyright 2022 CloudWeGo Authors.
 */

package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/cloudwego/hertz/internal/bytestr"
	"github.com/cloudwego/hertz/internal/nocopy"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/errors"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/client"
	"github.com/cloudwego/hertz/pkg/protocol/http1"
)

var errorInvalidURI = errors.NewPublic("invalid uri")

// Do performs the given http request and fills the given http response.
//
// Request must contain at least non-zero RequestURI with full url (including
// scheme and host) or non-zero Host header + RequestURI.©
//
// Client determines the server to be requested in the following order:
//
//   - from RequestURI if it contains full url with scheme and host;
//   - from Host header otherwise.
//
// The function doesn't follow redirects. Use Get* for following redirects.
//
// Response is ignored if resp is nil.
//
// ErrNoFreeConns is returned if all DefaultMaxConnsPerHost connections
// to the requested host are busy.
//
// It is recommended obtaining req and resp via AcquireRequest
// and AcquireResponse in performance-critical code.
func Do(ctx context.Context, req *protocol.Request, resp *protocol.Response) error {
	return defaultClient.Do(ctx, req, resp)
}

// DoTimeout performs the given request and waits for response during
// the given timeout duration.
//
// Request must contain at least non-zero RequestURI with full url (including
// scheme and host) or non-zero Host header + RequestURI.
//
// Client determines the server to be requested in the following order:
//
//   - from RequestURI if it contains full url with scheme and host;
//   - from Host header otherwise.
//
// The function doesn't follow redirects. Use Get* for following redirects.
//
// Response is ignored if resp is nil.
//
// errTimeout is returned if the response wasn't returned during
// the given timeout.
//
// ErrNoFreeConns is returned if all DefaultMaxConnsPerHost connections
// to the requested host are busy.
//
// It is recommended obtaining req and resp via AcquireRequest
// and AcquireResponse in performance-critical code.
//
// Warning: DoTimeout does not terminate the request itself. The request will
// continue in the background and the response will be discarded.
// If requests take too long and the connection pool gets filled up please
// try using a Client and setting a ReadTimeout.
func DoTimeout(ctx context.Context, req *protocol.Request, resp *protocol.Response, timeout time.Duration) error {
	return defaultClient.DoTimeout(ctx, req, resp, timeout)
}

// DoDeadline performs the given request and waits for response until
// the given deadline.
//
// Request must contain at least non-zero RequestURI with full url (including
// scheme and host) or non-zero Host header + RequestURI.
//
// Client determines the server to be requested in the following order:
//
//   - from RequestURI if it contains full url with scheme and host;
//   - from Host header otherwise.
//
// The function doesn't follow redirects. Use Get* for following redirects.
//
// Response is ignored if resp is nil.
//
// errTimeout is returned if the response wasn't returned until
// the given deadline.
//
// ErrNoFreeConns is returned if all DefaultMaxConnsPerHost connections
// to the requested host are busy.
//
// It is recommended obtaining req and resp via AcquireRequest
// and AcquireResponse in performance-critical code.
func DoDeadline(ctx context.Context, req *protocol.Request, resp *protocol.Response, deadline time.Time) error {
	return defaultClient.DoDeadline(ctx, req, resp, deadline)
}

// DoRedirects performs the given http request and fills the given http response,
// following up to maxRedirectsCount redirects. When the redirect count exceeds
// maxRedirectsCount, ErrTooManyRedirects is returned.
//
// Request must contain at least non-zero RequestURI with full url (including
// scheme and host) or non-zero Host header + RequestURI.
//
// Client determines the server to be requested in the following order:
//
//   - from RequestURI if it contains full url with scheme and host;
//   - from Host header otherwise.
//
// Response is ignored if resp is nil.
//
// ErrNoFreeConns is returned if all DefaultMaxConnsPerHost connections
// to the requested host are busy.
//
// It is recommended obtaining req and resp via AcquireRequest
// and AcquireResponse in performance-critical code.
func DoRedirects(ctx context.Context, req *protocol.Request, resp *protocol.Response, maxRedirectsCount int) error {
	_, _, err := client.DoRequestFollowRedirects(ctx, req, resp, req.URI().String(), maxRedirectsCount, &defaultClient)
	return err
}

// Get returns the status code and body of url.
//
// The contents of dst will be replaced by the body and returned, if the dst
// is too small a new slice will be allocated.
//
// The function follows redirects. Use Do* for manually handling redirects.
func Get(ctx context.Context, dst []byte, url string) (statusCode int, body []byte, err error) {
	return defaultClient.Get(ctx, dst, url)
}

// GetTimeout returns the status code and body of url.
//
// The contents of dst will be replaced by the body and returned, if the dst
// is too small a new slice will be allocated.
//
// The function follows redirects. Use Do* for manually handling redirects.
//
// errTimeout error is returned if url contents couldn't be fetched
// during the given timeout.
func GetTimeout(ctx context.Context, dst []byte, url string, timeout time.Duration) (statusCode int, body []byte, err error) {
	return defaultClient.GetTimeout(ctx, dst, url, timeout)
}

// GetDeadline returns the status code and body of url.
//
// The contents of dst will be replaced by the body and returned, if the dst
// is too small a new slice will be allocated.
//
// The function follows redirects. Use Do* for manually handling redirects.
//
// errTimeout error is returned if url contents couldn't be fetched
// until the given deadline.
func GetDeadline(ctx context.Context, dst []byte, url string, deadline time.Time) (statusCode int, body []byte, err error) {
	return defaultClient.GetDeadline(ctx, dst, url, deadline)
}

// Post sends POST request to the given url with the given POST arguments.
//
// The contents of dst will be replaced by the body and returned, if the dst
// is too small a new slice will be allocated.
//
// The function follows redirects. Use Do* for manually handling redirects.
//
// Empty POST body is sent if postArgs is nil.
func Post(ctx context.Context, dst []byte, url string, postArgs *protocol.Args) (statusCode int, body []byte, err error) {
	return defaultClient.Post(ctx, dst, url, postArgs)
}

var defaultClient = Client{DialTimeout: time.Second}

// Client implements http client.
//
// Copying Client by value is prohibited. Create new instance instead.
//
// It is safe calling Client methods from concurrently running goroutines.
type Client struct {
	noCopy nocopy.NoCopy //lint:ignore U1000 until noCopy is used

	// Client name. Used in User-Agent request header.
	//
	// Default client name is used if not set.
	Name string

	// NoDefaultUserAgentHeader when set to true, causes the default
	// User-Agent header to be excluded from the Request.
	NoDefaultUserAgentHeader bool

	// Callback for establishing new connections to hosts.
	//
	// Default Dial is used if not set.
	Dial client.DialFunc

	// Timeout for establishing new connections to hosts.
	//
	// Default DialTimeout is used if not set.
	DialTimeout time.Duration

	// Attempt to connect to both ipv4 and ipv6 addresses if set to true.
	//
	// This option is used only if default TCP dialer is used,
	// i.e. if Dial is blank.
	//
	// By default client connects only to ipv4 addresses,
	// since unfortunately ipv6 remains broken in many networks worldwide :)
	DialDualStack bool

	// TLS config for https connections.
	//
	// Default TLS config is used if not set.
	TLSConfig *tls.Config

	// Maximum number of connections per each host which may be established.
	//
	// DefaultMaxConnsPerHost is used if not set.
	MaxConnsPerHost int

	// Idle keep-alive connections are closed after this duration.
	//
	// By default idle connections are closed
	// after DefaultMaxIdleConnDuration.
	MaxIdleConnDuration time.Duration

	// Keep-alive connections are closed after this duration.
	//
	// By default connection duration is unlimited.
	MaxConnDuration time.Duration

	// Maximum number of attempts for idempotent calls
	//
	// DefaultMaxIdempotentCallAttempts is used if not set.
	MaxIdempotentCallAttempts int

	// Per-connection buffer size for responses' reading.
	// This also limits the maximum header size.
	//
	// Default buffer size is used if 0.
	ReadBufferSize int

	// Per-connection buffer size for requests' writing.
	//
	// Default buffer size is used if 0.
	WriteBufferSize int

	// Maximum duration for full response reading (including body).
	//
	// By default response read timeout is unlimited.
	ReadTimeout time.Duration

	// Maximum duration for full request writing (including body).
	//
	// By default request write timeout is unlimited.
	WriteTimeout time.Duration

	// Maximum response body size.
	//
	// The client returns ErrBodyTooLarge if this limit is greater than 0
	// and response body is greater than the limit.
	//
	// By default response body size is unlimited.
	MaxResponseBodySize int

	// Header names are passed as-is without normalization
	// if this option is set.
	//
	// Disabled header names' normalization may be useful only for proxying
	// responses to other clients expecting case-sensitive header names.
	//
	// By default request and response header names are normalized, i.e.
	// The first letter and the first letters following dashes
	// are uppercased, while all the other letters are lowercased.
	// Examples:
	//
	//     * HOST -> Host
	//     * content-type -> Content-Type
	//     * cONTENT-lenGTH -> Content-Length
	DisableHeaderNamesNormalizing bool

	// Path values are sent as-is without normalization
	//
	// Disabled path normalization may be useful for proxying incoming requests
	// to servers that are expecting paths to be forwarded as-is.
	//
	// By default path values are normalized, i.e.
	// extra slashes are removed, special characters are encoded.
	DisablePathNormalizing bool

	// Maximum duration for waiting for a free connection.
	//
	// By default will not waiting, return ErrNoFreeConns immediately
	MaxConnWaitTimeout time.Duration

	// RetryIf controls whether a retry should be attempted after an error.
	//
	// By default will use isIdempotent function
	RetryIf client.RetryIfFunc

	// Connection will close after each request when set this to true.
	KeepAlive bool

	// ResponseBodyStream enables response body streaming
	ResponseBodyStream bool

	mLock sync.Mutex
	m     map[string]client.HostClient
	ms    map[string]client.HostClient
	mws   Middleware

	// Proxy specifies a function to return a proxy for a given
	// Request. If the function returns a non-nil error, the
	// request is aborted with the provided error.
	//
	// The proxy type is determined by the URL scheme.
	// "http" and "https" are supported. If the scheme is empty,
	// "http" is assumed.
	//
	// If Proxy is nil or returns a nil *URL, no proxy is used.
	Proxy protocol.Proxy
}

// SetProxy is used to set client proxy.
//
// Don't SetProxy twice for a client.
// If you want to use another proxy, please create another client and set proxy to it.
func (c *Client) SetProxy(p protocol.Proxy) {
	c.Proxy = p
}

// Get returns the status code and body of url.
//
// The contents of dst will be replaced by the body and returned, if the dst
// is too small a new slice will be allocated.
//
// The function follows redirects. Use Do* for manually handling redirects.
func (c *Client) Get(ctx context.Context, dst []byte, url string, requestOptions ...config.RequestOption) (statusCode int, body []byte, err error) {
	return client.GetURL(ctx, dst, url, c, requestOptions...)
}

// GetTimeout returns the status code and body of url.
//
// The contents of dst will be replaced by the body and returned, if the dst
// is too small a new slice will be allocated.
//
// The function follows redirects. Use Do* for manually handling redirects.
//
// errTimeout error is returned if url contents couldn't be fetched
// during the given timeout.
func (c *Client) GetTimeout(ctx context.Context, dst []byte, url string, timeout time.Duration, requestOptions ...config.RequestOption) (statusCode int, body []byte, err error) {
	return client.GetURLTimeout(ctx, dst, url, timeout, c, requestOptions...)
}

// GetDeadline returns the status code and body of url.
//
// The contents of dst will be replaced by the body and returned, if the dst
// is too small a new slice will be allocated.
//
// The function follows redirects. Use Do* for manually handling redirects.
//
// errTimeout error is returned if url contents couldn't be fetched
// until the given deadline.
func (c *Client) GetDeadline(ctx context.Context, dst []byte, url string, deadline time.Time, requestOptions ...config.RequestOption) (statusCode int, body []byte, err error) {
	return client.GetURLDeadline(ctx, dst, url, deadline, c, requestOptions...)
}

// Post sends POST request to the given url with the given POST arguments.
//
// The contents of dst will be replaced by the body and returned, if the dst
// is too small a new slice will be allocated.
//
// The function follows redirects. Use Do* for manually handling redirects.
//
// Empty POST body is sent if postArgs is nil.
func (c *Client) Post(ctx context.Context, dst []byte, url string, postArgs *protocol.Args, requestOptions ...config.RequestOption) (statusCode int, body []byte, err error) {
	return client.PostURL(ctx, dst, url, postArgs, c, requestOptions...)
}

// DoTimeout performs the given request and waits for response during
// the given timeout duration.
//
// Request must contain at least non-zero RequestURI with full url (including
// scheme and host) or non-zero Host header + RequestURI.
//
// Client determines the server to be requested in the following order:
//
//   - from RequestURI if it contains full url with scheme and host;
//   - from Host header otherwise.
//
// The function doesn't follow redirects. Use Get* for following redirects.
//
// Response is ignored if resp is nil.
//
// errTimeout is returned if the response wasn't returned during
// the given timeout.
//
// ErrNoFreeConns is returned if all Client.MaxConnsPerHost connections
// to the requested host are busy.
//
// It is recommended obtaining req and resp via AcquireRequest
// and AcquireResponse in performance-critical code.
//
// Warning: DoTimeout does not terminate the request itself. The request will
// continue in the background and the response will be discarded.
// If requests take too long and the connection pool gets filled up please
// try setting a ReadTimeout.
func (c *Client) DoTimeout(ctx context.Context, req *protocol.Request, resp *protocol.Response, timeout time.Duration) error {
	return client.DoTimeout(ctx, req, resp, timeout, c)
}

// DoDeadline performs the given request and waits for response until
// the given deadline.
//
// Request must contain at least non-zero RequestURI with full url (including
// scheme and host) or non-zero Host header + RequestURI.
//
// Client determines the server to be requested in the following order:
//
//   - from RequestURI if it contains full url with scheme and host;
//   - from Host header otherwise.
//
// The function doesn't follow redirects. Use Get* for following redirects.
//
// Response is ignored if resp is nil.
//
// errTimeout is returned if the response wasn't returned until
// the given deadline.
//
// ErrNoFreeConns is returned if all Client.MaxConnsPerHost connections
// to the requested host are busy.
//
// It is recommended obtaining req and resp via AcquireRequest
// and AcquireResponse in performance-critical code.
func (c *Client) DoDeadline(ctx context.Context, req *protocol.Request, resp *protocol.Response, deadline time.Time) error {
	return client.DoDeadline(ctx, req, resp, deadline, c)
}

// DoRedirects performs the given http request and fills the given http response,
// following up to maxRedirectsCount redirects. When the redirect count exceeds
// maxRedirectsCount, ErrTooManyRedirects is returned.
//
// Request must contain at least non-zero RequestURI with full url (including
// scheme and host) or non-zero Host header + RequestURI.
//
// Client determines the server to be requested in the following order:
//
//   - from RequestURI if it contains full url with scheme and host;
//   - from Host header otherwise.
//
// Response is ignored if resp is nil.
//
// ErrNoFreeConns is returned if all DefaultMaxConnsPerHost connections
// to the requested host are busy.
//
// It is recommended obtaining req and resp via AcquireRequest
// and AcquireResponse in performance-critical code.
func (c *Client) DoRedirects(ctx context.Context, req *protocol.Request, resp *protocol.Response, maxRedirectsCount int) error {
	_, _, err := client.DoRequestFollowRedirects(ctx, req, resp, req.URI().String(), maxRedirectsCount, c)
	return err
}

// Do performs the given http request and fills the given http response.
//
// Request must contain at least non-zero RequestURI with full url (including
// scheme and host) or non-zero Host header + RequestURI.
//
// Client determines the server to be requested in the following order:
//
//   - from RequestURI if it contains full url with scheme and host;
//   - from Host header otherwise.
//
// Response is ignored if resp is nil.
//
// The function doesn't follow redirects. Use Get* for following redirects.
//
// ErrNoFreeConns is returned if all Client.MaxConnsPerHost connections
// to the requested host are busy.
//
// It is recommended obtaining req and resp via AcquireRequest
// and AcquireResponse in performance-critical code.
func (c *Client) Do(ctx context.Context, req *protocol.Request, resp *protocol.Response) error {
	if c.mws == nil {
		return c.do(ctx, req, resp)
	}
	return c.mws(c.do)(ctx, req, resp)
}

func (c *Client) do(ctx context.Context, req *protocol.Request, resp *protocol.Response) error {
	if !c.KeepAlive {
		req.Header.SetConnectionClose(true)
	}
	uri := req.URI()
	if uri == nil {
		return errorInvalidURI
	}

	var proxyURI *protocol.URI
	var err error

	if c.Proxy != nil {
		proxyURI, err = c.Proxy(req)
		if err != nil {
			return fmt.Errorf("proxy error=%w", err)
		}
	}

	isTLS := false
	scheme := uri.Scheme()
	if bytes.Equal(scheme, bytestr.StrHTTPS) {
		isTLS = true
	} else if !bytes.Equal(scheme, bytestr.StrHTTP) && !bytes.Equal(scheme, bytestr.StrSD) {
		return fmt.Errorf("unsupported protocol %q. http and https are supported", scheme)
	}
	host := uri.Host()
	startCleaner := false

	c.mLock.Lock()
	m := c.m
	if isTLS {
		m = c.ms
	}
	if m == nil {
		m = make(map[string]client.HostClient)
		if isTLS {
			c.ms = m
		} else {
			c.m = m
		}
	}
	h := string(host)
	hc := m[h]
	if hc == nil {
		hConfig := &client.HostClientConfig{
			Addr:                          h,
			Name:                          c.Name,
			NoDefaultUserAgentHeader:      c.NoDefaultUserAgentHeader,
			DialDualStack:                 c.DialDualStack,
			DisableHeaderNamesNormalizing: c.DisableHeaderNamesNormalizing,
			DisablePathNormalizing:        c.DisablePathNormalizing,
			TLSConfig:                     c.TLSConfig,
			IsTLS:                         isTLS,
			MaxConns:                      c.MaxConnsPerHost,
			MaxIdempotentCallAttempts:     c.MaxIdempotentCallAttempts,
			ReadBufferSize:                c.ReadBufferSize,
			WriteBufferSize:               c.WriteBufferSize,
			MaxResponseBodySize:           c.MaxResponseBodySize,
			Dial:                          c.Dial,
			RetryIf:                       c.RetryIf,
			DialTimeout:                   c.DialTimeout,
			MaxIdleConnDuration:           c.MaxIdleConnDuration,
			MaxConnDuration:               c.MaxConnDuration,
			ReadTimeout:                   c.ReadTimeout,
			WriteTimeout:                  c.WriteTimeout,
			MaxConnWaitTimeout:            c.MaxConnWaitTimeout,
			ResponseBodyStream:            c.ResponseBodyStream,
			ProxyURI:                      proxyURI,
		}
		hc = http1.NewHostClient(hConfig)
		m[string(host)] = hc
		if len(m) == 1 {
			startCleaner = true
		}
	}
	c.mLock.Unlock()

	if startCleaner {
		go c.mCleaner()
	}

	return hc.Do(ctx, req, resp)
}

// CloseIdleConnections closes any connections which were previously
// connected from previous requests but are now sitting idle in a
// "keep-alive" state. It does not interrupt any connections currently
// in use.
func (c *Client) CloseIdleConnections() {
	c.mLock.Lock()
	for _, v := range c.m {
		v.CloseIdleConnections()
	}
	c.mLock.Unlock()
}

func (c *Client) mCleaner() {
	mustStop := false

	for {
		c.mLock.Lock()
		for k, v := range c.m {
			shouldRemove := v.ShouldRemove()

			if shouldRemove {
				delete(c.m, k)
			}
		}
		if len(c.m) == 0 {
			mustStop = true
		}
		c.mLock.Unlock()

		if mustStop {
			break
		}
		time.Sleep(10 * time.Second)
	}
}

// NewClient return a client with options
func NewClient(opts ...config.ClientOption) (*Client, error) {
	opt := config.NewClientOptions(opts)
	c := &Client{
		DialTimeout:               opt.DialTimeout,
		MaxConnsPerHost:           opt.MaxConnsPerHost,
		MaxIdleConnDuration:       opt.MaxIdleConnDuration,
		MaxConnDuration:           opt.MaxConnDuration,
		MaxConnWaitTimeout:        opt.MaxConnWaitTimeout,
		MaxIdempotentCallAttempts: opt.MaxIdempotentCallAttempts,
		KeepAlive:                 opt.KeepAlive,
		ReadTimeout:               opt.ReadTimeout,
		ResponseBodyStream:        opt.ResponseBodyStream,
		TLSConfig:                 opt.TLSConfig,
		Dial:                      opt.DialFunc,
	}
	return c, nil
}

func (c *Client) Use(mws ...Middleware) {
	// Put the original middlewares to the first
	middlewares := make([]Middleware, 0, 1+len(mws))
	if c.mws != nil {
		middlewares = append(middlewares, c.mws)
	}
	middlewares = append(middlewares, mws...)
	c.mws = chain(middlewares...)
}
