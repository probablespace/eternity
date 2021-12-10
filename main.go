package main

import (
	nL "eternity/nymLib"

	"github.com/gorilla/websocket"
)

// fServe "eternity/eternityFS"

// "github.com/gorilla/websocket"

func main() {
	// nL.StartEternityServerNymClientWatcher()
	// nL.SendBinaryWithReply()

	// could put some gateway bs here

	uri := "ws://localhost:1977"

	conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
	if err != nil {
		panic(err)
	}
	wsh := nL.NewWebsocketHandler(conn)

	go wsh.RequestProcessor()
	println("starting reader routine")
	for {
		_, receivedResponse, err := wsh.Conn.ReadMessage()
		if err != nil {
			panic(err)
		}

		request, err := nL.ParseReceived(receivedResponse)
		if err == nil {
			wsh.RequestQueue <- request
		}
	}
}
