package restrpc

import (
	"io"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/lnd/lnrpc/restrpc/help"
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
func RESTCategory_help(category string, subCategory map[string]*RestCommandCategory) *RestCommandCategory {
	restCommandCategory := &RestCommandCategory{
		Description: help.CategoryDescription[category],
		Endpoints:   make(map[string]string),
		Subcategory: make(map[string]*RestCommandCategory),
	}

	//	add all endpoints for the category
	for _, function := range rpcFunctions {
		if function.category == category {
			restCommandCategory.Endpoints[URI_prefix+function.path] = function.description
		}
	}

	//	add all sub categories
	for name, value := range subCategory {
		restCommandCategory.Subcategory[name] = value
	}

	return restCommandCategory
}

//	the REST master help messsage
func RESTMaster_help() *RestMasterHelpResponse {
	masterHelpResp := &RestMasterHelpResponse{
		Name: "pld - Lightning Network Daemon REST interface (pld)",
		Description: []string{
			"General information about PLD",
		},
		Category: map[string]*RestCommandCategory{
			help.CategoryLightning: RESTCategory_help(help.CategoryLightning, map[string]*RestCommandCategory{
				help.SubcategoryChannel: RESTCategory_help(help.SubcategoryChannel, map[string]*RestCommandCategory{
					help.SubSubCategoryBackup: RESTCategory_help(help.SubSubCategoryBackup, map[string]*RestCommandCategory{}),
				}),
				help.SubCategoryGraph:   RESTCategory_help(help.SubCategoryGraph, map[string]*RestCommandCategory{}),
				help.SubCategoryInvoice: RESTCategory_help(help.SubCategoryInvoice, map[string]*RestCommandCategory{}),
				help.SubCategoryPayment: RESTCategory_help(help.SubCategoryPayment, map[string]*RestCommandCategory{}),
				help.SubCategoryPeer:    RESTCategory_help(help.SubCategoryPeer, map[string]*RestCommandCategory{}),
			}),
			help.CategoryMeta: RESTCategory_help(help.CategoryMeta, map[string]*RestCommandCategory{}),
			help.CategoryWallet: RESTCategory_help(help.CategoryWallet, map[string]*RestCommandCategory{
				help.SubCategoryNetworkStewardVote: RESTCategory_help(help.SubCategoryNetworkStewardVote, map[string]*RestCommandCategory{}),
				help.SubCategoryTransaction:        RESTCategory_help(help.SubCategoryTransaction, map[string]*RestCommandCategory{}),
				help.SubCategoryUnspent: RESTCategory_help(help.SubCategoryUnspent, map[string]*RestCommandCategory{
					help.SubSubCategoryLock: RESTCategory_help(help.SubSubCategoryLock, map[string]*RestCommandCategory{}),
				}),
				help.SubCategoryAddress: RESTCategory_help(help.SubCategoryAddress, map[string]*RestCommandCategory{}),
			}),
			help.CategoryNeutrino: RESTCategory_help(help.CategoryNeutrino, map[string]*RestCommandCategory{}),
			help.CategoryUtil: RESTCategory_help(help.CategoryUtil, map[string]*RestCommandCategory{
				help.SubCategorySeed: RESTCategory_help(help.SubCategorySeed, map[string]*RestCommandCategory{}),
			}),
			help.CategoryWatchtower: RESTCategory_help(help.CategoryWatchtower, map[string]*RestCommandCategory{}),
		},
	}

	return masterHelpResp
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
		http.Error(httpResponse, "405 - Request should be a GET because the help endpoint requires no input", http.StatusMethodNotAllowed)
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
