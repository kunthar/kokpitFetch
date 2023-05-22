package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"time"
)

// region Client

// Client converts the high level request data to an HTTP request.
// Internally it contains a Doer object to actually make the connection and
// pass the request.
// Client also implements the auth and retry mechanisms.
type Client interface {
	Request(ctx context.Context, uri string, resp any) error
}

func From(c Doer) Client {
	return &cli{c: c}
}

type cli struct {
	c Doer
}

func (c *cli) Request(ctx context.Context, uri string, resp any) (err error) {
	// log.Printf("[REQ] %s\n", uri)
	err = c.reqLoop(ctx, uri, resp)
	for err != nil {
		log.Printf("req fail: %v\n", err)
		time.Sleep(1000 * time.Millisecond)
		err = c.reqLoop(ctx, uri, resp)
	}
	return
}

func (c *cli) reqLoop(ctx context.Context, uri string, resp any) error {
	rq, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return err
	}
	rs, err := c.c.Do(rq)
	if err != nil {
		return err
	}
	defer func() { _ = rs.Body.Close() }()
	var buf []byte
	if sc := rs.StatusCode; sc == http.StatusOK {
		if buf, err = io.ReadAll(rs.Body); err == nil {
			if err == nil {
				err = json.Unmarshal(buf, resp)
			}
		}
	}
	return err
}

// endregion
// region Doer

// Doer performs the low-level HTTP request.
// *http.Client implements this interface.
type Doer interface {
	Do(r *http.Request) (*http.Response, error)
}

func NewHTTPClient() Doer {
	d := net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	return &http.Client{Transport: &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: time.Second,
		MaxIdleConns:          10,
		DialContext:           d.DialContext,
	}, Timeout: 10 * time.Second}
}

// endregion
