////////////////////////////////////////////////////////////////////////////////
//	test/webSocket_client.go  -  Mar-22-2022  -  aldebap
//
//	simple webSocket_client to test pld REST APIs
////////////////////////////////////////////////////////////////////////////////

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/runtime/protoimpl"
)

type WebSocketRequest struct {
	Endpoint  string `json:"endpoint,omitempty"`
	RequestId string `json:"request_id,omitempty"`
	HasMore   bool   `json:"has_more,omitempty"`
	Payload   []byte `json:"payload,omitempty"`
}

type WebSocketResponse struct {
	RequestId string          `json:"request_id,omitempty"`
	HasMore   bool            `json:"has_more,omitempty"`
	Payload   []byte          `json:"ok,omitempty"`
	Error     *WebSocketError `json:"error,omitempty"`
}

type WebSocketError struct {
	HttpCode uint32 `json:"http_code,omitempty"`
	Error    *Error `json:"error,omitempty"`
}

type Error struct {
	Message string `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
}

func main() {
	//	connect to pld webSocket URI
	var webSocketHost = "ws://localhost:8080"
	const resyncWebSocketURI = "/api/v1/meta/websocket"

	conn, _, err := websocket.DefaultDialer.Dial(webSocketHost+resyncWebSocketURI, nil)
	if err != nil {
		log.Fatal("Error connecting to Websocket Server:", err)
	}
	defer conn.Close()

	//	send a DebugLevel command
	err = sendDebugLevelCommand(conn)
	if err != nil {
		log.Println(err.Error())
		return
	}

	//	send a GetInfo command
	err = sendGetInfoCommand(conn)
	if err != nil {
		log.Println(err.Error())
		return
	}
}

type DebugLevelRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Show      bool   `protobuf:"varint,1,opt,name=show,proto3" json:"show,omitempty"`
	LevelSpec string `protobuf:"bytes,2,opt,name=level_spec,json=levelSpec,proto3" json:"level_spec,omitempty"`
}

func sendDebugLevelCommand(conn *websocket.Conn) error {

	//	create a debugLevel request message
	var webSocketReq = WebSocketRequest{
		Endpoint:  "/api/v1/meta/debuglevel",
		RequestId: uuid.New().String(),
	}
	var debugLevelReq = &DebugLevelRequest{
		Show:      true,
		LevelSpec: "debug",
	}

	debugLevelMessage, err := json.Marshal(debugLevelReq)
	if err != nil {
		return errors.New("Fail marshling debug level message: " + err.Error())
	}
	webSocketReq.Payload = debugLevelMessage

	//	marshal the request message to a JSon
	requestMessage, err := json.Marshal(webSocketReq)
	if err != nil {
		return errors.New("Fail marshling webSocker request message: " + err.Error())
	}

	fmt.Printf("[trace] debugLevel command request message: %s\n", requestMessage)

	//	write the result to the webSocket client
	err = conn.WriteMessage(websocket.TextMessage, requestMessage)
	if err != nil {
		return errors.New("Fail writing message to pld: " + err.Error())
	}

	//	get response message
	messageType, message, err := conn.ReadMessage()
	if err != nil {
		return errors.New("Fail reading message from pld: " + err.Error())
	}

	if messageType == websocket.TextMessage {
		fmt.Printf("[trace] debugLevel command response message: %s\n", message)

		//	unmarshal the response message to a JSon
		var responseMessage WebSocketResponse

		json.Unmarshal(message, &responseMessage)

		if responseMessage.Error != nil {
			return errors.New("Error response message received from pld: " + responseMessage.Error.Error.Message)
		}

		fmt.Printf("--> debugLevel response payload: %s\n", responseMessage.Payload)
	}

	return nil
}

func sendGetInfoCommand(conn *websocket.Conn) error {

	//	create a getInfo request message
	var webSocketReq = WebSocketRequest{
		Endpoint:  "/api/v1/meta/getinfo",
		RequestId: uuid.New().String(),
	}

	//	marshal the request message to a JSon
	requestMessage, err := json.Marshal(webSocketReq)
	if err != nil {
		return errors.New("Fail marshling webSocker request message: " + err.Error())
	}

	//	write the result to the webSocket client
	err = conn.WriteMessage(websocket.TextMessage, requestMessage)
	if err != nil {
		return errors.New("Fail writing message to pld: " + err.Error())
	}

	//	get response message
	messageType, message, err := conn.ReadMessage()
	if err != nil {
		return errors.New("Fail reading message from pld: " + err.Error())
	}

	if messageType == websocket.TextMessage {
		//	fmt.Printf("[trace] getInfo command response message: %s\n", message)

		//	unmarshal the response message to a JSon
		var resp WebSocketResponse

		json.Unmarshal(message, &resp)

		if resp.Error != nil {
			return errors.New("Error response message received from pld: " + resp.Error.Error.Message)
		}

		fmt.Printf("--> GetInfo response payload: %s\n", resp.Payload)
	}

	return nil
}
