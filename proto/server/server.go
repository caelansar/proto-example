package main

import (
	"context"
	"log"
	"net"
	"os"
	"sync"

	"proto-example/proto/consul"
	"proto-example/proto/interceptor"

	"github.com/gogo/protobuf/jsonpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
)

func init() {
	encoding.RegisterCodec(JSON{
		Marshaler: jsonpb.Marshaler{
			EmitDefaults: true,
			OrigName:     true,
		},
	})
}

type ServerConfig struct {
	Network string
	Addr    string
}

type Server struct {
	conf         *ServerConfig
	mutex        sync.RWMutex
	server       *grpc.Server
	interceptors []grpc.UnaryServerInterceptor
}

func NewServer(conf *ServerConfig, opt ...grpc.ServerOption) *Server {
	s := new(Server)
	if err := s.SetConfig(conf); err != nil {
		panic("set config failed, err: " + err.Error())
	}
	opt = append(opt, grpc.UnaryInterceptor(s.ChainInterceptor))
	s.server = grpc.NewServer(opt...)
	s.Use(interceptor.Recovery(), interceptor.ServerLogger())
	return s
}

func (s *Server) SetConfig(conf *ServerConfig) error {
	if conf == nil {
		conf = new(ServerConfig)
	}
	if len(conf.Network) == 0 {
		conf.Network = "tcp"
	}
	if len(conf.Addr) == 0 {
		conf.Addr = ":8080"
	}
	s.mutex.Lock()
	s.conf = conf
	s.mutex.Unlock()
	return nil
}

// ChainInterceptor is a single interceptor out of a chain of many interceptors.
func (s *Server) ChainInterceptor(ctx context.Context, req interface{}, args *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	var (
		i     int
		chain grpc.UnaryHandler
	)

	n := len(s.interceptors)
	if n == 0 {
		return handler(ctx, req)
	}

	chain = func(ic context.Context, ir interface{}) (interface{}, error) {
		if i == n-1 {
			return handler(ic, ir)
		}
		i++
		return s.interceptors[i](ic, ir, args, chain)
	}

	return s.interceptors[0](ctx, req, args, chain)
}

func (s *Server) Use(interceptors ...grpc.UnaryServerInterceptor) *Server {
	finalSize := len(s.interceptors) + len(interceptors)
	if finalSize >= 16 {
		panic("server use too many interceptor")
	}
	mergedInterceptors := make([]grpc.UnaryServerInterceptor, finalSize)
	copy(mergedInterceptors, s.interceptors)
	copy(mergedInterceptors[len(s.interceptors):], interceptors)
	s.interceptors = mergedInterceptors
	return s
}

func (s *Server) Start() error {
	log.Printf("start grpc server on: %s\n", s.conf.Addr)
	lis, err := net.Listen(s.conf.Network, s.conf.Addr)
	if err != nil {
		panic(err)
	}
	s.registerToConsul()
	return s.Serve(lis)
}

func (s *Server) Serve(lis net.Listener) error {
	return s.server.Serve(lis)
}

func (s *Server) GrpcServer() *grpc.Server {
	return s.server
}

func (s *Server) registerToConsul() {
	log.Println("register to consul")
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	err = consul.RegisterService("127.0.0.1:8500", &consul.ConsulService{
		Name: "hello_server",
		Tag:  []string{"hello_server"},
		Ip:   hostname,
		Port: 8080,
	})
	if err != nil {
		panic(err)
	}
}
