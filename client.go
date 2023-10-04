package main

import (
	"github.com/gorilla/websocket"
	"log"
)

type Client struct {
	conn *websocket.Conn
}

func ConnectClient(addr string) (*Client, error) {
	conn, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		return nil, err
	}

	return &Client{conn: conn}, nil
}

func (c *Client) SendText(text string) error {
	return c.conn.WriteMessage(websocket.TextMessage, []byte(text))
}

func (c *Client) Close() error {
	_ = c.conn.WriteMessage(websocket.CloseMessage, nil)
	log.Println("Closing connection")
	return c.conn.Close()
}

//func main() {
//	client, err := ConnectClient("ws://localhost:8080/chat/ws")
//	if err != nil {
//		panic(err)
//	}
//	defer func(client *Client) {
//		err := client.Close()
//		if err != nil {
//			panic(err)
//		}
//	}(client)
//
//	err = client.conn.WriteMessage(websocket.TextMessage, []byte("Hello, world!"))
//}
