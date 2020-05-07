package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/teru01/server-starter/listener"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "server pid %d\n", os.Getpid())
}

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)

	listers, err := listener.ListenAll()
	if err != nil {
		panic(err)
	}
	server := http.Server{
		Handler: http.HandlerFunc(handler),
	}
	go server.Serve(listers[0])

	<- signals
	server.Shutdown(context.Background())
}
