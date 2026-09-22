package client

import (
	"github.com/MAJinTao112/T-RPC-Go/internal/codec"
	"github.com/MAJinTao112/T-RPC-Go/internal/loadbalance"
	"time"
)

type ClientOption func(*Client) error

func WithClientCodec(t codec.Type) ClientOption {
	return func(c *Client) error {
		cc, err := codec.New(t)
		if err != nil {
			return err
		}
		c.codec = cc
		return nil
	}
}

func WithClientTimeout(d time.Duration) ClientOption {
	return func(c *Client) error {
		c.timeout = d
		return nil
	}
}

func WithClientLoadBalancer(lb loadbalance.LoadBalancer) ClientOption {
	return func(c *Client) error {
		c.lb = lb
		return nil
	}
}
