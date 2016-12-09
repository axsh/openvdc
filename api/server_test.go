package api

import (
	"net"
	"testing"
	"time"

	"github.com/axsh/openvdc/internal/unittest"
)

func TestNewAPIServer(t *testing.T) {
	// TODO: Set mock SchedulerDriver
	s := NewAPIServer(unittest.TestZkServer, nil)
	if s == nil {
		t.Error("NewAPIServer() returned nil")
	}
}

func TestAPIServerRun(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:8765")
	if err != nil {
		t.Error(err)
	}
	// TODO: Set mock SchedulerDriver
	s := NewAPIServer(unittest.TestZkServer, nil)
	go func() {
		time.Sleep(2 * time.Second)
		s.Stop()
	}()
	s.Serve(lis)
}
