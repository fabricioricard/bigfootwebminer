////////////////////////////////////////////////////////////////////////////////
//	websocket.go  -  Mar-24-2022  -  aldebap
//
//	websocket handler for pld commands
////////////////////////////////////////////////////////////////////////////////

package restrpc

import (
	"bytes"
	"encoding/json"
	"net/http"
	reflect "reflect"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkt-cash/pktd/pktlog/log"
	"google.golang.org/protobuf/runtime/protoiface"
)

type websocketConn websocket.Conn

type WebSocketJSonRequest struct {
	Endpoint  string          `json:"endpoint,omitempty"`
	RequestId string          `json:"request_id,omitempty"`
	HasMore   bool            `json:"has_more,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

type WebSocketJSonResponse struct {
	RequestId string          `json:"request_id,omitempty"`
	HasMore   bool            `json:"has_more,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Error     WebSocketError  `json:"error,omitempty"`
}

var upgrader = websocket.Upgrader{}

func webSocketHandler(ctx *RpcContext, httpResponse http.ResponseWriter, httpRequest *http.Request) {
	//	upgrade raw HTTP connection to a websocket
	conn, err := upgrader.Upgrade(httpResponse, httpRequest, nil)
	if err != nil {
		httpResponse.Header().Set("Content-Type", "text/plain")
		http.Error(httpResponse, "503 - Service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer conn.Close()

	//	webSocket communication loop
	var wsConn websocketConn = websocketConn(*conn)

	for {
		msgType, message, err := conn.ReadMessage()
		if err != nil {
			log.Errorf("Fail during message reading:", err)
			return
		}

		//	handle the messages according to it's type
		switch msgType {

		case websocket.TextMessage:
			wsConn.HandleJsonMessage(ctx, message)

		case websocket.BinaryMessage:
			wsConn.HandleProtobufMessage(ctx, message)

		case websocket.CloseMessage:
			log.Info("WebSocket closed by the client")
			return

		default:
			wsConn.WriteErrorMessage("", "Expecting a text/json or binary/protobuf request message", nil)
		}
	}
}

func (conn *websocketConn) HandleJsonMessage(ctx *RpcContext, jsonMessage []byte) {

	//	unmarshal the request message
	var webSocketReq WebSocketJSonRequest

	err := jsoniter.Unmarshal(jsonMessage, &webSocketReq)
	if err != nil {
		conn.WriteErrorMessage("", "Error unmarshaling the request message", err)
		return
	}

	//	based on the endpoint, find the appropriate handler for the message request
	var endpoint = strings.TrimPrefix(string(webSocketReq.Endpoint), URI_prefix)

	log.Info("WebSocket message received for endpoint: " + endpoint)

	for _, rpcFunc := range rpcFunctions {
		if endpoint == rpcFunc.path {
			var valueMessage protoiface.MessageV1 = nil

			if rpcFunc.req != nil {
				//	reflect the request value protobuf type
				valueProto := reflect.New(reflect.TypeOf(rpcFunc.req).Elem())
				valueMessage, _ = valueProto.Interface().(proto.Message)

				//	unmarshal the request value
				err = jsonpb.Unmarshal(bytes.NewReader(webSocketReq.Payload), valueMessage)
				if err != nil {
					conn.WriteErrorMessage(webSocketReq.RequestId, "Error unmarshaling the request value message", err)
					return
				}
			}

			//	invoke the RPC message handler
			responseMessage, errr := rpcFunc.f(ctx, valueMessage)
			if errr != nil {
				conn.WriteErrorMessage(webSocketReq.RequestId, "Error handling the request value message", errr.Native())
				return
			}

			//	marshal the response message
			marshaler := jsonpb.Marshaler{
				OrigName:     false,
				EnumsAsInts:  false,
				EmitDefaults: true,
				Indent:       "    ",
			}

			respPayload, err := marshaler.MarshalToString(responseMessage)
			if err != nil {
				conn.WriteErrorMessage(webSocketReq.RequestId, "Error marshaling the response value message", err)
				return
			}

			//	write the result message to the webSocket client
			conn.WriteMessage(webSocketReq.RequestId, []byte(respPayload))
			return
		}
	}

	//	if the requested endpoint wasn't found, send an error message to the client
	conn.WriteErrorMessage(webSocketReq.RequestId, "Unknown endpoint URI: "+webSocketReq.Endpoint, err)
}

func (conn *websocketConn) HandleProtobufMessage(ctx *RpcContext, protobufMessage []byte) {

	//	reflect the webSocket request protobuf type
	var webSocketReqProto proto.Message = (*WebSocketProtobufRequest)(nil)

	webSocketProtobuf := reflect.New(reflect.TypeOf(webSocketReqProto).Elem())
	reqMessage, _ := webSocketProtobuf.Interface().(proto.Message)

	//	unmarshal the request message
	err := jsonpb.Unmarshal(bytes.NewReader(protobufMessage), reqMessage)
	if err != nil {
		conn.WriteErrorMessage("", "Error unmarshaling the request message", err)
		return
	}

	webSocketReq, ok := reqMessage.(*WebSocketProtobufRequest)
	if !ok {
		conn.WriteErrorMessage(webSocketReq.GetRequestId(), "Request message is not a WebSocketProtobufRequest", nil)
		return
	}

	//	based on the endpoint, find the appropriate handler for the message request
	var endpoint = strings.TrimPrefix(string(webSocketReq.Endpoint), URI_prefix)

	log.Info("WebSocket message received for endpoint: " + endpoint)

	for _, rpcFunc := range rpcFunctions {
		if endpoint == rpcFunc.path {
			var valueMessage protoiface.MessageV1 = nil

			if rpcFunc.req != nil {
				//	reflect the request value protobuf type
				valueProto := reflect.New(reflect.TypeOf(rpcFunc.req).Elem())
				valueMessage, _ = valueProto.Interface().(proto.Message)

				//	unmarshal the request value
				err = jsonpb.Unmarshal(bytes.NewReader(webSocketReq.GetPayload().Value), valueMessage)
				if err != nil {
					conn.WriteErrorMessage(webSocketReq.RequestId, "Error unmarshaling the request value message", err)
					return
				}
			}

			//	invoke the RPC message handler
			responseMessage, errr := rpcFunc.f(ctx, valueMessage)
			if errr != nil {
				conn.WriteErrorMessage(webSocketReq.RequestId, "Error handling the request value message", errr.Native())
				return
			}

			//	marshal the response message
			marshaler := jsonpb.Marshaler{
				OrigName:     false,
				EnumsAsInts:  false,
				EmitDefaults: true,
				Indent:       "    ",
			}

			respPayload, err := marshaler.MarshalToString(responseMessage)
			if err != nil {
				conn.WriteErrorMessage(webSocketReq.RequestId, "Error marshaling the response value message", err)
				return
			}

			//	write the result message to the webSocket client
			conn.WriteMessage(webSocketReq.RequestId, []byte(respPayload))
			return
		}
	}

	//	if the requested endpoint wasn't found, send an error message to the client
	conn.WriteErrorMessage(webSocketReq.RequestId, "Unknown endpoint URI: "+webSocketReq.Endpoint, err)
}

func (conn *websocketConn) WriteMessage(requestId string, responsePayload []byte) {

	//	valid response message
	var responseMsg = &WebSocketJSonResponse{
		RequestId: requestId,
		Payload:   responsePayload,
	}

	//	marshal the response message
	resp, err := jsoniter.Marshal(responseMsg)
	if err != nil {
		conn.WriteErrorMessage(requestId, "Error marshaling the response value message", err)
		return
	}

	//	write the result message to the webSocket client
	err = (*websocket.Conn)(conn).WriteMessage(websocket.TextMessage, resp)
	if err != nil {
		log.Errorf("Cannot write error message to webSocket client: ", err)
	}
}

func (conn *websocketConn) WriteErrorMessage(requestId string, errorMsg string, err error) {

	//	error response message
	var errorResponseMsg = &WebSocketJSonResponse{
		RequestId: requestId,
		Error:     WebSocketError{},
	}

	if err == nil {
		errorResponseMsg.Error = WebSocketError{
			HttpCode: 500,
			Message:  errorMsg,
		}
	} else {
		errorResponseMsg.Error = WebSocketError{
			HttpCode: 500,
			Message:  errorMsg + ": " + err.Error(),
		}
	}

	//	marshal the response message
	resp, err := jsoniter.Marshal(errorResponseMsg)
	if err != nil {
		log.Errorf("Error marshaling the response value message: ", err)
		return
	}

	//	write the error message to the webSocket client
	err = (*websocket.Conn)(conn).WriteMessage(websocket.TextMessage, resp)
	if err != nil {
		log.Errorf("Cannot write error message to webSocket client: ", err)
	}
}
