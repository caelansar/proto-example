package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"proto-example/proto/consul"
	pb "proto-example/proto/testproto"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type helloServer struct {
	addr string
}

func init() {
	encoding.RegisterCodec(JSON{
		Marshaler: jsonpb.Marshaler{
			EmitDefaults: true,
			OrigName:     true,
		},
	})
}

func (s *helloServer) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Println("call SayHello")
	return &pb.HelloReply{Message: fmt.Sprintf("hello %s from %s", in.Name, s.addr), Success: true}, nil
}

func (s *helloServer) StreamHello(ss pb.Greeter_StreamHelloServer) error {
	log.Println("call StreamHello")
	for i := 0; i < 3; i++ {
		in, err := ss.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		ret := &pb.HelloReply{Message: "Hello " + in.Name, Success: true}
		err = ss.Send(ret)
		if err != nil {
			return err
		}
	}
	return nil
}

// HealthImpl is impl of grpc_health_v1.HealthServer interface
type HealthImpl struct {
}

func (h *HealthImpl) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	log.Println("hello server health check")
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (h *HealthImpl) Watch(req *grpc_health_v1.HealthCheckRequest, w grpc_health_v1.Health_WatchServer) error {
	return nil
}

func registerToConsul() {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	err = consul.RegisterService("127.0.0.1:8500", &consul.ConsulService{
		Name: "hello_server",
		Tag:  []string{"hello_server"},
		Ip:   hostname,
		Port: 8081,
	})
	if err != nil {
		panic(err)
	}
}

func runServer(addr string) {
	log.Printf("start grpc server on: %s\n", addr)
	server := grpc.NewServer()
	pb.RegisterGreeterServer(server, &helloServer{addr: addr})
	grpc_health_v1.RegisterHealthServer(server, &HealthImpl{})
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	registerToConsul()
	err = server.Serve(lis)
	if err != nil {
		panic(err)
	}
}

func main() {
	runServer(":8081")
}

// JSON is impl of encoding.Codec
type JSON struct {
	jsonpb.Marshaler
	jsonpb.Unmarshaler
}

func (j JSON) Name() string {
	return "json"
}

func (j JSON) Marshal(v interface{}) (out []byte, err error) {
	if pm, ok := v.(proto.Message); ok {
		b := new(bytes.Buffer)
		err := j.Marshaler.Marshal(b, pm)
		if err != nil {
			return nil, err
		}
		return b.Bytes(), nil
	}
	return json.Marshal(v)
}

func (j JSON) Unmarshal(data []byte, v interface{}) (err error) {
	if pm, ok := v.(proto.Message); ok {
		b := bytes.NewBuffer(data)
		return j.Unmarshaler.Unmarshal(b, pm)
	}
	return json.Unmarshal(data, v)
}
