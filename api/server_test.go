package api

import (
	"net"
	"testing"
	"time"
)

func TestNewAPIServer(t *testing.T) {
	s := NewAPIServer()
	if s == nil {
		t.Error("NewAPIServer() returned nil")
	}
}

func TestAPIServerRun(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:8765")
	if err != nil {
		t.Error(err)
	}
	s := NewAPIServer()
	go func() {
		time.Sleep(2 * time.Second)
		s.Stop()
	}()
	s.Serve(lis)
}
