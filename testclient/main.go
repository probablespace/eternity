package main

import (
	nL "eternity/nymLib"
	"fmt"
	"io/ioutil"

	"github.com/gorilla/websocket"
)

func main() {
	// message := "Hello Nym!"

	uri := "ws://localhost:1977"

	conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	rec := "J24dDRezY2BULGEAK9zKWfDxkywvg5EyPYmVEmUJVFNH.3CHWoonozybwVUopVTwkWvobaVD5WfQdfiNoyjew9aBw@6LdVTJhRfJKsrUtnjFqE3TpEbCYs3VZoxmaoNFqRWn4x"
	readData, err := ioutil.ReadFile("file_example_PNG_2500kB.jpg")
	if err != nil {
		panic(err)
	}

	sendRequest := nL.MakeSendRequest([]byte(rec), readData, true)
	fmt.Printf("sending content of 'file_example_PNG_2500KB.jpg' over the mix network...\n")
	if err = conn.WriteMessage(websocket.BinaryMessage, sendRequest); err != nil {
		panic(err)
	}
}
