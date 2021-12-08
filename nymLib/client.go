package nymLib

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/gorilla/websocket"
)

const sendRequestTag = 0x00
const replyRequestTag = 0x01
const selfAddressRequestTag = 0x02

// response tags
const errorResponseTag = 0x00
const receivedResponseTag = 0x01
const selfAddressResponseTag = 0x02

func MakeSelfAddressRequest() []byte {
	return []byte{selfAddressRequestTag}
}

func parseSelfAddressResponse(rawResponse []byte) []byte {
	if len(rawResponse) != 97 || rawResponse[0] != selfAddressResponseTag {
		panic("Received invalid response")
	}
	return rawResponse[1:]
}

func makeSendRequest(recipient []byte, message []byte, withReplySurb bool) []byte {
	messageLen := make([]byte, 8)
	binary.BigEndian.PutUint64(messageLen, uint64(len(message)))

	surbByte := byte(0)
	if withReplySurb {
		surbByte = 1
	}

	out := []byte{sendRequestTag, surbByte}
	fmt.Printf("request tag and surb byte is %d bytes long\n", len(out))
	fmt.Printf("the recipient is %d bytes long\n", len(recipient))
	out = append(out, recipient...)
	out = append(out, messageLen...)
	out = append(out, message...)

	return out
}

func MakeReplyRequest(message []byte, replySURB []byte) []byte {
	messageLen := make([]byte, 8)
	binary.BigEndian.PutUint64(messageLen, uint64(len(message)))

	surbLen := make([]byte, 8)
	binary.BigEndian.PutUint64(surbLen, uint64(len(replySURB)))

	out := []byte{replyRequestTag}
	out = append(out, surbLen...)
	out = append(out, replySURB...)
	out = append(out, messageLen...)
	out = append(out, message...)

	return out
}

func ParseReceived(rawResponse []byte) ([]byte, []byte) {
	if rawResponse[0] != receivedResponseTag {
		panic("Received invalid response!")
	}

	hasSurb := false
	if rawResponse[1] == 1 {
		hasSurb = true
	} else if rawResponse[1] == 0 {
		hasSurb = false
	} else {
		panic("malformed received response!")
	}

	data := rawResponse[2:]
	if hasSurb {
		surbLen := binary.BigEndian.Uint64(data[:8])
		other := data[8:]

		surb := other[:surbLen]
		msgLen := binary.BigEndian.Uint64(other[surbLen : surbLen+8])

		if len(other[surbLen+8:]) != int(msgLen) {
			panic("invalid msg len")
		}

		msg := other[surbLen+8:]
		return msg, surb
	} else {
		msgLen := binary.BigEndian.Uint64(data[:8])
		other := data[8:]

		if len(other) != int(msgLen) {
			panic("invalid msg len")
		}

		msg := other[:msgLen]
		return msg, nil
	}
}

func SendBinaryWithoutReply() {
	uri := "ws://localhost:1977"

	conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	selfAddressRequest := MakeSelfAddressRequest()
	if err = conn.WriteMessage(websocket.BinaryMessage, selfAddressRequest); err != nil {
		panic(err)
	}
	_, receivedResponse, err := conn.ReadMessage()
	if err != nil {
		panic(err)
	}
	selfAddress := parseSelfAddressResponse(receivedResponse)

	readData, err := ioutil.ReadFile("dummy_file")
	if err != nil {
		panic(err)
	}

	sendRequest := makeSendRequest(selfAddress, readData, false)
	fmt.Printf("sending content of 'dummy file' over the mix network...\n")
	if err = conn.WriteMessage(websocket.BinaryMessage, sendRequest); err != nil {
		panic(err)
	}

	fmt.Printf("waiting to receive a message from the mix network...\n")
	_, receivedResponse, err = conn.ReadMessage()
	if err != nil {
		panic(err)
	}

	fileData, replySURB := ParseReceived(receivedResponse)
	if replySURB != nil {
		panic("did not expect a replySURB!")
	}
	fmt.Printf("writing the file back to the disk!\n")
	ioutil.WriteFile("received_file_no_reply", fileData, 0644)
}

