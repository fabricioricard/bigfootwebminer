////////////////////////////////////////////////////////////////////////////////
//	test/webSocket_client.go  -  Mar-22-2022  -  aldebap
//
//	simple webSocket_client to test pld REST APIs
////////////////////////////////////////////////////////////////////////////////

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

type DebugLevelRequest struct {
	Show      bool   `json:"show,omitempty"`
	LevelSpec string `json:"level_spec,omitempty"`
}

type Any struct {
	TypeUrl string `json:"type_url,omitempty"`
	Value   []byte `json:"value,omitempty"`
}

type WebSocketRequest struct {
	Endpoint  string `json:"endpoint,omitempty"`
	RequestId string `json:"request_id,omitempty"`
	Payload   *Any   `json:"payload,omitempty"`
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

	//	create a debugLevel request message
	var debugLevelReq = DebugLevelRequest{
		Show:      true,
		LevelSpec: "debug",
	}
	var webSocketReq = WebSocketRequest{
		Endpoint:  "/api/v1/meta/debuglevel",
		RequestId: "req1234",
		Payload:   &Any{},
	}

	debugLevelMessage, err := json.Marshal(debugLevelReq)
	if err != nil {
		log.Println("Error marshling debug level message:", err)
		return
	}
	webSocketReq.Payload.Value = debugLevelMessage

	//	marshal the request message to a JSon
	requestMessage, err := json.Marshal(webSocketReq)
	if err != nil {
		log.Println("Error marshling webSocker request message:", err)
		return
	}

	//	write the result to the webSocket client
	err = conn.WriteMessage(websocket.TextMessage, requestMessage)
	if err != nil {
		log.Println("Error writing message to pld:", err)
		return
	}

	//	get response message
	messageType, message, err := conn.ReadMessage()
	if err != nil {
		log.Println("Error during message reading:", err)
		return
	}
	if messageType == websocket.TextMessage {
		fmt.Printf("debugLevel command response message: %s\n", message)
	}

	//	create a getInfo request message
	webSocketReq = WebSocketRequest{
		Endpoint:  "/api/v1/meta/getinfo",
		RequestId: "req2345",
		Payload:   &Any{},
	}

	//	marshal the request message to a JSon
	requestMessage, err = json.Marshal(webSocketReq)
	if err != nil {
		log.Println("Error marshling webSocker request message:", err)
		return
	}

	//	write the result to the webSocket client
	err = conn.WriteMessage(websocket.TextMessage, requestMessage)
	if err != nil {
		log.Println("Error writing message to pld:", err)
		return
	}

	//	get response message
	messageType, message, err = conn.ReadMessage()
	if err != nil {
		log.Println("Error during message reading:", err)
		return
	}
	if messageType == websocket.TextMessage {
		fmt.Printf("getInfo command response message: %s\n", message)
	}
}
