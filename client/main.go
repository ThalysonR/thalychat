package main

import (
	"bufio"
	"log"
	"net"
	"net/textproto"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	conn, err := net.Dial("tcp", "cc460c08c1df.sn.mynetname.net:7000")
	if err != nil {
		panic("Erro ao conectar.")
	}
	done := make(chan os.Signal, 1)
	readDone := make(chan interface{}, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	connReader := textproto.NewReader(bufio.NewReader(conn))
	connWriter := textproto.NewWriter(bufio.NewWriter(conn))
	stdoutWriter := textproto.NewWriter(bufio.NewWriter(os.Stdout))
	stdinReader := textproto.NewReader(bufio.NewReader(os.Stdin))

	go func() {
	out:
		for {
			select {
			case <-readDone:
				break out
			default:
				line, err := connReader.ReadLine()
				log.Println("chegou ", line)
				if err != nil {
					stdoutWriter.PrintfLine("Erro ao ler mensagem.")
				} else {
					stdoutWriter.PrintfLine(line)
				}
			}
		}
	}()

	for {
		select {
		case <-done:
			os.Exit(0)
		default:
			line, err := stdinReader.ReadLine()
			if err == nil {
				log.Println("Escrevendo")
				connWriter.PrintfLine(line)
				log.Println("Escrevendo")
			}
		}
	}
}
