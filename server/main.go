package main

import (
	"bufio"
	"fmt"
	"log"
	"os/signal"
	"syscall"

	"github.com/lestrrat/go-server-starter/listener"

	// "net"
	"io"
)

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)

	listener, err := listener.ListenAll()
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err.Error())
			continue
		}
		defer conn.Close()
		go func() {
			r := bufio.NewReader(conn)
			for {
				line, err := r.ReadString('\n')
				if err == io.EOF {
					break
				} else if err != nil {
					log.Println(err.Error())
				}
				fmt.Printf(line)
			}
		}()
	}
}
