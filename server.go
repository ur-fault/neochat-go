package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Message struct {
	Author string
	Body   string
}

type Connection struct {
	conn   *websocket.Conn
	stop   chan struct{}
	server *Server
}

func (c *Connection) Logger() *log.Logger {
	return c.server.Logger()
}

func (c *Connection) Handle() {
	defer func(conn *websocket.Conn) {
		_ = conn.Close()
	}(c.conn)
	for {
		mt, msg, err := c.conn.ReadMessage()
		if err != nil {
			c.Logger().Printf("conn: error reading message: %v", err)
			return
		}

		if mt == websocket.CloseMessage {
			return
		}

		c.Logger().Printf("conn: received %s of %d", msg, mt)
	}
}

type Server struct {
	port     uint
	addr     string
	inner    *http.Server
	messages chan Message
	logger   *log.Logger
}

func NewServer(port uint, addr string) (*Server, error) {
	if port == 0 {
		port = 8080
	}
	if addr == "" {
		return nil, errors.New("addr must be non-empty")
	}

	inner := &http.Server{Addr: fmt.Sprintf(":%d", port)}
	server := &Server{
		port:     port,
		inner:    inner,
		addr:     addr,
		messages: make(chan Message),
		logger:   log.New(os.Stdout, "server: ", log.LstdFlags),
	}

	return server, nil
}

func (s *Server) Logger() *log.Logger {
	return s.logger
}

func (s *Server) HandleConnection(conn *websocket.Conn) Connection {
	stop := make(chan struct{})
	c := Connection{
		conn: conn,
		stop: stop,
	}
	go c.Handle()

	return c
}

func (s *Server) Stop() error {
	err := s.inner.Shutdown(context.TODO())
	close(s.messages)
	s.Logger().Println("Server stopped")
	return err
}

func (s *Server) Run() {
	http.HandleFunc(s.addr, func(writer http.ResponseWriter, request *http.Request) {
		log.Println("Got connection")
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(writer, request, nil)
		s.HandleConnection(conn)
		if err != nil {
			log.Fatalf("Failed to upgrade connection: %v", err)
		}
	})

	log.Printf("Listening on port %d", s.port)
	_ = s.inner.ListenAndServe()
}

func main() {
	server, err := NewServer(8080, "/chat/ws")
	if err != nil {
		panic(err)
	}

	go server.Run()

	group := sync.WaitGroup{}
	group.Add(5)
	for i := 0; i < 5; i++ {
		go func(i int) {
			defer group.Done()
			client, err := ConnectClient("ws://localhost:8080/chat/ws")
			if err != nil {
				panic(err)
			}
			log.Println("Connected")

			err = client.SendText(fmt.Sprintf("Hello, world! from %d", i))
			if err != nil {
				panic(err)
			}
			client.Close()
		}(i)
	}

	time.Sleep(1 * time.Second)

	server.Stop()
}
