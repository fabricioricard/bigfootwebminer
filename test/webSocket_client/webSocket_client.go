////////////////////////////////////////////////////////////////////////////////
//	test/webSocket_client.go  -  Mar-22-2022  -  aldebap
//
//	simple webSocket_client to test pld REST APIs
////////////////////////////////////////////////////////////////////////////////

package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkt-cash/pktd/lnd/lnrpc"
	"github.com/pkt-cash/pktd/lnd/lnrpc/restrpc"
)

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

	//	send a close socket message to server
	err = conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Client closing"),
	)
	if err != nil {
		log.Fatal("Error closing the Websocket:", err)
	}

	//	sleep so the server can read the close message before the socket be effectively closed
	time.Sleep(1 * time.Second)
}

func sendDebugLevelCommand(conn *websocket.Conn) error {

	//	create a debugLevel payload
	var debugLevelReq = lnrpc.DebugLevelRequest{
		Show:      true,
		LevelSpec: "debug",
	}

	debugLevelPayload, err := jsoniter.Marshal(&debugLevelReq)
	if err != nil {
		return errors.New("Fail marshling debug level message: " + err.Error())
	}

	//	marshal the request message to a JSon
	var req = restrpc.WebSocketJSonRequest{
		Endpoint:  "/api/v1/meta/debuglevel",
		RequestId: uuid.New().String(),
		Payload:   debugLevelPayload,
	}

	requestMessage, err := jsoniter.Marshal(req)
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

		//	unmarshal the response message
		var resp restrpc.WebSocketJSonResponse

		err = jsoniter.Unmarshal(message, &resp)
		if err != nil {
			return errors.New("Fail parsing payload message: " + err.Error())
		}

		if resp.Error.HttpCode != 0 {
			return errors.New("Error response message received from pld: " + resp.Error.Message)
		}

		//	unmarshal the payload within response
		var debugLevelResp lnrpc.DebugLevelResponse

		err = jsoniter.Unmarshal(resp.Payload, &debugLevelResp)
		if err != nil {
			return errors.New("Fail parsing payload message: " + err.Error())
		}

		fmt.Printf("--> debugLevel response payload: %s\n", resp.Payload)
	}

	return nil
}

func sendGetInfoCommand(conn *websocket.Conn) error {

	//	marshal the request message to a JSon
	var req = restrpc.WebSocketJSonRequest{
		Endpoint:  "/api/v1/meta/getinfo",
		RequestId: uuid.New().String(),
		Payload:   nil,
	}

	requestMessage, err := jsoniter.Marshal(req)
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

		//	unmarshal the response message to a JSon
		var resp restrpc.WebSocketJSonResponse

		jsoniter.Unmarshal(message, &resp)

		if resp.Error.HttpCode != 0 {
			return errors.New("Error response message received from pld: " + resp.Error.Message)
		}

		fmt.Printf("--> GetInfo response payload: %s\n", resp.Payload)
	}

	return nil
}
