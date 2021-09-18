package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/textproto"
	"sync"

	"github.com/google/uuid"
)

type ChatServer struct {
	broadcastMsgChan      chan Message
	broadcastShutdownChan chan interface{}
	connChan              chan net.Conn
	connErrChan           chan error
	exitWord              string
	msgChannels           sync.Map
	shutdownChannels      sync.Map
}

func NewChatServer(exitWord string) *ChatServer {
	return &ChatServer{
		broadcastMsgChan:      make(chan Message),
		broadcastShutdownChan: make(chan interface{}),
		connChan:              make(chan net.Conn),
		connErrChan:           make(chan error),
		exitWord:              exitWord,
	}
}

func (c *ChatServer) Start(l net.Listener) error {
	qChan := make(chan interface{}, 1)
	c.shutdownChannels.Store(uuid.New().String(), qChan)
	log.Println("Server starting...")
	go c.broadcastMsg()
	go c.listenForConnections(l)

	for {
		select {
		case conn := <-c.connChan:
			go c.handleConn(conn)
		case <-qChan:
			log.Println("Server shutting down")
			return nil
		case err := <-c.connErrChan:
			log.Println("Server failure: ", err)
			c.Shutdown()
			return err
		}
	}
}

func (c *ChatServer) broadcastMsg() {
	qChan := make(chan interface{}, 1)
	c.shutdownChannels.Store(uuid.New().String(), qChan)
out:
	for {
		select {
		case msg := <-c.broadcastMsgChan:
			c.msgChannels.Range(func(key, value interface{}) bool {
				if outChan, ok := value.(chan Message); ok {
					outChan <- msg
				}
				return true
			})
		case <-qChan:
			break out
		}
	}
}

func (c *ChatServer) handleConn(conn net.Conn) {
	var name string
	id := uuid.New().String()

	tpReader := textproto.NewReader(bufio.NewReader(conn))
	tpWriter := textproto.NewWriter(bufio.NewWriter(conn))

	tpWriter.PrintfLine("Digite seu nome: ")
	line, err := tpReader.ReadLine()
	if err != nil {
		log.Println("Error reading name")
		name = id
	} else {
		name = line
	}
	tpWriter.PrintfLine("Digite %s para sair", c.exitWord)

	msgChan := make(chan Message, 1)
	c.msgChannels.Store(id, msgChan)

	c.broadcastMsgChan <- Message{
		content:  fmt.Sprintf("%s entrou na sala.", name),
		userID:   "",
		userName: "",
	}

	go c.writeMessage(id, conn, msgChan)

	for {
		line, err = tpReader.ReadLine()
		if err != nil {
			if err != io.EOF {
				log.Println("Error reading line: ", err)
			}
			break
		}

		if line == c.exitWord {
			conn.Close()
			break
		}

		c.broadcastMsgChan <- Message{
			content:  line,
			userID:   id,
			userName: name,
		}
	}
	c.broadcastMsgChan <- Message{
		content: fmt.Sprintf("%s saiu da sala.", name),
		userID:  id,
	}
}

func (c *ChatServer) listenForConnections(l net.Listener) {
	qChan := make(chan interface{}, 1)
	id := uuid.New().String()
	c.shutdownChannels.Store(id, qChan)
	for {
		select {
		case <-qChan:
			return
		default:
			conn, err := l.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					c.shutdownChannels.Delete(id)
					c.connErrChan <- err
					return
				}
				log.Println("Error acceptinng conn: ", err)
				continue
			}
			c.connChan <- conn
		}
	}
}

func (c *ChatServer) Shutdown() {
	c.shutdownChannels.Range(func(key, value interface{}) bool {
		if qChan, ok := value.(chan interface{}); ok {
			qChan <- nil
			c.shutdownChannels.Delete(key)
			close(qChan)
		}
		return true
	})
}

func (c *ChatServer) writeMessage(ignoreID string, w io.Writer, msgChan <-chan Message) {
	qChan := make(chan interface{}, 1)
	id := uuid.New().String()
	c.shutdownChannels.Store(id, qChan)
	tpWriter := textproto.NewWriter(bufio.NewWriter(w))

out:
	for {
		select {
		case msg := <-msgChan:
			if msg.userID != ignoreID {
				var prefix string
				if msg.userName != "" {
					prefix = msg.userName + ": "
				}
				tpWriter.PrintfLine("%s%s", prefix, msg.content)
			}
		case <-qChan:
			break out
		}
	}
}
