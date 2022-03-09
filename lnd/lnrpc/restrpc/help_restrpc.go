package restrpc

import (
	"io"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/lnd/pkthelp"
	"github.com/pkt-cash/pktd/pktlog/log"
)

const (
	helpURI_prefix = "/help"
)

//	the main help messsage
func mainREST_help() pkthelp.Method {
	return pkthelp.Method{
		Name: "pld - Lightning Network Daemon REST interface (pld)",
		Description: []string{
			"For help on a specific command, use the following URIs:",
			"    getinfo          /api/v1/help/meta/getinfo",
			"    getrecoveryinfo  /api/v1/meta/getrecoveryinfo",
			"    debuglevel       /api/v1/meta/debuglevel",
			"    stop             /api/v1/meta/stop",
			"    version          /api/v1/meta/version",
		},
	}
}

//	convert	type to proto struct
func convertHelpType(t pkthelp.Type) *Type {
	resultType := &Type{
		Name:        t.Name,
		Description: t.Description,
	}

	return resultType
}

//	marshal rest help response
func marshalHelp(httpResponse http.ResponseWriter, helpInfo pkthelp.Method) er.R {

	marshaler := jsonpb.Marshaler{
		OrigName:     false,
		EnumsAsInts:  false,
		EmitDefaults: true,
		Indent:       "\t",
	}

	s, err := marshaler.MarshalToString(&RestHelpResponse{
		Name:        helpInfo.Name,
		Service:     helpInfo.Service,
		Description: helpInfo.Description,
		Request:     convertHelpType(helpInfo.Req),
		Response:    convertHelpType(helpInfo.Res),
	})
	if err != nil {
		return er.E(err)
	}

	_, err = io.WriteString(httpResponse, s)
	if err != nil {
		return er.E(err)
	}

	return nil
}

//	add the main help HTTP handler
func RestHandlersHelp(router *mux.Router) {
	router.HandleFunc("/", getMainHelp)
}

//	get main help
func getMainHelp(httpResponse http.ResponseWriter, httpRequest *http.Request) {

	//	fill response payload
	if httpRequest.Method != "GET" {
		httpResponse.Header().Set("Content-Type", "text/plain")
		http.Error(httpResponse, "400 - Request should be a GET because the help endpoint requires no input", http.StatusBadRequest)
		return
	}
	err := marshalHelp(httpResponse, mainREST_help())
	if err != nil {
		httpResponse.Header().Set("Content-Type", "text/plain")
		http.Error(httpResponse, "500 - Internal Error", http.StatusInternalServerError)
		log.Errorf("Error replying to request for [%s] from [%s] - error sending error, giving up: [%s]",
			httpRequest.RequestURI, httpRequest.RemoteAddr, err)
	}
}
