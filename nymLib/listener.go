package nymLib

import (
	"fmt"
	"io/ioutil"

	"github.com/gorilla/websocket"
)

// const sendRequestTag = 0x00
// const replyRequestTag = 0x01
// const selfAddressRequestTag = 0x02

// // response tags
// const errorResponseTag = 0x00
// const receivedResponseTag = 0x01
// const selfAddressResponseTag = 0x02

type nymMessage struct {
}

func CheckComm() {
	// check if we can marshal the reponse
	uri := "ws://localhost:1977"

	conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	for {
		_, receivedResponse, err := conn.ReadMessage()
		if err != nil {
			panic(err)
		}

		fileData, _ := ParseReceived(receivedResponse)

		fmt.Printf("writing the file back to the disk!\n")
		ioutil.WriteFile("received_file_withreply", fileData, 0644)
	}

}

func saveFile(message []byte) {}

func sendFile(fileData []byte, recipient []byte) {}

// func ListenForMessage() {
// 	// initialize websocket
// 	uri := "ws://localhost:1977"

// 	conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer conn.Close()

// 	// event loop where we poll the client for any and all requests
// 	for {
// 		_, receivedResponse, err := conn.ReadMessage()
// 		if err != nil {
// 			panic(err)
// 		}

// 		fileData, replySURB := ParseReceived(receivedResponse)
// 		fmt.Printf("received the following message: \n %s", message)
// 		// switch nL.messageType(reveivedMessage) {
// 		// default:
// 		// }
// 	}
// }
