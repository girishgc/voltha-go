package grpc

import (
	"net"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"strconv"
	"github.com/opencord/voltha-go/common/log"
	"strings"
)

/*
To add a GRPC server to your existing component simply follow these steps:

1. Create a server instance by passing the host and port where it should run and optionally add certificate information

	e.g.
	s.server = server.NewGrpcServer(s.config.GrpcHost, s.config.GrpcPort, nil, false)

2. Create a function that will register your service with the GRPC server

    e.g.
	f := func(gs *grpc.Server) {
		voltha.RegisterVolthaReadOnlyServiceServer(
			gs,
			core.NewReadOnlyServiceHandler(s.root),
		)
	}

3. Add the service to the server

    e.g.
	s.server.AddService(f)

4. Start the server

	s.server.Start(ctx)
 */



type GrpcServer struct {
	gs       *grpc.Server
	address  string
	port     int
	secure   bool
	services []func(*grpc.Server)

	*GrpcSecurity
}

/*
Instantiate a GRPC server data structure
*/
func NewGrpcServer(
	address string,
	port int,
	certs *GrpcSecurity,
	secure bool,
) *GrpcServer {
	server := &GrpcServer{
		address:      address,
		port:         port,
		secure:       secure,
		GrpcSecurity: certs,
	}
	return server
}

/*
Start prepares the GRPC server and starts servicing requests
*/
func (s *GrpcServer) Start(ctx context.Context) {
	host := strings.Join([]string{
		s.address,
		strconv.Itoa(int(s.port)),
	}, ":")

	lis, err := net.Listen("tcp", host)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	if s.secure && s.GrpcSecurity != nil {
		creds, err := credentials.NewServerTLSFromFile(s.CertFile, s.KeyFile)
		if err != nil {
			log.Fatalf("could not load TLS keys: %s", err)
		}
		s.gs = grpc.NewServer(grpc.Creds(creds))

	} else {
		log.Info("starting-insecure-grpc-server")
		s.gs = grpc.NewServer()
	}

	// Register all required services
	for _, service := range s.services {
		service(s.gs)
	}

	if err := s.gs.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v\n", err)
	}
}

/*
Stop servicing GRPC requests
*/
func (s *GrpcServer) Stop() {
	s.gs.Stop()
}

/*
AddService appends a generic service request function
*/
func (s *GrpcServer) AddService(
	registerFunction func(*grpc.Server),
) {
	s.services = append(s.services, registerFunction)
}