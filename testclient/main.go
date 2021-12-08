package main

import (
	"encoding/json"
	nL "eternity/nymLib"
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

	selfAddress := nL.GetSelfAddress(conn)
	fmt.Printf("our address is: %v\n", selfAddress)
	sendRequest, err := json.Marshal(map[string]interface{}{
		"type":          "send",
		"recipient":     selfAddress,
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
