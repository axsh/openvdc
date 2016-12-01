package api

import (
	"net"
	"testing"
	"time"

	"github.com/axsh/openvdc/internal/unittest"
)

func TestNewAPIServer(t *testing.T) {
	c := make(APIOffer)
	// TODO: Set mock SchedulerDriver
	s := NewAPIServer(c, unittest.TestZkServer, nil)
	if s == nil {
		t.Error("NewAPIServer() returned nil")
	}
}

func TestAPIServerRun(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:8765")
	if err != nil {
		t.Error(err)
	}
	c := make(APIOffer)
	// TODO: Set mock SchedulerDriver
	s := NewAPIServer(c, unittest.TestZkServer, nil)
	go func() {
		time.Sleep(2 * time.Second)
		s.Stop()
	}()
	s.Serve(lis)
}
