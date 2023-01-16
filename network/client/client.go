package client

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	Id       uuid.UUID
	GameChan chan *[]byte
	WSConn   *websocket.Conn
}

func NewClient(id uuid.UUID, wsConn *websocket.Conn, gameChannel chan *[]byte) *Client {
	return &Client{
		Id:       id,
		WSConn:   wsConn,
		GameChan: gameChannel,
	}
}

func (c *Client) Listen() {
	go c.serverListenWrite(c.WSConn, c.GameChan)
	c.serverListenRead(c.WSConn)
}

func (c *Client) serverListenWrite(conn *websocket.Conn, ch chan *[]byte) {
	for {
		message := []byte(fmt.Sprintf("clientId: %v ", c.Id) + string(*<-ch))
		err := conn.WriteMessage(1, message)
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("game server resp: %s", message)
	}
}

func (c *Client) serverListenRead(conn *websocket.Conn) {

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("game server recv: %s", message)

		// echo for now, but this should send inputs somewhere to be calculated and forwarded to all clients
		c.GameChan <- &message
	}
}
