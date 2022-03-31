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
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkt-cash/pktd/lnd/lnrpc"
	"github.com/pkt-cash/pktd/lnd/lnrpc/restrpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
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

	fmt.Printf("[info] testing JSon based messages\n")

	//	send a JSon based DebugLevel command
	err = sendJSonDebugLevelCommand(conn)
	if err != nil {
		log.Println(err.Error())
		return
	}

	//	send a JSon based GetInfo command
	err = sendGetInfoCommand(conn)
	if err != nil {
		log.Println(err.Error())
		return
	}

	fmt.Printf("[info] testing Protobuf based messages\n")

	//	send a Protobuf based DebugLevel command
	err = sendProtobufDebugLevelCommand(conn)
	if err != nil {
		log.Println(err.Error())
		return
	}

	//	send a Protobuf based GetInfo command
	err = sendProtobufGetInfoCommand(conn)
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

func sendJSonDebugLevelCommand(conn *websocket.Conn) error {

	//	create and marshal a debugLevel payload
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

	//	write the request to the webSocket client
	err = conn.WriteMessage(websocket.TextMessage, requestMessage)
	if err != nil {
		return errors.New("Fail writing message to pld: " + err.Error())
	}

	//	get response message
	messageType, message, err := conn.ReadMessage()
	if err != nil {
		return errors.New("Fail reading message from pld: " + err.Error())
	}

	if messageType != websocket.TextMessage {
		return errors.New("expecting a text based response message from pld")
	}

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

	//	write the request to the webSocket client
	err = conn.WriteMessage(websocket.TextMessage, requestMessage)
	if err != nil {
		return errors.New("Fail writing message to pld: " + err.Error())
	}

	//	get response message
	messageType, message, err := conn.ReadMessage()
	if err != nil {
		return errors.New("Fail reading message from pld: " + err.Error())
	}

	if messageType != websocket.TextMessage {
		return errors.New("expecting a text based response message from pld")
	}

	//	unmarshal the response message to a JSon
	var resp restrpc.WebSocketJSonResponse

	jsoniter.Unmarshal(message, &resp)

	if resp.Error.HttpCode != 0 {
		return errors.New("Error response message received from pld: " + resp.Error.Message)
	}

	fmt.Printf("--> GetInfo response payload: %s\n", resp.Payload)

	return nil
}

func sendProtobufDebugLevelCommand(conn *websocket.Conn) error {

	//	create and marshal a debugLevel payload
	var debugLevelReq = lnrpc.DebugLevelRequest{
		Show:      true,
		LevelSpec: "info",
	}

	debugLevelPayload, err := proto.Marshal(&debugLevelReq)
	if err != nil {
		return errors.New("Fail marshling debug level request message: " + err.Error())
	}

	//	marshal the request message to a Protobuf
	var req = restrpc.WebSocketProtobufRequest{
		Endpoint:  "/api/v1/meta/debuglevel",
		RequestId: uuid.New().String(),
		Payload: &anypb.Any{
			TypeUrl: "github.com/pkt-cash/pktd/lnd/lnrpc.DebugLevelRequest",
			Value:   debugLevelPayload,
		},
	}

	requestMessage, err := proto.Marshal(&req)
	if err != nil {
		return errors.New("Fail marshling webSocker request message: " + err.Error())
	}

	//	write the request to the webSocket client
	err = conn.WriteMessage(websocket.BinaryMessage, requestMessage)
	if err != nil {
		return errors.New("Fail writing message to pld: " + err.Error())
	}

	//	get response message
	messageType, message, err := conn.ReadMessage()
	if err != nil {
		return errors.New("Fail reading message from pld: " + err.Error())
	}

	if messageType != websocket.BinaryMessage {
		return errors.New("expecting a binary based response message from pld")
	}

	//	reflect the webSocket response protobuf type
	var webSocketRespProto proto.Message = (*restrpc.WebSocketProtobufResponse)(nil)

	webSocketProtobuf := reflect.New(reflect.TypeOf(webSocketRespProto).Elem())
	respMessage, _ := webSocketProtobuf.Interface().(proto.Message)

	//	unmarshal the response message
	err = proto.Unmarshal(message, respMessage)
	if err != nil {
		return errors.New("Fail unmarshaling the response message: " + err.Error())
	}

	webSocketResp, ok := respMessage.(*restrpc.WebSocketProtobufResponse)
	if !ok {
		return errors.New("Request message is not a WebSocketProtobufResponse")
	}

	if webSocketResp.GetError() != nil {
		return errors.New("Error response message received from pld: " + webSocketResp.GetError().GetMessage())
	}

	//	unmarshal the DebugLevel response value
	var debugLevelResp lnrpc.DebugLevelResponse

	err = proto.Unmarshal(webSocketResp.GetOk().GetValue(), &debugLevelResp)
	if err != nil {
		return errors.New("Fail unmarshaling the payload message: " + err.Error())
	}

	fmt.Printf("--> debugLevel response payload: %s\n", debugLevelResp.String())

	return nil
}

