package main

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

func main() {
	message := "Hello Nym!"

	uri := "ws://localhost:1977"

	conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	rec := "J24dDRezY2BULGEAK9zKWfDxkywvg5EyPYmVEmUJVFNH.3CHWoonozybwVUopVTwkWvobaVD5WfQdfiNoyjew9aBw@6LdVTJhRfJKsrUtnjFqE3TpEbCYs3VZoxmaoNFqRWn4x"
	sendRequest, err := json.Marshal(map[string]interface{}{
		"type":          "send",
		"recipient":     rec,
		"message":       message,
		"withReplySurb": true,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("sending '%v' (*with* reply SURB) over the mix network...\n", message)
	if err = conn.WriteMessage(websocket.TextMessage, []byte(sendRequest)); err != nil {
		panic(err)
	}
}
