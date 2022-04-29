////////////////////////////////////////////////////////////////////////////////
//	restrpc/help/help.go  -  Apr-26-2022  -  aldebap
//
//	Data structs to build the help for REST APIs
////////////////////////////////////////////////////////////////////////////////

package help

import "github.com/pkt-cash/pktd/lnd/pkthelp"

const (
	URI_prefix = "/api/v1"
)

//	constants for categories and subcategories of commands
const (
	CategoryLightning             = "Lightning"
	SubcategoryChannel            = "Channel"
	SubSubCategoryBackup          = "Backup"
	SubCategoryGraph              = "Graph"
	SubCategoryInvoice            = "Invoice"
	SubCategoryPayment            = "Payment"
	SubCategoryPeer               = "Peer"
	CategoryMeta                  = "Meta"
	CategoryWallet                = "Wallet"
	SubCategoryNetworkStewardVote = "Network Steward Vote"
	SubCategoryTransaction        = "Transaction"
	SubCategoryUnspent            = "Unspent"
	SubSubCategoryLock            = "Lock"
	SubCategoryAddress            = "Address"
	CategoryNeutrino              = "Neutrino"
	CategoryUtil                  = "Util"
	SubCategorySeed               = "Seed"
	CategoryWatchtower            = "Watchtower"
)

//	mapping with the description for categories and subcategories of commands
var (
	CategoryDescription map[string][]string = map[string][]string{
		CategoryLightning:  {"The Lightning Network component of the wallet"},
		SubcategoryChannel: {"Management of lightning channels to direct peers of this pld node"},
		SubSubCategoryBackup: {"Backup and recovery of the state of active Lightning channels",
			"to and from this pld node"},
		SubCategoryGraph:   {"Information about the global known Lightning Network"},
		SubCategoryInvoice: {"Management of invoices which are used to request payment over Lightning"},
		SubCategoryPayment: {"Lightning network payments which have been made, or have been forwarded, through this node"},
		SubCategoryPeer:    {"Connections to other nodes in the Lightning Network"},
		CategoryMeta:       {"API endpoints which are relevant to the entire pld node, not any specific part"},
		CategoryWallet: {"APIs for management of on-chain (non-Lightning) payments,",
			"seed export and recovery, and on-chain transaction detection"},
		SubCategoryNetworkStewardVote: {"Control how this wallet votes on PKT Network Steward"},
		SubCategoryTransaction:        {"Create and manage on-chain transactions with the wallet"},
		SubCategoryUnspent:            {"Detected unspent transactions associated with one of our wallet addresses"},
		SubSubCategoryLock: {"Manipulation of unspent outputs which are 'locked'",
			"and therefore will not be used to source funds for any transaction"},
		SubCategoryAddress: {"Management of individual wallet addresses"},
		CategoryNeutrino:   {"Management of the Neutrino interface which is used to communicate with the p2p nodes in the network"},
		CategoryUtil:       {"Stateless utility functions which do not affect, not query, the node in any way"},
		SubCategorySeed:    {"Manipulation of mnemonic seed phrases which represent wallet keys"},
		CategoryWatchtower: {"Interact with the watchtower client"},
	}
)