func SendBinaryWithReply(recipient string) {
	uri := "ws://localhost:1977"

	conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	selfAddressRequest := MakeSelfAddressRequest()
	if err = conn.WriteMessage(websocket.BinaryMessage, selfAddressRequest); err != nil {
		panic(err)
	}
	_, receivedResponse, err := conn.ReadMessage()
	if err != nil {
		panic(err)
	}
	selfAddress := parseSelfAddressResponse(receivedResponse)

	readData, err := ioutil.ReadFile("file_example_PNG_2500kB.jpg")
	if err != nil {
		panic(err)
	}

	sendRequest := makeSendRequest(selfAddress, readData, true)
	fmt.Printf("sending content of 'file_example_PNG_2500KB.jpg' over the mix network...\n")
	if err = conn.WriteMessage(websocket.BinaryMessage, sendRequest); err != nil {
		panic(err)
	}

	fmt.Printf("waiting to receive a message from the mix network...\n")
	_, receivedResponse, err = conn.ReadMessage()
	if err != nil {
		panic(err)
	}

	fileData, replySURB := ParseReceived(receivedResponse)

	fmt.Printf("writing the file back to the disk!\n")
	ioutil.WriteFile("received_file_withreply", fileData, 0644)

	replyMessage := []byte("hello from reply SURB! - thanks for sending me the file!")
	replyRequest := MakeReplyRequest(replyMessage, replySURB)

	fmt.Printf("sending '%v' (using reply SURB) over the mix network...\n", string(replyMessage))
	if err = conn.WriteMessage(websocket.BinaryMessage, replyRequest); err != nil {
		panic(err)
	}

	fmt.Printf("waiting to receive a message from the mix network...\n")
	_, receivedResponse, err = conn.ReadMessage()
	if err != nil {
		panic(err)
	}

	receivedMessage, replySURB := ParseReceived(receivedResponse)
	if replySURB != nil {
		panic("did not expect a replySURB!")
	}

	fmt.Printf("received %v from the mix network!\n", string(receivedMessage))

}

func GetSelfAddress(conn *websocket.Conn) string {
	selfAddressRequest, err := json.Marshal(map[string]string{"type": "selfAddress"})
	if err != nil {
		panic(err)
	}

	if err = conn.WriteMessage(websocket.TextMessage, []byte(selfAddressRequest)); err != nil {
		panic(err)
	}

	responseJSON := make(map[string]interface{})
	err = conn.ReadJSON(&responseJSON)
	if err != nil {
		panic(err)
	}

	return responseJSON["address"].(string)
}

func SendTextWithoutReply() {
	message := "Hello Nym!"

	uri := "ws://localhost:1977"

	conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	selfAddress := GetSelfAddress(conn)
	fmt.Printf("our address is: %v\n", selfAddress)
	sendRequest, err := json.Marshal(map[string]interface{}{
		"type":          "send",
		"recipient":     selfAddress,
		"message":       message,
		"withReplySurb": false,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("sending '%v' (*without* reply SURB) over the mix network...\n", message)
	if err = conn.WriteMessage(websocket.TextMessage, []byte(sendRequest)); err != nil {
		panic(err)
	}

	fmt.Printf("waiting to receive a message from the mix network...\n")
	_, receivedMessage, err := conn.ReadMessage()
	if err != nil {
		panic(err)
	}
	fmt.Printf("received %v from the mix network!\n", string(receivedMessage))
}

func SendTextWithReply() {
	message := "Hello Nym!"

	uri := "ws://localhost:1977"

	conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	selfAddress := GetSelfAddress(conn)
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

	fmt.Printf("waiting to receive a message from the mix network...\n")
	_, receivedMessage, err := conn.ReadMessage()
	if err != nil {
		panic(err)
	}
	fmt.Printf("received %v from the mix network!\n", string(receivedMessage))

	receivedMessageJSON := make(map[string]interface{})
	if err := json.Unmarshal(receivedMessage, &receivedMessageJSON); err != nil {
		panic(err)
	}

	// use the received surb to send an anonymous reply!
	replySurb := receivedMessageJSON["replySurb"]
	replyMessage := "hello from reply SURB!"

	reply, err := json.Marshal(map[string]interface{}{
		"type":      "reply",
		"message":   replyMessage,
		"replySurb": replySurb,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("sending '%v' (using reply SURB) over the mix network...\n", replyMessage)
	if err = conn.WriteMessage(websocket.TextMessage, []byte(reply)); err != nil {
		panic(err)
	}

	fmt.Printf("waiting to receive a message from the mix network...\n")
	_, receivedMessage, err = conn.ReadMessage()
	if err != nil {
		panic(err)
	}
	fmt.Printf("received %v from the mix network!\n", string(receivedMessage))
}
