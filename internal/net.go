package internal

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/docker/go-connections/proxy"
)

// Ctx is a synonym for convenience.
type Ctx = context.Context

var errTCPPortOpen = errors.New("tcp port is open")

// WaitTCPPort tries to connect to addr until success or ctx.Done.
func WaitTCPPort(ctx Ctx, addr fmt.Stringer) error {
	const delay = time.Second / 20
	backOff := backoff.WithContext(backoff.NewConstantBackOff(delay), ctx)
	op := func() error {
		var dialer net.Dialer
		conn, err := dialer.DialContext(ctx, "tcp", addr.String())
		if err == nil {
			err = conn.Close()
		}
		return err
	}
	return backoff.Retry(op, backOff)
}

// WaitTCPPortClosed tries to connect to addr until failed or ctx.Done.
func WaitTCPPortClosed(ctx Ctx, addr fmt.Stringer) error {
	const delay = time.Second / 20
	backOff := backoff.WithContext(backoff.NewConstantBackOff(delay), ctx)
	op := func() error {
		var dialer net.Dialer
		conn, err := dialer.DialContext(ctx, "tcp", addr.String())
		if err != nil {
			return nil
		}
		_ = conn.Close()
		return fmt.Errorf("%w: %s", errTCPPortOpen, addr)
	}
	return backoff.Retry(op, backOff)
}

// NewTCPProxy creates, Run() and returns new TCP proxy.
func NewTCPProxy(ctx Ctx, frontendAddr string, backendAddr string) (*proxy.TCPProxy, error) {
	frontendTCPAddr, err := net.ResolveTCPAddr("tcp", frontendAddr)
	if err != nil {
		return nil, fmt.Errorf("ResolveTCPAddr(%s): %w", frontendAddr, err)
	}
	backendTCPAddr, err := net.ResolveTCPAddr("tcp", backendAddr)
	if err != nil {
		return nil, fmt.Errorf("ResolveTCPAddr(%s): %w", backendAddr, err)
	}
	tcpProxy, err := proxy.NewTCPProxy(frontendTCPAddr, backendTCPAddr)
	if err != nil {
		return nil, fmt.Errorf("NewTCPProxy(%s, %s): %w", frontendAddr, backendAddr, err)
	}
	go tcpProxy.Run()
	err = WaitTCPPort(ctx, tcpProxy.FrontendAddr())
	if err != nil {
		tcpProxy.Close()
		return nil, err
	}
	return tcpProxy, nil
}
