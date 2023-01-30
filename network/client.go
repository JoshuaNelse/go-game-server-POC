package network

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
	go c.serverListenWrite()
	c.serverListenRead()
}

func (c *Client) serverListenWrite() {
	for {
		message := []byte(fmt.Sprintf("clientId: %v ", c.Id) + string(*<-c.GameChan))
		err := c.WSConn.WriteMessage(1, message)
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("game server resp: %s", message)
	}
}

func (c *Client) serverListenRead() {

	for {
		_, message, err := c.WSConn.ReadMessage()
		if err != nil {
			log.Println("read err:", err)
			break
		}
		log.Printf("game server recv: %s", message)

		// echo for now, but this should send inputs somewhere to be calculated and result forwarded to all clients
		c.GameChan <- &message
	}
}
