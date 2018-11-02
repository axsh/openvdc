package api

import (
	"net"
	"testing"
	"time"

	"github.com/axsh/openvdc/internal/unittest"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/model/backend"
	"github.com/axsh/openvdc/scheduler"
	"golang.org/x/net/context"
)

func TestNewAPIServer(t *testing.T) {
	ze := &backend.ZkEndpoint{}
	ze.Set(unittest.TestZkServer)

	sched := scheduler.Schedule{}

	// TODO: Set mock SchedulerDriver
	s := NewAPIServer(ze, nil, sched, model.WithMockClusterBackendCtx(context.Background()))
	if s == nil {
		t.Error("NewAPIServer() returned nil")
	}
}

func TestAPIServerRun(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:8765")
	if err != nil {
		t.Error(err)
	}
	ze := &backend.ZkEndpoint{}
	ze.Set(unittest.TestZkServer)

	sched := scheduler.Schedule{}

	// TODO: Set mock SchedulerDriver
	s := NewAPIServer(ze, nil, sched, model.WithMockClusterBackendCtx(context.Background()))
	go func() {
		time.Sleep(2 * time.Second)
		s.Stop()
	}()
	s.Serve(lis)
}