func sendProtobufGetInfoCommand(conn *websocket.Conn) error {

	//	marshal the request message to a Protobuf
	var requestId = uuid.New().String()
	var req = restrpc.WebSocketProtobufRequest{
		Endpoint:  "/api/v1/meta/getinfo",
		RequestId: requestId,
		Payload: &anypb.Any{
			TypeUrl: "github.com/pkt-cash/pktd/lnd/lnrpc.GetInfoRequest",
			Value:   nil,
		},
	}

	requestMessage, err := proto.Marshal(&req)
	if err != nil {
		return errors.New("Fail marshling webSocker request message: " + err.Error())
	}

	//	write the request to the webSocket client
	err = conn.WriteMessage(websocket.BinaryMessage, requestMessage)
	if err != nil {
		return errors.New("Fail writing message to pld: " + err.Error())
	}

	fmt.Printf("[trace] GetInfo command requestId: %s\n", requestId)

	//	get response message
	messageType, message, err := conn.ReadMessage()
	if err != nil {
		return errors.New("Fail reading message from pld: " + err.Error())
	}

	if messageType != websocket.BinaryMessage {
		return errors.New("expecting a binary based response message from pld")
	}

	//	reflect the webSocket response protobuf type
	var webSocketRespProto proto.Message = (*restrpc.WebSocketProtobufResponse)(nil)

	webSocketProtobuf := reflect.New(reflect.TypeOf(webSocketRespProto).Elem())
	respMessage, _ := webSocketProtobuf.Interface().(proto.Message)

	//	unmarshal the response message
	err = proto.Unmarshal(message, respMessage)
	if err != nil {
		return errors.New("Fail unmarshaling the response message: " + err.Error())
	}

	webSocketResp, ok := respMessage.(*restrpc.WebSocketProtobufResponse)
	if !ok {
		return errors.New("Request message is not a WebSocketProtobufResponse")
	}

	fmt.Printf("[trace] GetInfo response received: requestId: %s\n", webSocketResp.RequestId)

	if webSocketResp.GetError() != nil {
		return errors.New("Error response message received from pld: " + webSocketResp.GetError().GetMessage())
	}

	fmt.Printf("[trace] Response payload TypeUrl: %s\n", webSocketResp.GetOk().TypeUrl)

	//	reflect the response value protobuf type
	var getInfoRespProto proto.Message = (*lnrpc.GetInfoResponse)(nil)

	valueProto := reflect.New(reflect.TypeOf(getInfoRespProto).Elem())
	valueMessage, _ := valueProto.Interface().(proto.Message)

	//	unmarshal the GetInfo response value
	err = proto.Unmarshal(webSocketResp.GetOk().Value, valueMessage)
	if err != nil {
		return errors.New("Fail unmarshaling the payload message: " + err.Error())
	}

	//	get the response payload
	var getInfoResp *lnrpc.GetInfoResponse

	getInfoResp, ok = valueMessage.(*lnrpc.GetInfoResponse)
	if !ok {
		return errors.New("Response payload message is not a GetInfoResponse: " + err.Error())
	}

	fmt.Printf("--> GetInfo response payload: %s\n", getInfoResp.String())

	return nil
}
