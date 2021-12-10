package nymLib

import (
	"encoding/binary"
	"eternity/eternityFS"
	"sync"

	"github.com/gorilla/websocket"
)

// const sendRequestTag = 0x00
// const replyRequestTag = 0x01
// const selfAddressRequestTag = 0x02

// // response tags
// const errorResponseTag = 0x00
// const receivedResponseTag = 0x01
// const selfAddressResponseTag = 0x02

type RawRequest struct {
	Action  []byte `json:"action"`
	Payload []byte `json:"payload"`
}

type ServerRequest struct {
	SURB    []byte
	Action  byte
	FileSig []byte
	PubKey  []byte
	Body    []byte
}

type ServerResponse struct {
	SURB    []byte
	Message []byte
}

type WebSocketHandler struct {
	writeMut      sync.Mutex // mutex for writing to the connection
	Conn          *websocket.Conn
	RequestQueue  chan ServerRequest
	ResponseQueue chan ServerResponse
	Efs           eternityFS.EternityFS
}

func saveFile(message []byte) {}

func sendFile(fileData []byte, recipient []byte) {}

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

func NewWebsocketHandler(conn *websocket.Conn) *WebSocketHandler {
	efs, err := eternityFS.InitEFS("/Users/orchid/eternity")

	if err != nil {
		panic(err)
	}

	wsh := &WebSocketHandler{
		Conn:          conn,
		Efs:           efs,
		RequestQueue:  make(chan ServerRequest, 50),
		ResponseQueue: make(chan ServerResponse, 50),
	}
	return wsh
}
func (wsh *WebSocketHandler) ReaderRoutine() {
}

func (wsh *WebSocketHandler) RequestProcessor() {
	println("starting request processor")
	for {
		for request := range wsh.RequestQueue {
			go wsh.HandleRequest(request)
		}
	}
}

func (wsh *WebSocketHandler) HandleRequest(sR ServerRequest) {
	switch sR.Action {
	case 0x00: // search
		hash := string(sR.Body)
		response := &ServerResponse{
			SURB: sR.SURB,
		}
		if wsh.Efs.Search(hash) {
			response.Message = []byte("file found")
		} else {
			response.Message = []byte("file not found")
		}
		wsh.ResponseQueue <- *response
	case 0x01: // store
		wsh.Efs.Store(sR.Body, sR.PubKey, sR.FileSig)
	case 0x02: // serve
		hash := string(sR.Body)
		response := &ServerResponse{
			SURB: sR.SURB,
		}
		file, err := wsh.Efs.GetFile(hash)
		out := make([]byte, 0)
		if err != nil {
			out = append(out, 0x00)
		} else {
			out = append(out, 0x01)
			out = append(out, file...)
			response.Message = out
		}

		wsh.ResponseQueue <- *response
	case 0x03: // delete
	}
}

func (wsh *WebSocketHandler) ResponseProcessor() {
	for {
		for response := range wsh.ResponseQueue {
			wsh.SendResponse(response.SURB, response.Message)
		}
	}
}

func (wsh *WebSocketHandler) SendResponse(message []byte, replySURB []byte) {
	messageLen := make([]byte, 8)
	binary.BigEndian.PutUint64(messageLen, uint64(len(message)))

	surbLen := make([]byte, 8)
	binary.BigEndian.PutUint64(surbLen, uint64(len(replySURB)))

	out := []byte{replyRequestTag}
	out = append(out, surbLen...)
	out = append(out, replySURB...)
	out = append(out, messageLen...)
	out = append(out, message...)

	if err := wsh.Conn.WriteMessage(websocket.BinaryMessage, out); err != nil {
		panic(err)
	}

}
