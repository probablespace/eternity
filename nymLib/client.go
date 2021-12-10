package nymLib

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

const sendRequestTag = 0x00
const replyRequestTag = 0x01
const selfAddressRequestTag = 0x02

// response tags
const errorResponseTag = 0x00
const receivedResponseTag = 0x01
const selfAddressResponseTag = 0x02

type InvalidRequestError struct{}

func (m *InvalidRequestError) Error() string {
	return "malformed or invalid request"
}

func MakeSelfAddressRequest() []byte {
	return []byte{selfAddressRequestTag}
}

func parseSelfAddressResponse(rawResponse []byte) []byte {
	if len(rawResponse) != 97 || rawResponse[0] != selfAddressResponseTag {
		panic("Received invalid response")
	}
	return rawResponse[1:]
}

func MakeSendRequest(recipient []byte, message []byte, withReplySurb bool) []byte {
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

func ParseReceived(rawResponse []byte) (ServerRequest, error) {
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
		actionByte := msg[0]
		SR := &ServerRequest{
			SURB: surb,
		}
		msg = msg[1:]
		switch actionByte {
		case 0x00: // search
		case 0x01: // store
			// pubByte := msg[0] // if the pubByte = 0, this is a public file
			msg = msg[1:]
			publicKey := msg[:32] // 32 byte ED25519 public key
			fileSig := msg[32:96] // 64 byte ED25519 signature of file
			fileBody := msg[96:]  // file body

			SR.Action = actionByte
			SR.FileSig = fileSig
			SR.PubKey = publicKey
			SR.Body = fileBody
		case 0x02: // serve
		case 0x03: // delete
		}
		var clientRequest RawRequest
		err := json.Unmarshal(msg, clientRequest)
		if err != nil {

		}

		return *SR, nil
	} else {
		return ServerRequest{}, &InvalidRequestError{}
	}
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
