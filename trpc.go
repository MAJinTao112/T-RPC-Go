// Package trpc is the supported public API for T-RPC-Go.
// Framework implementation details remain under internal/.
package trpc

import (
	"context"
	"time"

	"github.com/MAJinTao112/T-RPC-Go/internal/client"
	"github.com/MAJinTao112/T-RPC-Go/internal/codec"
	"github.com/MAJinTao112/T-RPC-Go/internal/registry"
	"github.com/MAJinTao112/T-RPC-Go/internal/server"
)

const defaultTimeout = 5 * time.Second

type Instance struct {
	Addr string
}

type Registry struct {
	inner *registry.Registry
}

func NewRegistry(endpoints []string) (*Registry, error) {
	inner, err := registry.NewRegistry(endpoints)
	if err != nil {
		return nil, err
	}
	return &Registry{inner: inner}, nil
}

func (r *Registry) Register(service string, instance Instance, ttl int64) error {
	return r.inner.Register(service, registry.Instance{Addr: instance.Addr}, ttl)
}

func (r *Registry) Discover(service string) ([]Instance, error) {
	instances, err := r.inner.Discover(service)
	if err != nil {
		return nil, err
	}
	result := make([]Instance, 0, len(instances))
	for _, instance := range instances {
		result = append(result, Instance{Addr: instance.Addr})
	}
	return result, nil
}

func (r *Registry) Close() error {
	return r.inner.Close()
}

type clientConfig struct {
	timeout time.Duration
}

type ClientOption func(*clientConfig)

func WithClientTimeout(timeout time.Duration) ClientOption {
	return func(config *clientConfig) {
		if timeout > 0 {
			config.timeout = timeout
		}
	}
}

type Client struct {
	inner *client.Client
}

func NewClient(reg *Registry, options ...ClientOption) (*Client, error) {
	config := clientConfig{timeout: defaultTimeout}
	for _, option := range options {
		option(&config)
	}

	inner, err := client.NewClient(
		reg.inner,
		client.WithClientCodec(codec.JSON),
		client.WithClientTimeout(config.timeout),
	)
	if err != nil {
		return nil, err
	}
	return &Client{inner: inner}, nil
}

func (c *Client) Invoke(ctx context.Context, service, method string, request, reply interface{}) error {
	return c.inner.Invoke(ctx, service, method, request, reply)
}

func (c *Client) Close() {
	c.inner.Close()
}

type Server struct {
	inner *server.Server
}

func NewServer(address string) (*Server, error) {
	inner, err := server.NewServer(address, server.WithServerCodec(codec.JSON))
	if err != nil {
		return nil, err
	}
	return &Server{inner: inner}, nil
}

func (s *Server) Register(name string, service interface{}) {
	s.inner.Register(name, service)
}

func (s *Server) Start() error {
	return s.inner.Start()
}

func (s *Server) Close() {
	s.inner.Close()
}

func (s *Server) Shutdown() {
	s.inner.Shutdown()
}
