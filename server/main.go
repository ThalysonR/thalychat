package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

const (
	HOST        = "0.0.0.0"
	PORT        = "7000"
	EXIT_STRING = ":q"
)

func main() {
	runtime.GOMAXPROCS(2)
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	l, err := net.Listen("tcp", HOST+":"+PORT)
	if err != nil {
		log.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	chatServer := NewChatServer(EXIT_STRING)

	go chatServer.Start(l)
	defer chatServer.Shutdown()
	<-done
}
