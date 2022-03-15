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

var (
	categoryDescription map[string][]string = map[string][]string{
		categoryLightning:  {"The Lightning Network component of the wallet"},
		subcategoryChannel: {"Management of lightning channels to direct peers of this pld node"},
		subSubCategoryBackup: {"Backup and recovery of the state of active Lightning channels",
			"to and from this pld node"},
		subCategoryGraph:   {"Information about the global known Lightning Network"},
		subCategoryInvoice: {"Management of invoices which are used to request payment over Lightning"},
		subCategoryPayment: {"Lightning network payments which have been made, or have been forwarded, through this node"},
		subCategoryPeer:    {"Connections to other nodes in the Lightning Network"},
		categoryMeta:       {"API endpoints which are relevant to the entire pld node, not any specific part"},
		categoryWallet: {"APIs for management of on-chain (non-Lightning) payments,",
			"seed export and recovery, and on-chain transaction detection"},
		subCategoryNetworkStewardVote: {"Control how this wallet votes on PKT Network Steward"},
		subCategoryTransaction:        {"Create and manage on-chain transactions with the wallet"},
		subCategoryUnspent:            {"Detected unspent transactions associated with one of our wallet addresses"},
		subSubCategoryLock: {"Manipulation of unspent outputs which are 'locked'",
			"and therefore will not be used to source funds for any transaction"},
		subCategoryAddress: {"Management of individual wallet addresses"},
		categoryNeutrino:   {"Management of the Neutrino interface which is used to communicate with the p2p nodes in the network"},
		categoryUtil:       {"Stateless utility functions which do not affect, not query, the node in any way"},
		subCategorySeed:    {"Manipulation of mnemonic seed phrases which represent wallet keys"},
		categoryWatchtower: {"Interact with the watchtower client"},
	}
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
		Description: categoryDescription[category],
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
			categoryLightning: RESTCategory_help(categoryLightning, map[string]*RestCommandCategory{
				subcategoryChannel: RESTCategory_help(subcategoryChannel, map[string]*RestCommandCategory{
					subSubCategoryBackup: RESTCategory_help(subSubCategoryBackup, map[string]*RestCommandCategory{}),
				}),
				subCategoryGraph:   RESTCategory_help(subCategoryGraph, map[string]*RestCommandCategory{}),
				subCategoryInvoice: RESTCategory_help(subCategoryInvoice, map[string]*RestCommandCategory{}),
				subCategoryPayment: RESTCategory_help(subCategoryPayment, map[string]*RestCommandCategory{}),
				subCategoryPeer:    RESTCategory_help(subCategoryPeer, map[string]*RestCommandCategory{}),
			}),
			categoryMeta: RESTCategory_help(categoryMeta, map[string]*RestCommandCategory{}),
			categoryWallet: RESTCategory_help(categoryWallet, map[string]*RestCommandCategory{
				subCategoryNetworkStewardVote: RESTCategory_help(subCategoryNetworkStewardVote, map[string]*RestCommandCategory{}),
				subCategoryTransaction:        RESTCategory_help(subCategoryTransaction, map[string]*RestCommandCategory{}),
				subCategoryUnspent: RESTCategory_help(subCategoryUnspent, map[string]*RestCommandCategory{
					subSubCategoryLock: RESTCategory_help(subSubCategoryLock, map[string]*RestCommandCategory{}),
				}),
				subCategoryAddress: RESTCategory_help(subCategoryAddress, map[string]*RestCommandCategory{}),
			}),
			categoryNeutrino: RESTCategory_help(categoryNeutrino, map[string]*RestCommandCategory{}),
			categoryUtil: RESTCategory_help(categoryUtil, map[string]*RestCommandCategory{
				subCategorySeed: RESTCategory_help(subCategorySeed, map[string]*RestCommandCategory{}),
			}),
			categoryWatchtower: RESTCategory_help(categoryWatchtower, map[string]*RestCommandCategory{}),
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
