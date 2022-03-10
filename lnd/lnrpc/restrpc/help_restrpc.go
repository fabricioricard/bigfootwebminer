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
			"    getinfo                /help/meta/getinfo",
			"    getrecoveryinfo        /help/wallet/getrecoveryinfo",
			"    debuglevel             /help/meta/debuglevel",
			"    stop                   /help/meta/stop",
			"    version                /help/meta/version",
			"",
			"Channels:",
			"    openchannel            /help/channels/open",
			"    closechannel           /help/channels/close",
			"    closeallchannels       /help/channels/closeall",
			"    abandonchannel         /help/lightning/channel/abandon",
			"    channelbalance         /help/lightning/channel/balance",
			"    pendingchannels        /help/lightning/channel/pending",
			"    listchannels           /help/lightning/channel",
			"    closedchannels         /help/lightning/channel/closed",
			"    getnetworkinfo         /help/lightning/channel/networkinfo",
			"    feereport              /help/lightning/channel/feereport",
			"    updatechanpolicy       /help/lightning/channel/policy",
			"    exportchanbackup       /help/lightning/channel/backup/export",
			"    verifychanbackup       /help/lightning/channel/backup/verify",
			"    restorechanbackup      /help/lightning/channel/backup/restore",
			"",
			"Graph:",
			"    describegraph          /help/lightning/graph",
			"    getnodemetrics         /help/lightning/graph/nodemetrics",
			"    getchaninfo            /help/lightning/graph/channel",
			"    getnodeinfo            /help/lightning/graph/nodeinfo",
			"",
			"Invoices:",
			"    addinvoice             /help/lightning/invoice/create",
			"    lookupinvoice          /help/lightning/invoice/lookup",
			"    listinvoices           /help/lightning/invoice",
			"    decodepayreq           /help/lightning/invoice/decodepayreq",
			"",
			"On-chain:",
			"    estimatefee            /help/neutrino/estimatefee",
			"    sendmany               /help/wallet/sendmany",
			"    sendcoins              /help/wallet/sendcoins",
			"    listunspent            /help/wallet/unspent",
			"    listchaintxns          /help/lightning/gettransactions",
			"    setnetworkstewardvote  /help/wallet/setnetworkstewardvote",
			"    getnetworkstewardvote  /help/wallet/getnetworkstewardvote",
			"    bcasttransaction       /help/neutrino/bcasttransaction",
			"",
			"Payments:",
			"    sendpayment            /help/lightning/payment/send",
			"    payinvoice             /help/payment/payinvoice",
			"    sendtoroute            /help/lightning/payment/sendroutes",
			"    listpayments           /help/lightning/payment",
			"    queryroutes            /help/lightning/payment/queryroutes",
			"    fwdinghistory          /help/lightning/payment/fwdinghistory",
			"    trackpayment (deprecated?)",
			"    querymc                /help/lightning/payment/querymc",
			"    queryprob              /help/lightning/payment/queryprob",
			"    resetmc                /help/lightning/payment/resetmc",
			"    buildroute             /help/lightning/payment/buildroute",
			"",
			"Peers:",
			"    connect                /help/lightning/peers/connect",
			"    disconnect             /help/lightning/peers/disconnect",
			"    listpeers              /help/lightning/peers",
			"",
			"Startup:",
			"    create                 /help/wallet/create",
			"    unlock                 /help/wallet/unlock",
			"    changepassword         /help/wallet/password/change",
			"",
			"Wallet:",
			"    newaddress             /help/wallet/addreses/create",
			"    walletbalance          /help/lightning/walletbalance",
			"    getaddressbalances     /help/wallet/addresses/balances",
			"    signmessage            /help/wallet/addresses/signmessage",
			"    resync                 /help/wallet/resync",
			"    stopresync             /help/wallet/stopresync",
			"    genseed                /help/util/createseed",
			"    getwalletseed          /help/wallet/seed",
			"    getsecret              /help/wallet/getsecret",
			"    importprivkey          /help/wallet/addresses/import",
			"    listlockunspent        /help/wallet/lockunspent",
			"    lockunspent            /help/wallet/lockunspent/create",
			"    createtransaction      /help/wallet/transactions/create",
			"    dumpprivkey            /help/wallet/addresses/dumpprivkey",
			"    getnewaddress (same as newaddress?)",
			"    gettransaction         /help/wallet/transaction",
			"    sendfrom               /help/lightning/sendfrom",
			"",
			"Watchtower:",
			"    AddTower               /help",
		},
		Req: pkthelp.Type{},
		Res: pkthelp.Type{},
	}
}

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
