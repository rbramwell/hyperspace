package main

import (
	"encoding/json"
	"log"
)

// A Client is a connected player and associated websocket connection.
type Client struct {
	// WebSocket connection (communicate with this via send and receive channels)
	conn *Connection

	// ID of the player that this client represents
	playerId string
}

func makeClient(conn *Connection) *Client {
	c := &Client{conn: conn}
	game.register <- c
	return c
}

func (c *Client) Initialize(playerId string, gameState *UpdateData) {
	c.playerId = playerId

	// send initial player data to client
	b, err := json.Marshal(&InitData{playerId, gameState})
	if err != nil {
		panic(err)
	}
	raw := json.RawMessage(b)
	c.Send(&Message{"init", &raw})

	// boot client message handler
	go c.run()
}

func (c *Client) run() {
	defer func() {
		game.unregister <- c
		close(c.conn.send)
	}()

	for {
		select {
		case message, ok := <-c.conn.receive:
			if !ok {
				log.Println("Client stopping", c.playerId)
				return
			}
			c.handleMessage(message)
		}
	}
}

func (c *Client) handleMessage(message *Message) {
	switch message.Type {
	case "position":
		var data PositionData
		err := json.Unmarshal([]byte(*message.Data), &data)
		if err != nil {
			log.Fatal(err)
		}

		game.events <- &Event{"position", c.playerId, &data}
	case "fire":
		var data FireData
		err := json.Unmarshal([]byte(*message.Data), &data)
		if err != nil {
			log.Fatal(err)
		}

		game.events <- &Event{"fire", c.playerId, &data}
	}

}

func (c *Client) Send(message *Message) {
	c.conn.send <- message
}