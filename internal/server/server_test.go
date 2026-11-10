package server

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestShutdownIsConcurrentAndIdempotent(t *testing.T) {
	srv, err := NewServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	startDone := make(chan error, 1)
	go func() {
		startDone <- srv.Start()
	}()

	address := waitForListener(t, srv)
	var clientsMu sync.Mutex
	clients := make([]net.Conn, 0, 100)
	var dialers sync.WaitGroup
	for i := 0; i < 100; i++ {
		dialers.Add(1)
		go func() {
			defer dialers.Done()
			conn, dialErr := net.DialTimeout("tcp", address, time.Second)
			if dialErr != nil {
				return // Shutdown may close the listener before this dial wins the race.
			}
			clientsMu.Lock()
			clients = append(clients, conn)
			clientsMu.Unlock()
		}()
	}

	var shutdowns sync.WaitGroup
	for i := 0; i < 8; i++ {
		shutdowns.Add(1)
		go func() {
			defer shutdowns.Done()
			srv.Shutdown()
		}()
	}

	dialers.Wait()
	shutdowns.Wait()
	clientsMu.Lock()
	for _, conn := range clients {
		_ = conn.Close()
	}
	clientsMu.Unlock()

	select {
	case startErr := <-startDone:
		if startErr != nil {
			t.Fatalf("Start: %v", startErr)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after Shutdown")
	}
}

func TestShutdownBeforeStart(t *testing.T) {
	srv, err := NewServer("127.0.0.1:0")
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	srv.Shutdown()
	if err := srv.Start(); err != nil {
		t.Fatalf("Start after Shutdown: %v", err)
	}
}

func waitForListener(t *testing.T, srv *Server) string {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		srv.stateMu.Lock()
		listener := srv.listener
		srv.stateMu.Unlock()
		if listener != nil {
			return listener.Addr().String()
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("server listener did not start")
	return ""
}
