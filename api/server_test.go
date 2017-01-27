package api

import (
	"net"
	"testing"
	"time"

	"github.com/axsh/openvdc/internal/unittest"
	"github.com/axsh/openvdc/model/backend"
)

func TestNewAPIServer(t *testing.T) {
	// TODO: Set mock SchedulerDriver
	ze := &backend.ZkEndpoint{}
	ze.Set(unittest.TestZkServer)
	s := NewAPIServer(ze, nil)
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
	ze := &backend.ZkEndpoint{}
	ze.Set(unittest.TestZkServer)
	s := NewAPIServer(ze, nil)
	go func() {
		time.Sleep(2 * time.Second)
		s.Stop()
	}()
	s.Serve(lis)
}
