// Full credits to Ivan Velichko
// Run docker build -t http_server .
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func main() {
	lc := net.ListenConfig{
		Control: func(network, address string, conn syscall.RawConn) error {
			var ret error
			if err := conn.Control(func(fd uintptr) {
				ret = syscall.SetsockoptInt(
					int(fd),
					unix.SOL_SOCKET,
					unix.SO_REUSEPORT,
					1,
				)
			}); err != nil {
				return err
			}
			return ret
		},
	}

	ln, err := lc.Listen(
		context.Background(),
		"tcp",
		os.Getenv("HOST")+":"+os.Getenv("PORT"),
	)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, _req *http.Request) {
		w.Write([]byte(fmt.Sprintf("Hello from %s\n", os.Getenv("INSTANCE"))))
	})

	if err := http.Serve(ln, nil); err != nil {
		panic(err)
	}
}