//	constants for every pld command
const (
	//	lightning/channel subCategory commands
	CommandOpenChannel      = "openchannel"
	CommandCloseChannel     = "closechannel"
	CommandAbandonChannel   = "abandonchannel"
	CommandChannelBalance   = "channelbalance"
	CommandPendingChannels  = "pendingchannels"
	CommandListChannels     = "listchannels"
	CommandClosedChannels   = "closedchannels"
	CommandGetNetworkInfo   = "getnetworkinfo"
	CommandFeeReport        = "feereport"
	CommandUpdateChanPolicy = "updatechanpolicy"
	//	lightning/channel/backup subCategory commands
	CommandExportChanBackup  = "exportchanbackup"
	CommandVerifyChanBackup  = "verifychanbackup"
	CommandRestoreChanBackup = "restorechanbackup"
	//	lightning/graph subCategory commands
	CommandDescribeGraph  = "describegraph"
	CommandGetNodeMetrics = "getnodemetrics"
	CommandGetChanInfo    = "getchaninfo"
	CommandGetNodeInfo    = "getnodeinfo"
	//	lightning/invoice subCategory commands
	CommandAddInvoice    = "addinvoice"
	CommandLookupInvoice = "lookupinvoice"
	CommandListInvoices  = "listinvoices"
	CommandDecodePayreq  = "decodepayreq"
	//	lightning/payment subCategory command
	CommandSendPayment   = "sendpayment"
	CommandPayInvoice    = "payinvoice"
	CommandSendToRoute   = "sendtoroute"
	CommandListPayments  = "listpayments"
	CommandTrackPayment  = "trackpayment"
	CommandQueryRoutes   = "queryroutes"
	CommandFwdingHistory = "fwdinghistory"
	CommandQueryMc       = "querymc"
	CommandQueryProb     = "queryprob"
	CommandResetMc       = "resetmc"
	CommandBuildRoute    = "buildroute"
	//	lightning/peer subCategory command
	CommandConnectPeer    = "connectpeer"
	CommandDisconnectPeer = "disconnectpeer"
	CommandListPeers      = "listpeers"
	//	meta category command
	CommandDebugLevel = "debuglevel"
	CommandGetInfo    = "getinfo"
	CommandStop       = "stop"
	CommandVersion    = "version"
	CommandCrash      = "crash"
	//	wallet category command
	CommandWalletBalance    = "walletbalance"
	CommandChangePassphrase = "changePassphrase"
	CommandCheckPassphrase  = "checkPassphrase"
	CommandCreateWallet     = "createwallet"
	CommandGetSecret        = "getsecret"
	CommandGetWalletSeed    = "getwalletseed"
	CommandUnlockWallet     = "unlockwallet"
	//	wallet/networkstewardvote subCategory command
	CommandGetNetworkStewardVote = "getnetworkstewardvote"
	CommandSetNetworkStewardVote = "setnetworkstewardvote"
	//	wallet/transaction subCategory command
	CommandGetTransaction    = "gettransaction"
	CommandCreateTransaction = "createtransaction"
	CommandQueryTransactions = "querytransactions"
	CommandSendCoins         = "sendcoins"
	CommandSendFrom          = "sendfrom"
	CommandSendMany          = "sendmany"
	//	wallet/unspent subCategory command
	CommandListUnspent = "listunspent"
	CommandResync      = "resync"
	CommandStopResync  = "stopresync"
	//	wallet/unspent/lock subCategory command
	CommandListLockUnspent = "listlockunspent"
	CommandLockUnspent     = "lockunspent"
	//	wallet/address subCategory command
	CommandGetAddressBalances = "getaddressbalances"
	CommandNewAddress         = "newaddress"
	CommandDumpPrivkey        = "dumpprivkey"
	CommandImportPrivkey      = "importprivkey"
	CommandSignMessage        = "signmessage"
	//	neutrino category command
	CommandBcastTransaction = "bcasttransaction"
	CommandEstimateFee      = "estimatefee"
	//	util/seed subCategory command
	CommandChangeSeedPassphrase = "changeseedpassphrase"
	CommandGenSeed              = "genseed"
	//	wtclient/tower subCategory command
	CommandCreateWatchTower = "createwatchtower"
	CommandRemoveTower      = "removewatchtower"
	CommandListTowers       = "listtowers"
	CommandGetTowerInfo     = "gettowerinfo"
	CommandGetTowerStats    = "gettowerstats"
	CommandGetTowerPolicy   = "gettowerpolicy"
)

type CommandInfo struct {
	Category    string
	Description string
	Path        string
	HelpInfo    func() pkthelp.Method
}

