package main

import (
	nL "eternity/nymLib"
)

// fServe "eternity/eternityFS"

// "github.com/gorilla/websocket"

func main() {
	// nL.StartEternityServerNymClientWatcher()
	// nL.SendBinaryWithReply()

	nL.CheckComm()
	// initialize the server
	// uri := "ws://localhost:1977"

	// conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
	// if err != nil {
	// 	panic(err)
	// }
	// defer conn.Close()
	// event loop where we poll the client for any and all requests
	// for {
	// 	_, reveivedMessage, err := conn.ReadMessage()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	message, err := nL.ParseReceived(reveivedMessage)
	// 	fmt.Printf("received the following message: \n %s", message)
	// 	switch nL.messageType(reveivedMessage) {
	// 	default:
	// 	}
	// }
	// efs, err := fServe.InitEFS("/Users/orchid/eternity")
	// if err != nil {
	// 	panic(err)
	// }

	// err = efs.IndexFiles(efs.Opts.FileDir)

	// if err != nil {
	// 	panic(err)
	// }

	// readData, err := ioutil.ReadFile("file_example_PNG_2500kB.jpg")
	// if err != nil {
	// 	panic(err)
	// }

	// hash, _ := efs.Store(readData)

	// println("stored file with hash: ", hash)

	// nL.SendBinaryWithReply()
	// nL.SendBinaryWithoutReply()
	// nL.SendTextWithReply()
	// for {
	// 	fmt.Printf("waiting to receive a message from the mix network...\n")
	// 	_, receivedMessage, err := conn.ReadMessage()
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	fmt.Printf("received %v from the mix network!\n", string(receivedMessage))
	// }
}
