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

		//	expected message is a text/jSon
		if msgType != websocket.TextMessage {
			wsConn.WriteErrorMessage("", "Expecting a text/json request message", nil)
			continue
		}

		//	reflect the webSocket request protobuf type
		var webSocketReqProto proto.Message = (*WebSocketRequest)(nil)

		webSocketProto := reflect.New(reflect.TypeOf(webSocketReqProto).Elem())
		reqMessage, _ := webSocketProto.Interface().(proto.Message)

		//	unmarshal the request message
		err = jsonpb.Unmarshal(bytes.NewReader(message), reqMessage)
		if err != nil {
			wsConn.WriteErrorMessage("", "Error unmarshaling the request message", err)
			continue
		}

		webSocketReq, ok := reqMessage.(*WebSocketRequest)
		if !ok {
			wsConn.WriteErrorMessage(webSocketReq.GetRequestId(), "Request message is not a WebSocketRequest", nil)
			continue
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
					//err = jsonpb.Unmarshal(bytes.NewReader(webSocketReq.Payload.Value), valueMessage)
					err = jsonpb.Unmarshal(bytes.NewReader(webSocketReq.Payload), valueMessage)
					if err != nil {
						wsConn.WriteErrorMessage(webSocketReq.GetRequestId(), "Error unmarshaling the request value message", err)
						break
					}
				}

				//	invoke the RPC message handler
				responseMessage, errr := rpcFunc.f(ctx, valueMessage)
				if errr != nil {
					wsConn.WriteErrorMessage(webSocketReq.GetRequestId(), "Error handling the request value message", errr.Native())
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
					wsConn.WriteErrorMessage(webSocketReq.GetRequestId(), "Error marshaling the response value message", err)
					break
				}

				//	write the result message to the webSocket client
				wsConn.WriteMessage(webSocketReq.GetRequestId(), []byte(resp))
			}
		}
	}
}

type websocketConn websocket.Conn

func (conn *websocketConn) WriteMessage(requestId string, response []byte) {

	//	valid response message
	/*
		var responseMsg = &WebSocketResponse{
			RequestId: requestId,
			Payload: &WebSocketResponse_Ok{
				Ok: &anypb.Any{
					TypeUrl: "ws",
					Value:   response,
				},
			},
		}
	*/
	var responseMsg = &WebSocketResponse{
		RequestId: requestId,
		Payload: &WebSocketResponse_Ok{
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
	var errorResponseMsg = &WebSocketResponse{
		RequestId: requestId,
		Payload:   &WebSocketResponse_Error{},
	}

	if err == nil {
		errorResponseMsg.Payload = &WebSocketResponse_Error{
			Error: &WebSocketError{
				HttpCode: 500,
				Error: &Error{
					Message: errorMsg,
				},
			},
		}
	} else {
		errorResponseMsg.Payload = &WebSocketResponse_Error{
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