//	mapping with the category, description and path for every command
var (
	CommandInfoData map[string]CommandInfo = map[string]CommandInfo{
		//	lightning/channel subCategory commands
		CommandOpenChannel: {
			Category:    SubcategoryChannel,
			Description: "Open a channel to a node or an existing peer",
			Path:        "/lightning/channel/open",
			HelpInfo:    pkthelp.Lightning_OpenChannel,
		},
		CommandCloseChannel: {
			Category:    SubcategoryChannel,
			Description: "Close an existing channel",
			Path:        "/lightning/channel/close",
			HelpInfo:    pkthelp.Lightning_CloseChannel,
		},
		CommandAbandonChannel: {
			Category:    SubcategoryChannel,
			Description: "Abandons an existing channel",
			Path:        "/lightning/channel/abandon",
			HelpInfo:    pkthelp.Lightning_AbandonChannel,
		},
		CommandChannelBalance: {
			Category:    SubcategoryChannel,
			Description: "Returns the sum of the total available channel balance across all open channels",
			Path:        "/lightning/channel/balance",
			HelpInfo:    pkthelp.Lightning_ChannelBalance,
		},
		CommandPendingChannels: {
			Category:    SubcategoryChannel,
			Description: "Display information pertaining to pending channels",
			Path:        "/lightning/channel/pending",
			HelpInfo:    pkthelp.Lightning_PendingChannels,
		},
		CommandListChannels: {
			Category:    SubcategoryChannel,
			Description: "List all open channels",
			Path:        "/lightning/channel",
			HelpInfo:    pkthelp.Lightning_ListChannels,
		},
		CommandClosedChannels: {
			Category:    SubcategoryChannel,
			Description: "List all closed channels",
			Path:        "/lightning/channel/closed",
			HelpInfo:    pkthelp.Lightning_ClosedChannels,
		},
		CommandGetNetworkInfo: {
			Category:    SubcategoryChannel,
			Description: "Get statistical information about the current state of the network",
			Path:        "/lightning/channel/networkinfo",
			HelpInfo:    pkthelp.Lightning_GetNetworkInfo,
		},
		CommandFeeReport: {
			Category:    SubcategoryChannel,
			Description: "Display the current fee policies of all active channels",
			Path:        "/lightning/channel/feereport",
			HelpInfo:    pkthelp.Lightning_FeeReport,
		},
		CommandUpdateChanPolicy: {
			Category:    SubcategoryChannel,
			Description: "Update the channel policy for all channels, or a single channel",
			Path:        "/lightning/channel/policy",
			HelpInfo:    pkthelp.Lightning_UpdateChannelPolicy,
		},
		//	lightning/channel/backup subCategory commands
		CommandExportChanBackup: {
			Category:    SubSubCategoryBackup,
			Description: "Obtain a static channel back up for a selected channels, or all known channels",
			Path:        "/lightning/channel/backup/export",
			HelpInfo:    pkthelp.Lightning_ExportChannelBackup,
		},
		CommandRestoreChanBackup: {
			Category:    SubSubCategoryBackup,
			Description: "Verify an existing channel backup",
			Path:        "/lightning/channel/backup/verify",
			HelpInfo:    pkthelp.Lightning_VerifyChanBackup,
		},
		CommandVerifyChanBackup: {
			Category:    SubSubCategoryBackup,
			Description: "Restore an existing single or multi-channel static channel backup",
			Path:        "/lightning/channel/backup/restore",
			HelpInfo:    pkthelp.Lightning_RestoreChannelBackups,
		},
		//	lightning/graph subCategory commands
		CommandDescribeGraph: {
			Category:    SubCategoryGraph,
			Description: "Describe the network graph",
			Path:        "/lightning/graph",
			HelpInfo:    pkthelp.Lightning_DescribeGraph,
		},
		CommandGetNodeMetrics: {
			Category:    SubCategoryGraph,
			Description: "Get node metrics",
			Path:        "/lightning/graph/nodemetrics",
			HelpInfo:    pkthelp.Lightning_GetNodeMetrics,
		},
		CommandGetChanInfo: {
			Category:    SubCategoryGraph,
			Description: "Get the state of a channel",
			Path:        "/lightning/graph/channel",
			HelpInfo:    pkthelp.Lightning_GetChanInfo,
		},
		CommandGetNodeInfo: {
			Category:    SubCategoryGraph,
			Description: "Get information on a specific node",
			Path:        "/lightning/graph/nodeinfo",
			HelpInfo:    pkthelp.Lightning_GetNodeInfo,
		},
		//	lightning/invoice subCategory commands
		CommandAddInvoice: {
			Category:    SubCategoryInvoice,
			Description: "Add a new invoice",
			Path:        "/lightning/invoice/create",
			HelpInfo:    pkthelp.Lightning_AddInvoice,
		},
		CommandLookupInvoice: {
			Category:    SubCategoryInvoice,
			Description: "Lookup an existing invoice by its payment hash",
			Path:        "/lightning/invoice/lookup",
			HelpInfo:    pkthelp.Lightning_LookupInvoice,
		},
		CommandListInvoices: {
			Category:    SubCategoryInvoice,
			Description: "List all invoices currently stored within the database. Any active debug invoices are ignored",
			Path:        "/lightning/invoice",
			HelpInfo:    pkthelp.Lightning_ListInvoices,
		},
		CommandDecodePayreq: {
			Category:    SubCategoryInvoice,
			Description: "Decode a payment request",
			Path:        "/lightning/invoice/decodepayreq",
			HelpInfo:    pkthelp.Lightning_DecodePayReq,
		},
		//	lightning/payment subCategory command
		CommandSendPayment: {
			Category:    SubCategoryPayment,
			Description: "Send a payment over lightning",
			Path:        "/lightning/payment/send",
			HelpInfo:    pkthelp.Lightning_SendPaymentSync,
		},
		CommandPayInvoice: {
			Category:    SubCategoryPayment,
			Description: "Pay an invoice over lightning",
			Path:        "/lightning/payment/payinvoice",
			HelpInfo:    nil,
		},
		CommandSendToRoute: {
			Category:    SubCategoryPayment,
			Description: "Send a payment over a predefined route",
			Path:        "/lightning/payment/sendtoroute",
			HelpInfo:    pkthelp.Lightning_SendToRouteSync,
		},
		CommandListPayments: {
			Category:    SubCategoryPayment,
			Description: "List all outgoing payments",
			Path:        "/lightning/payment",
			HelpInfo:    pkthelp.Lightning_ListPayments,
		},
		CommandTrackPayment: {
			Category: SubCategoryPayment,
			HelpInfo: nil,
		},
		CommandQueryRoutes: {
			Category:    SubCategoryPayment,
			Description: "Query a route to a destination",
			Path:        "/lightning/payment/queryroutes",
			HelpInfo:    pkthelp.Lightning_QueryRoutes,
		},
		CommandFwdingHistory: {
			Category:    SubCategoryPayment,
			Description: "Query the history of all forwarded HTLCs",
			Path:        "/lightning/payment/fwdinghistory",
			HelpInfo:    pkthelp.Lightning_ForwardingHistory,
		},
		CommandQueryMc: {
			Category:    SubCategoryPayment,
			Description: "Query the internal mission control state",
			Path:        "/lightning/payment/querymc",
			HelpInfo:    pkthelp.Router_QueryMissionControl,
		},
		CommandQueryProb: {
			Category:    SubCategoryPayment,
			Description: "Estimate a success probability",
			Path:        "/lightning/payment/queryprob",
			HelpInfo:    pkthelp.Router_QueryProbability,
		},
		CommandResetMc: {
			Category:    SubCategoryPayment,
			Description: "Reset internal mission control state",
			Path:        "/lightning/payment/resetmc",
			HelpInfo:    pkthelp.Router_ResetMissionControl,
		},
		CommandBuildRoute: {
			Category:    SubCategoryPayment,
			Description: "Build a route from a list of hop pubkeys",
			Path:        "/lightning/payment/buildroute",
			HelpInfo:    pkthelp.Router_BuildRoute,
		},
		//	lightning/peer subCategory command
		CommandConnectPeer: {
			Category:    SubCategoryPeer,
			Description: "Connect to a remote pld peer",
			Path:        "/lightning/peer/connect",
			HelpInfo:    pkthelp.Lightning_ConnectPeer,
		},
		CommandDisconnectPeer: {
			Category:    SubCategoryPeer,
			Description: "Disconnect a remote pld peer identified by public key",
			Path:        "/lightning/peer/disconnect",
			HelpInfo:    pkthelp.Lightning_DisconnectPeer,
		},
		CommandListPeers: {
			Category:    SubCategoryPeer,
			Description: "List all active, currently connected peers",
			Path:        "/lightning/peer",
			HelpInfo:    pkthelp.Lightning_ListPeers,
		},
		//	meta category command
		CommandDebugLevel: {
			Category:    CategoryMeta,
			Description: "Set the debug level",
			Path:        "/meta/debuglevel",
			HelpInfo:    pkthelp.Lightning_DebugLevel,
		},
		CommandGetInfo: {
			Category:    CategoryMeta,
			Description: "Returns basic information related to the active daemon",
			Path:        "/meta/getinfo",
			HelpInfo:    pkthelp.Lightning_GetInfo,
		},
		CommandStop: {
			Category:    CategoryMeta,
			Description: "Stop and shutdown the daemon",
			Path:        "/meta/stop",
			HelpInfo:    pkthelp.Lightning_StopDaemon,
		},
		CommandVersion: {
			Category:    CategoryMeta,
			Description: "Display pldctl and pld version info",
			Path:        "/meta/version",
			HelpInfo:    pkthelp.Versioner_GetVersion,
		},
		CommandCrash: {
			Category:    CategoryMeta,
			Description: "Force pld to crash (for debugging purposes)",
			Path:        "/meta/crash",
			HelpInfo:    pkthelp.MetaService_ForceCrash,
		},
		//	wallet category command
		CommandWalletBalance: {
			Category:    CategoryWallet,
			Description: "Compute and display the wallet's current balance",
			Path:        "/wallet/balance",
			HelpInfo:    pkthelp.Lightning_WalletBalance,
		},
		CommandChangePassphrase: {
			Category:    CategoryWallet,
			Description: "Change an encrypted wallet's password at startup",
			Path:        "/wallet/changepassphrase",
			HelpInfo:    pkthelp.MetaService_ChangePassword,
		},
		CommandCheckPassphrase: {
			Category:    CategoryWallet,
			Description: "Check the wallet's password",
			Path:        "/wallet/checkpassphrase",
			HelpInfo:    pkthelp.MetaService_CheckPassword,
		},
		CommandCreateWallet: {
			Category:    CategoryWallet,
			Description: "Initialize a wallet when starting lnd for the first time",
			Path:        "/wallet/create",
			HelpInfo:    pkthelp.WalletUnlocker_InitWallet,
		},
		CommandGetSecret: {
			Category:    CategoryWallet,
			Description: "Get a secret seed",
			Path:        "/wallet/getsecret",
			HelpInfo:    pkthelp.Lightning_GetSecret,
		},
		CommandGetWalletSeed: {
			Category:    CategoryWallet,
			Description: "Get the wallet seed words for this wallet",
			Path:        "/wallet/seed",
			HelpInfo:    pkthelp.Lightning_GetWalletSeed,
		},
		CommandUnlockWallet: {
			Category:    CategoryWallet,
			Description: "Unlock an encrypted wallet at startup",
			Path:        "/wallet/unlock",
			HelpInfo:    pkthelp.WalletUnlocker_UnlockWallet,
		},
		//	wallet/networkstewardvote subCategory command
		CommandGetNetworkStewardVote: {
			Category:    SubCategoryNetworkStewardVote,
			Description: "Find out how the wallet is currently configured to vote in a network steward election",
			Path:        "/wallet/networkstewardvote",
			HelpInfo:    pkthelp.Lightning_GetNetworkStewardVote,
		},
		CommandSetNetworkStewardVote: {
			Category:    SubCategoryNetworkStewardVote,
			Description: "Configure the wallet to vote for a network steward when making payments (note: payments to segwit addresses cannot vote)",
			Path:        "/wallet/networkstewardvote/set",
			HelpInfo:    pkthelp.Lightning_SetNetworkStewardVote,
		},
		//	wallet/transaction subCategory command
		CommandGetTransaction: {
			Category:    SubCategoryTransaction,
			Description: "Returns a JSON object with details regarding a transaction relevant to this wallet",
			Path:        "/wallet/transaction",
			HelpInfo:    pkthelp.Lightning_GetTransaction,
		},
		CommandCreateTransaction: {
			Category:    SubCategoryTransaction,
			Description: "Create a transaction but do not send it to the chain",
			Path:        "/wallet/transaction/create",
			HelpInfo:    pkthelp.Lightning_CreateTransaction,
		},
		CommandQueryTransactions: {
			Category:    SubCategoryTransaction,
			Description: "List transactions from the wallet",
			Path:        "/wallet/transaction/query",
			HelpInfo:    pkthelp.Lightning_GetTransactions,
		},
		CommandSendCoins: {
			Category:    SubCategoryTransaction,
			Description: "Send bitcoin on-chain to an address",
			Path:        "/wallet/transaction/sendcoins",
			HelpInfo:    pkthelp.Lightning_SendCoins,
		},
		CommandSendFrom: {
			Category:    SubCategoryTransaction,
			Description: "Authors, signs, and sends a transaction that outputs some amount to a payment address",
			Path:        "/wallet/transaction/sendfrom",
			HelpInfo:    pkthelp.Lightning_SendFrom,
		},
		CommandSendMany: {
			Category:    SubCategoryTransaction,
			Description: "Send bitcoin on-chain to multiple addresses",
			Path:        "/wallet/transaction/sendmany",
			HelpInfo:    pkthelp.Lightning_SendMany,
		},
		//	wallet/unspent subCategory command
		CommandListUnspent: {
			Category:    SubCategoryUnspent,
			Description: "List utxos available for spending",
			Path:        "/wallet/unspent",
			HelpInfo:    pkthelp.Lightning_ListUnspent,
		},
		CommandResync: {
			Category:    SubCategoryUnspent,
			Description: "Scan over the chain to find any transactions which may not have been recorded in the wallet's database",
			Path:        "/wallet/unspent/resync",
			HelpInfo:    pkthelp.Lightning_ReSync,
		},
		CommandStopResync: {
			Category:    SubCategoryUnspent,
			Description: "Stop a re-synchronization job before it's completion",
			Path:        "/wallet/unspent/stopresync",
			HelpInfo:    pkthelp.Lightning_StopReSync,
		},
		//	wallet/unspent/lock subCategory command
		CommandListLockUnspent: {
			Category:    SubSubCategoryLock,
			Description: "Returns a JSON array of outpoints marked as locked (with lockunspent) for this wallet session",
			Path:        "/wallet/unspent/lock",
			HelpInfo:    pkthelp.Lightning_ListLockUnspent,
		},
		CommandLockUnspent: {
			Category:    SubSubCategoryLock,
			Description: "Locks or unlocks an unspent output",
			Path:        "/wallet/unspent/lock/create", // TODO: /wallet/unspent/lock/delete
			HelpInfo:    pkthelp.Lightning_LockUnspent,
		},
		//	wallet/address subCategory command
		CommandGetAddressBalances: {
			Category:    SubCategoryAddress,
			Description: "Compute and display balances for each address in the wallet",
			Path:        "/wallet/address/balances",
			HelpInfo:    pkthelp.Lightning_GetAddressBalances,
		},
		CommandNewAddress: {
			Category:    SubCategoryAddress,
			Description: "Generates a new address",
			Path:        "/wallet/address/create",
			HelpInfo:    pkthelp.Lightning_GetNewAddress,
		},
		CommandDumpPrivkey: {
			Category:    SubCategoryAddress,
			Description: "Returns the private key in WIF encoding that controls some wallet address",
			Path:        "/wallet/address/dumpprivkey",
			HelpInfo:    pkthelp.Lightning_DumpPrivKey,
		},
		CommandImportPrivkey: {
			Category:    SubCategoryAddress,
			Description: "Imports a WIF-encoded private key to the 'imported' account",
			Path:        "/wallet/address/import",
			HelpInfo:    pkthelp.Lightning_ImportPrivKey,
		},
		CommandSignMessage: {
			Category:    SubCategoryAddress,
			Description: "Signs a message using the private key of a payment address",
			Path:        "/wallet/address/signmessage",
			HelpInfo:    pkthelp.Signer_SignMessage,
		},
		//	neutrino category command
		CommandBcastTransaction: {
			Category:    CategoryNeutrino,
			Description: "Broadcast a transaction onchain",
			Path:        "/neutrino/bcasttransaction",
			HelpInfo:    pkthelp.Lightning_BcastTransaction,
		},
		CommandEstimateFee: {
			Category:    CategoryNeutrino,
			Description: "Get fee estimates for sending bitcoin on-chain to multiple addresses",
			Path:        "/neutrino/estimatefee",
			HelpInfo:    pkthelp.Lightning_EstimateFee,
		},
		//	util/seed subCategory command
		CommandChangeSeedPassphrase: {
			Category:    SubCategorySeed,
			Description: "Alter the passphrase which is used to encrypt a wallet seed",
			Path:        "/util/seed/changepassphrase",
			HelpInfo:    pkthelp.Lightning_ChangeSeedPassphrase,
		},
		CommandGenSeed: {
			Category:    SubCategorySeed,
			Description: "Create a secret seed",
			Path:        "/util/seed/create",
			HelpInfo:    pkthelp.WalletUnlocker_GenSeed,
		},
		//	wtclient/tower subCategory command
		CommandCreateWatchTower: {
			Category:    CategoryWatchtower,
			Description: "Register a watchtower to use for future sessions/backups",
			Path:        "/wtclient/tower/create",
			HelpInfo:    pkthelp.WatchtowerClient_AddTower,
		},
		CommandRemoveTower: {
			Category:    CategoryWatchtower,
			Description: "Remove a watchtower to prevent its use for future sessions/backups",
			Path:        "/wtclient/tower/remove",
			HelpInfo:    pkthelp.WatchtowerClient_RemoveTower,
		},
		CommandListTowers: {
			Category:    CategoryWatchtower,
			Description: "Display information about all registered watchtowers",
			Path:        "/wtclient/tower",
			HelpInfo:    pkthelp.WatchtowerClient_ListTowers,
		},
		CommandGetTowerInfo: {
			Category:    CategoryWatchtower,
			Description: "Display information about a specific registered watchtower",
			Path:        "/wtclient/tower/getinfo",
			HelpInfo:    pkthelp.WatchtowerClient_GetTowerInfo,
		},
		CommandGetTowerStats: {
			Category:    CategoryWatchtower,
			Description: "Display the session stats of the watchtower client",
			Path:        "/wtclient/tower/stats",
			HelpInfo:    pkthelp.WatchtowerClient_Stats,
		},
		CommandGetTowerPolicy: {
			Category:    CategoryWatchtower,
			Description: "Display the active watchtower client policy configuration",
			Path:        "/wtclient/tower/policy",
			HelpInfo:    pkthelp.WatchtowerClient_Policy,
		},
	}
)

