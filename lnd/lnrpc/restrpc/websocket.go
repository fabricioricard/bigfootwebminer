////////////////////////////////////////////////////////////////////////////////
//	websocket.go  -  Mar-24-2022  -  aldebap
//
//	websocket handler for pld commands
////////////////////////////////////////////////////////////////////////////////

package restrpc

import (
	"bytes"
	"net/http"
	reflect "reflect"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"github.com/pkt-cash/pktd/pktlog/log"
	"google.golang.org/protobuf/runtime/protoiface"
)

type websocketConn websocket.Conn

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

		default:
			wsConn.WriteErrorMessage("", "Expecting a text/json or binary/protobuf request message", nil)
		}
	}
}

func (conn *websocketConn) HandleJsonMessage(ctx *RpcContext, jsonMessage []byte) {

	//	reflect the webSocket request protobuf type
	var webSocketReqProto proto.Message = (*WebSocketJSonRequest)(nil)

	webSocketProto := reflect.New(reflect.TypeOf(webSocketReqProto).Elem())
	reqMessage, _ := webSocketProto.Interface().(proto.Message)

	//	unmarshal the request message
	err := jsonpb.Unmarshal(bytes.NewReader(jsonMessage), reqMessage)
	if err != nil {
		conn.WriteErrorMessage("", "Error unmarshaling the request message", err)
		return
	}

	webSocketReq, ok := reqMessage.(*WebSocketJSonRequest)
	if !ok {
		conn.WriteErrorMessage(webSocketReq.GetRequestId(), "Request message is not a WebSocketRequest", nil)
		return
	}

	//	based on the endpoint, find the appropriate handler for the message request
	var endpoint = strings.TrimPrefix(webSocketReq.GetEndpoint(), URI_prefix)

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
					conn.WriteErrorMessage(webSocketReq.GetRequestId(), "Error unmarshaling the request value message", err)
					break
				}
			}

			//	invoke the RPC message handler
			responseMessage, errr := rpcFunc.f(ctx, valueMessage)
			if errr != nil {
				conn.WriteErrorMessage(webSocketReq.GetRequestId(), "Error handling the request value message", errr.Native())
				break
			}

			//	marshal the response message
			marshaler := jsonpb.Marshaler{
				OrigName:     false,
				EnumsAsInts:  false,
				EmitDefaults: true,
				Indent:       "    ",
			}

			resp, err := marshaler.MarshalToString(responseMessage)
			if err != nil {
				conn.WriteErrorMessage(webSocketReq.GetRequestId(), "Error marshaling the response value message", err)
				break
			}

			//	write the result message to the webSocket client
			conn.WriteMessage(webSocketReq.GetRequestId(), []byte(resp))
			break
		}
	}
}

func (conn *websocketConn) HandleProtobufMessage(ctx *RpcContext, jsonMessage []byte) {
}

func (conn *websocketConn) WriteMessage(requestId string, response []byte) {

	//	valid response message
	var responseMsg = &WebSocketJSonResponse{
		RequestId: requestId,
		Payload: &WebSocketJSonResponse_Ok{
			Ok: response,
		},
	}

	//	marshal the response message
	marshaler := jsonpb.Marshaler{
		OrigName:     false,
		EnumsAsInts:  false,
		EmitDefaults: true,
		Indent:       "    ",
	}

	resp, err := marshaler.MarshalToString(responseMsg)
	if err != nil {
		conn.WriteErrorMessage(requestId, "Error marshaling the response value message", err)
		return
	}

	//	write the result message to the webSocket client
	err = (*websocket.Conn)(conn).WriteMessage(websocket.TextMessage, []byte(resp))
	if err != nil {
		log.Errorf("Cannot write error message to webSocket client: ", err)
	}
}

func (conn *websocketConn) WriteErrorMessage(requestId string, errorMsg string, err error) {

	//	error response message
	var errorResponseMsg = &WebSocketJSonResponse{
		RequestId: requestId,
		Payload:   &WebSocketJSonResponse_Error{},
	}

	if err == nil {
		errorResponseMsg.Payload = &WebSocketJSonResponse_Error{
			Error: &WebSocketError{
				HttpCode: 500,
				Error: &Error{
					Message: errorMsg,
				},
			},
		}
	} else {
		errorResponseMsg.Payload = &WebSocketJSonResponse_Error{
			Error: &WebSocketError{
				HttpCode: 500,
				Error: &Error{
					Message: errorMsg + ": " + err.Error(),
				},
			},
		}
	}

	//	marshal the response message
	marshaler := jsonpb.Marshaler{
		OrigName:     false,
		EnumsAsInts:  false,
		EmitDefaults: true,
		Indent:       "    ",
	}

	resp, err := marshaler.MarshalToString(errorResponseMsg)
	if err != nil {
		log.Errorf("Error marshaling the response value message: ", err)
		return
	}

	//	write the error message to the webSocket client
	err = (*websocket.Conn)(conn).WriteMessage(websocket.TextMessage, []byte(resp))
	if err != nil {
		log.Errorf("Cannot write error message to webSocket client: ", err)
	}
}
