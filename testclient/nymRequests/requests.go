package nymRequests

/*****************

A Request to the server is in one of the following formats:

# File Search
1 byte   	: 	Request/Response tag (0x00, 0x01, or 0x02)
1 byte   	: 	SURB byte (we require these, ie must equal 1)
8 bytes  	: 	SURB Length (SL) telling you how long the SURB is
SL bytes 	: 	Singlue Use Reply Block
// The request body starts here
32 bytes 	: 	SHA256 Hash of file

# File Download
1 byte   	: 	Request/Response tag (0x00, 0x01, or 0x02)
1 byte   	: 	SURB byte (we require these, ie must equal 1)
8 bytes  	: 	SURB Length (SL) telling you how long the SURB is
SL bytes 	: 	Single Use Reply Block
// The request body starts here
32 bytes : SHA256 Hash of file

# File Upload
1 byte   	: 	Request/Response tag (0x00, 0x01, or 0x02)
1 byte   	: 	SURB byte (we require these, ie must equal 1)
8 bytes  	: 	SURB Length (SL) telling you how long the SURB is
SL bytes 	: 	Single Use Reply Block
// The request body starts here
32 bytes 	: 	ED25519 Public Key to validate message
64 bytes 	: 	ED25519 Signature of message
[96:] bytes : 	the file

# File Delete
1 byte   	: 	Request/Response tag (0x00, 0x01, or 0x02)
1 byte   	: 	SURB byte (we require these, ie must equal 1)
8 bytes  	: 	SURB Length (SL) telling you how long the SURB is
SL bytes 	: 	Singlue Use Reply Block
// The request body starts here
32 bytes 	: 	SHA256 file hash
64 bytes 	: 	ED25519 Signature of hash of file to be validated
				against saved public key
 bytes : the actual message



*****************/

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"io/ioutil"

	"github.com/gorilla/websocket"
)

// request tags
const sendRequestTag = 0x00
const replyRequestTag = 0x01
const selfAddressRequestTag = 0x02

// response tags
const errorResponseTag = 0x00
const receivedResponseTag = 0x01
const selfAddressResponseTag = 0x02

type ClientVars struct {
	ServerAddress string
	Pubkey        ed25519.PrivateKey
	Privkey       ed25519.PrivateKey
	AESkey        []byte
}

func makeSelfAddressRequest() []byte {
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

	vanilla := make([]byte, 97)

	out := []byte{sendRequestTag, surbByte}
	out = append(out, recipient...)
	out = append(out, messageLen...)
	out = append(out, vanilla...)
	out = append(out, message...)

	return out
}

func makeReplyRequest(message []byte, replySURB []byte) []byte {
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

func parseReceived(rawResponse []byte) ([]byte, []byte) {
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

func (cV ClientVars) SendBinaryWithReply(file []byte) {
	uri := "ws://localhost:1977"

	conn, _, err := websocket.DefaultDialer.Dial(uri, nil)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	sendRequest := makeSendRequest([]byte(cV.ServerAddress), file, true)
	fmt.Printf("sending content of 'dummy file' over the mix network...\n")
	if err = conn.WriteMessage(websocket.BinaryMessage, sendRequest); err != nil {
		panic(err)
	}

	fmt.Printf("waiting to receive a message from the mix network...\n")
	_, receivedResponse, err := conn.ReadMessage()
	if err != nil {
		panic(err)
	}

	fileData, replySURB := parseReceived(receivedResponse)

	fmt.Printf("writing the file back to the disk!\n")
	ioutil.WriteFile("received_file_withreply", fileData, 0644)

	replyMessage := []byte("hello from reply SURB! - thanks for sending me the file!")
	replyRequest := makeReplyRequest(replyMessage, replySURB)

	fmt.Printf("sending '%v' (using reply SURB) over the mix network...\n", string(replyMessage))
	if err = conn.WriteMessage(websocket.BinaryMessage, replyRequest); err != nil {
		panic(err)
	}

	fmt.Printf("waiting to receive a message from the mix network...\n")
	_, receivedResponse, err = conn.ReadMessage()
	if err != nil {
		panic(err)
	}

	receivedMessage, replySURB := parseReceived(receivedResponse)
	if replySURB != nil {
		panic("did not expect a replySURB!")
	}

	fmt.Printf("received %v from the mix network!\n", string(receivedMessage))

}
