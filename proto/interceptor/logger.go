package interceptor

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

func ServerLogger() grpc.UnaryServerInterceptor {
	log.Println("register server logger interceptor")
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		var remoteIP string
		if peerInfo, ok := peer.FromContext(ctx); ok {
			remoteIP = peerInfo.Addr.String()
		}
		// call server handler
		resp, err := handler(ctx, req)
		log.Println("elapse time:", time.Since(startTime))
		log.Println("remote ip: ", remoteIP)
		log.Println("req: ", req)
		return resp, err
	}
}
