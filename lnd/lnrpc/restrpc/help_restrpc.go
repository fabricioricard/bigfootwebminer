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

//	convert	pkthelp.type to REST help proto struct
func convertHelpType(t pkthelp.Type) *Type {
	resultType := &Type{
		Name:        t.Name,
		Description: t.Description,
	}

	//	convert the array of fields
	for _, field := range t.Fields {
		resultType.Fields = append(resultType.Fields, &Field{
			Name:        field.Name,
			Description: field.Description,
			Repeated:    field.Repeated,
			Type:        convertHelpType(field.Type),
		})
	}

	return resultType
}

//	marshal rest help response
func marshalHelp(httpResponse http.ResponseWriter, helpInfo pkthelp.Method) er.R {

	marshaler := jsonpb.Marshaler{
		OrigName:     false,
		EnumsAsInts:  false,
		EmitDefaults: false,
		Indent:       "   ",
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

//	the REST master help messsage
func RESTCategory_help(category string) *RestCommandCategory {
	restCommandCategory := &RestCommandCategory{
		Description: []string{""},
		Endpoints:   make(map[string]string),
	}

	for _, function := range rpcFunctions {
		if function.category == category {
			restCommandCategory.Endpoints[URI_prefix+function.path] = function.description
		}
	}

	return restCommandCategory
}

//	the REST master help messsage
func RESTMaster_help() *RestMasterHelpResponse {
	return &RestMasterHelpResponse{
		Name: "pld - Lightning Network Daemon REST interface (pld)",
		Description: []string{
			"General information about PLD",
		},
		Category: map[string]*RestCommandCategory{
			categoryMeta:       RESTCategory_help(categoryMeta),
			categoryChannels:   RESTCategory_help(categoryChannels),
			categoryGraph:      RESTCategory_help(categoryGraph),
			categoryInvoices:   RESTCategory_help(categoryInvoices),
			categoryOnChain:    RESTCategory_help(categoryOnChain),
			categoryPayments:   RESTCategory_help(categoryPayments),
			categoryPeers:      RESTCategory_help(categoryPeers),
			categoryStartup:    RESTCategory_help(categoryStartup),
			categoryWallet:     RESTCategory_help(categoryWallet),
			categoryWatchtower: RESTCategory_help(categoryWatchtower),
		},
	}
}

//	add the REST master help HTTP handler
func RestHandlersHelp(router *mux.Router) {
	router.HandleFunc("/", getMainHelp)
	router.HandleFunc(URI_prefix, getMainHelp)
}

//	get REST master help handler
func getMainHelp(httpResponse http.ResponseWriter, httpRequest *http.Request) {

	//	fill response payload
	if httpRequest.Method != "GET" {
		httpResponse.Header().Set("Content-Type", "text/plain")
		http.Error(httpResponse, "400 - Request should be a GET because the help endpoint requires no input", http.StatusBadRequest)
		return
	}
	//	marshal the REST master help
	marshaler := jsonpb.Marshaler{
		OrigName:     false,
		EnumsAsInts:  false,
		EmitDefaults: false,
		Indent:       "   ",
	}

	s, err := marshaler.MarshalToString(RESTMaster_help())
	if err != nil {
		httpResponse.Header().Set("Content-Type", "text/plain")
		http.Error(httpResponse, "500 - Internal Error", http.StatusInternalServerError)
		log.Errorf("Error replying to request for [%s] from [%s] - error sending error, giving up: [%s]",
			httpRequest.RequestURI, httpRequest.RemoteAddr, err)

		return
	}

	_, err = io.WriteString(httpResponse, s)
	if err != nil {
		httpResponse.Header().Set("Content-Type", "text/plain")
		http.Error(httpResponse, "500 - Internal Error", http.StatusInternalServerError)
		log.Errorf("Error replying to request for [%s] from [%s] - error sending error, giving up: [%s]",
			httpRequest.RequestURI, httpRequest.RemoteAddr, err)
	}
}
