package scheduler

import (
	"fmt"
	"time"

	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"github.com/axsh/openvdc/model"
	"github.com/axsh/openvdc/api/agent"
	empty "github.com/golang/protobuf/ptypes/empty"
	mesos "github.com/mesos/mesos-go/mesosproto"
)

type resourceCollector struct {
	id        string 
	grpcConn  *grpc.ClientConn
	resources *model.ComputingResources
	ip        string
	port      int
}

func newResourceCollector(id string, offer *mesos.Offer) *resourceCollector {
	// todo: get proper port from somewhere (attribute?)
	return &resourceCollector{
		ip:   offer.GetUrl().GetAddress().GetIp(),
		port: 9092,
		id:   id,
	}
}

func (rc *resourceCollector) connectResourceCollector() error {
	copts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}

	slaveAddr := fmt.Sprintf("%s:%d", rc.ip, rc.port)
	ctx, _ := context.WithTimeout(context.Background(), time.Second * 1)
	conn, err := grpc.DialContext(ctx, slaveAddr, copts...)
	if err != nil {
		return err
	}
	rc.grpcConn = conn

	return nil
}

func (rc *resourceCollector) updateResources() error {
	c := agent.NewResourceCollectorClient(rc.grpcConn)
	resp, err := c.GetResources(context.Background(), &empty.Empty{})
	if err != nil {
		return err
	}
	rc.resources = resp
	
	return nil
}