//	the category help in REST master
func RESTCategory_help(category string, subCategory map[string]*RestCommandCategory) *RestCommandCategory {
	restCommandCategory := &RestCommandCategory{
		Description: CategoryDescription[category],
		Endpoints:   make(map[string]string),
		Subcategory: make(map[string]*RestCommandCategory),
	}

	//	add all endpoints for the category
	for _, commandInfo := range CommandInfoData {

		if commandInfo.Category == category {
			restCommandCategory.Endpoints[URI_prefix+commandInfo.Path] = commandInfo.Description
		}
	}

	//	add all sub categories
	for name, value := range subCategory {
		restCommandCategory.Subcategory[name] = value
	}

	return restCommandCategory
}

//	return the REST master help messsage
func RESTMaster_help() *RestMasterHelpResponse {
	masterHelpResp := &RestMasterHelpResponse{
		Name: "pld - Lightning Network Daemon REST interface (pld)",
		Description: []string{
			"General information about PLD",
		},
		Category: map[string]*RestCommandCategory{
			CategoryLightning: RESTCategory_help(CategoryLightning, map[string]*RestCommandCategory{
				SubcategoryChannel: RESTCategory_help(SubcategoryChannel, map[string]*RestCommandCategory{
					SubSubCategoryBackup: RESTCategory_help(SubSubCategoryBackup, map[string]*RestCommandCategory{}),
				}),
				SubCategoryGraph:   RESTCategory_help(SubCategoryGraph, map[string]*RestCommandCategory{}),
				SubCategoryInvoice: RESTCategory_help(SubCategoryInvoice, map[string]*RestCommandCategory{}),
				SubCategoryPayment: RESTCategory_help(SubCategoryPayment, map[string]*RestCommandCategory{}),
				SubCategoryPeer:    RESTCategory_help(SubCategoryPeer, map[string]*RestCommandCategory{}),
			}),
			CategoryMeta: RESTCategory_help(CategoryMeta, map[string]*RestCommandCategory{}),
			CategoryWallet: RESTCategory_help(CategoryWallet, map[string]*RestCommandCategory{
				SubCategoryNetworkStewardVote: RESTCategory_help(SubCategoryNetworkStewardVote, map[string]*RestCommandCategory{}),
				SubCategoryTransaction:        RESTCategory_help(SubCategoryTransaction, map[string]*RestCommandCategory{}),
				SubCategoryUnspent: RESTCategory_help(SubCategoryUnspent, map[string]*RestCommandCategory{
					SubSubCategoryLock: RESTCategory_help(SubSubCategoryLock, map[string]*RestCommandCategory{}),
				}),
				SubCategoryAddress: RESTCategory_help(SubCategoryAddress, map[string]*RestCommandCategory{}),
			}),
			CategoryNeutrino: RESTCategory_help(CategoryNeutrino, map[string]*RestCommandCategory{}),
			CategoryUtil: RESTCategory_help(CategoryUtil, map[string]*RestCommandCategory{
				SubCategorySeed: RESTCategory_help(SubCategorySeed, map[string]*RestCommandCategory{}),
			}),
			CategoryWatchtower: RESTCategory_help(CategoryWatchtower, map[string]*RestCommandCategory{}),
		},
	}

	return masterHelpResp
}
