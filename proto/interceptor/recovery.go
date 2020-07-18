package interceptor

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"

	"google.golang.org/grpc"
)

// recovery is a server interceptor that recovers from any panics.
func Recovery() grpc.UnaryServerInterceptor {
	log.Println("register recovery interceptor")
	return func(ctx context.Context, req interface{}, args *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if rerr := recover(); rerr != nil {
				const size = 64 << 10
				buf := make([]byte, size)
				rs := runtime.Stack(buf, false)
				if rs > size {
					rs = size
				}
				buf = buf[:rs]
				pl := fmt.Sprintf("grpc server panic: %v\n%v\n%s\n", req, rerr, buf)
				fmt.Fprintf(os.Stderr, pl)
				log.Println(pl)
			}
		}()
		resp, err = handler(ctx, req)
		return
	}
}
