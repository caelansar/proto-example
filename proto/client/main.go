package main

import (
	"context"
	"flag"
	"log"
	"proto-example/proto/consul"

	"github.com/gogo/protobuf/jsonpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/resolver"
)

type Reply struct {
	res []byte
}

var (
	data   string
	method string
	addr   string
)

func init() {
	encoding.RegisterCodec(JSON{
		Marshaler: jsonpb.Marshaler{
			EmitDefaults: true,
			OrigName:     true,
		},
	})
	flag.StringVar(&data, "data", `{"name":"cae","age":1}`, `{"name":"cae","age":1}`)
	flag.StringVar(&method, "method", "/testproto.Greeter/SayHello", `/testproto.Greeter/SayHello`)
	flag.StringVar(&addr, "addr", "consul://127.0.0.1:8500/hello_server", `consul://127.0.0.1:8500/hello_server`)
}

func main() {
	resolver.Register(consul.NewBuilder())
	flag.Parse()
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype(JSON{}.Name())),
		grpc.WithBalancerName(roundrobin.Name),
	}
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		panic(err)
	}

	var reply Reply
	err = conn.Invoke(context.Background(), method, []byte(data), &reply)
	if err != nil {
		panic(err)
	}
	log.Println(string(reply.res))
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
	return v.([]byte), nil
}

func (j JSON) Unmarshal(data []byte, v interface{}) (err error) {
	v.(*Reply).res = data
	return nil
}
