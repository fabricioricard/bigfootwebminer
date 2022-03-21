package restrpc

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/mux"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkt-cash/pktd/btcjson"
	"github.com/pkt-cash/pktd/btcutil/er"
	"github.com/pkt-cash/pktd/connmgr/banmgr"
	"github.com/pkt-cash/pktd/lnd/chainreg"
	"github.com/pkt-cash/pktd/lnd/lnrpc"
	"github.com/pkt-cash/pktd/lnd/lnrpc/routerrpc"
	"github.com/pkt-cash/pktd/lnd/lnrpc/verrpc"
	"github.com/pkt-cash/pktd/lnd/lnrpc/wtclientrpc"
	"github.com/pkt-cash/pktd/lnd/pkthelp"
	"github.com/pkt-cash/pktd/lnd/walletunlocker"
	"github.com/pkt-cash/pktd/neutrino"
	"github.com/pkt-cash/pktd/pktlog/log"
	"github.com/pkt-cash/pktd/pktwallet/waddrmgr"
	"github.com/pkt-cash/pktd/pktwallet/wallet"
)

const (
	URI_prefix = "/api/v1"
)

const (
	categoryLightning             = "Lightning"
	subcategoryChannel            = "Channel"
	subSubCategoryBackup          = "Backup"
	subCategoryGraph              = "Graph"
	subCategoryInvoice            = "Invoice"
	subCategoryPayment            = "Payment"
	subCategoryPeer               = "Peer"
	categoryMeta                  = "Meta"
	categoryWallet                = "Wallet"
	subCategoryNetworkStewardVote = "Network Steward Vote"
	subCategoryTransaction        = "Transaction"
	subCategoryUnspent            = "Unspent"
	subSubCategoryLock            = "Lock"
	subCategoryAddress            = "Address"
	categoryNeutrino              = "Neutrino"
	categoryUtil                  = "Util"
	subCategorySeed               = "Seed"
	categoryWatchtower            = "Watchtower"
)

type RpcFunc struct {
	category    string
	description string
	path        string
	req         proto.Message
	res         proto.Message
	f           func(c *RpcContext, m proto.Message) (proto.Message, er.R)

	getHelpInfo func() pkthelp.Method
}

var rpcFunctions []RpcFunc = []RpcFunc{
	//WalletUnlocker: Wallet unlock
	//Will try to unlock the wallet with the password(s) provided
	{
		category:    categoryWallet,
		description: "Unlock an encrypted wallet at startup",

		path: "/wallet/unlock",
		req:  (*lnrpc.UnlockWalletRequest)(nil),
		res:  nil,
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {
			req, ok := m.(*lnrpc.UnlockWalletRequest)
			if !ok {
				return nil, er.New("Argument is not a UnlockWalletRequest")
			}
			if u, err := c.withUnlocker(); err != nil {
				return nil, err
			} else if _, err := u.UnlockWallet0(context.TODO(), req); err != nil {
				return nil, err
			}
			return nil, nil
		},

		getHelpInfo: pkthelp.WalletUnlocker_UnlockWallet,
	},
	//WalletUnlocker: Wallet create
	//Will try to create/restore wallet
	{
		category:    categoryWallet,
		description: "Initialize a wallet when starting lnd for the first time",

		path: "/wallet/create",
		req:  (*lnrpc.InitWalletRequest)(nil), // Use init wallet structure to create
		res:  nil,

		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			initWalletReq, ok := m.(*lnrpc.InitWalletRequest)
			if !ok {
				return nil, er.New("Argument is not a InitWalletRequest")
			}

			//	init wallet
			cc, errr := c.withUnlocker()
			if errr != nil {
				return nil, errr
			}

			_, err := cc.InitWallet(context.TODO(), initWalletReq)
			if err != nil {
				return nil, er.E(err)
			}
			return nil, nil
		},

		getHelpInfo: pkthelp.WalletUnlocker_InitWallet,
	},
	//MetaService get info
	{
		category:    categoryMeta,
		description: "Returns basic information related to the active daemon",

		path: "/meta/getinfo",
		req:  nil,
		res:  (*lnrpc.GetInfo2Response)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			var ni lnrpc.NeutrinoInfo
			if n, _ := c.withNeutrino(); n != nil {
				neutrinoPeers := n.Peers()
				for i := range neutrinoPeers {
					var peerDesc lnrpc.PeerDesc
					neutrinoPeer := neutrinoPeers[i]

					peerDesc.BytesReceived = neutrinoPeer.BytesReceived()
					peerDesc.BytesSent = neutrinoPeer.BytesSent()
					peerDesc.LastRecv = neutrinoPeer.LastRecv().String()
					peerDesc.LastSend = neutrinoPeer.LastSend().String()
					peerDesc.Connected = neutrinoPeer.Connected()
					peerDesc.Addr = neutrinoPeer.Addr()
					peerDesc.Inbound = neutrinoPeer.Inbound()
					na := neutrinoPeer.NA()
					if na != nil {
						peerDesc.Na = na.IP.String() + ":" + strconv.Itoa(int(na.Port))
					}
					peerDesc.Id = neutrinoPeer.ID()
					peerDesc.UserAgent = neutrinoPeer.UserAgent()
					peerDesc.Services = neutrinoPeer.Services().String()
					peerDesc.VersionKnown = neutrinoPeer.VersionKnown()
					peerDesc.AdvertisedProtoVer = neutrinoPeer.Describe().AdvertisedProtoVer
					peerDesc.ProtocolVersion = neutrinoPeer.ProtocolVersion()
					peerDesc.SendHeadersPreferred = neutrinoPeer.Describe().SendHeadersPreferred
					peerDesc.VerAckReceived = neutrinoPeer.VerAckReceived()
					peerDesc.WitnessEnabled = neutrinoPeer.Describe().WitnessEnabled
					peerDesc.WireEncoding = strconv.Itoa(int(neutrinoPeer.Describe().WireEncoding))
					peerDesc.TimeOffset = neutrinoPeer.TimeOffset()
					peerDesc.TimeConnected = neutrinoPeer.Describe().TimeConnected.String()
					peerDesc.StartingHeight = neutrinoPeer.StartingHeight()
					peerDesc.LastBlock = neutrinoPeer.LastBlock()
					if neutrinoPeer.LastAnnouncedBlock() != nil {
						peerDesc.LastAnnouncedBlock = neutrinoPeer.LastAnnouncedBlock().CloneBytes()
					}
					peerDesc.LastPingNonce = neutrinoPeer.LastPingNonce()
					peerDesc.LastPingTime = neutrinoPeer.LastPingTime().String()
					peerDesc.LastPingMicros = neutrinoPeer.LastPingMicros()

					ni.Peers = append(ni.Peers, &peerDesc)
				}
				n.BanMgr().ForEachIp(func(bi banmgr.BanInfo) er.R {
					ban := lnrpc.NeutrinoBan{}
					ban.Addr = bi.Addr
					ban.Reason = bi.Reason
					ban.EndTime = bi.BanExpiresTime.String()
					ban.BanScore = bi.BanScore

					ni.Bans = append(ni.Bans, &ban)
					return nil
				})

				neutrionoQueries := n.GetActiveQueries()
				for i := range neutrionoQueries {
					nq := lnrpc.NeutrinoQuery{}
					query := neutrionoQueries[i]
					if query.Peer != nil {
						nq.Peer = query.Peer.String()
					} else {
						nq.Peer = "<nil>"
					}
					nq.Command = query.Command
					nq.ReqNum = query.ReqNum
					nq.CreateTime = query.CreateTime
					nq.LastRequestTime = query.LastRequestTime
					nq.LastResponseTime = query.LastResponseTime

					ni.Queries = append(ni.Queries, &nq)
				}

				bb, err := n.BestBlock()
				if err != nil {
					return nil, err
				}
				ni.BlockHash = bb.Hash.String()
				ni.Height = bb.Height
				ni.BlockTimestamp = bb.Timestamp.String()
				ni.IsSyncing = !n.IsCurrent()
			}

			var walletInfo *lnrpc.WalletInfo
			if w, _ := c.withWallet(); w != nil {
				mgrStamp := w.Manager.SyncedTo()
				walletStats := &lnrpc.WalletStats{}
				w.ReadStats(func(ws *btcjson.WalletStats) {
					walletStats.MaintenanceInProgress = ws.MaintenanceInProgress
					walletStats.MaintenanceName = ws.MaintenanceName
					walletStats.MaintenanceCycles = int32(ws.MaintenanceCycles)
					walletStats.MaintenanceLastBlockVisited = int32(ws.MaintenanceLastBlockVisited)
					walletStats.Syncing = ws.Syncing
					if ws.SyncStarted != nil {
						walletStats.SyncStarted = ws.SyncStarted.String()
					}
					walletStats.SyncRemainingSeconds = ws.SyncRemainingSeconds
					walletStats.SyncCurrentBlock = ws.SyncCurrentBlock
					walletStats.SyncFrom = ws.SyncFrom
					walletStats.SyncTo = ws.SyncTo
					walletStats.BirthdayBlock = ws.BirthdayBlock
				})
				walletInfo = &lnrpc.WalletInfo{
					CurrentBlockHash:      mgrStamp.Hash.String(),
					CurrentHeight:         mgrStamp.Height,
					CurrentBlockTimestamp: mgrStamp.Timestamp.String(),
					WalletVersion:         int32(waddrmgr.LatestMgrVersion),
					WalletStats:           walletStats,
				}
			}

			// Get Lightning info
			var lightning *lnrpc.GetInfoResponse
			if cc, _ := c.withRpcServer(); cc != nil {
				if l, err := cc.GetInfo(context.TODO(), nil); err != nil {
					return nil, er.E(err)
				} else {
					lightning = l
				}
			}

			return &lnrpc.GetInfo2Response{
				Neutrino:  &ni,
				Wallet:    walletInfo,
				Lightning: lightning,
			}, nil
		},

		getHelpInfo: pkthelp.Lightning_GetInfo,
	},
	//MetaService change wallet password
	{
		category:    categoryWallet,
		description: "Change an encrypted wallet's password at startup",

		path: "/wallet/changepassphrase",
		req:  (*lnrpc.ChangePasswordRequest)(nil),
		res:  nil, // no response
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {
			req, ok := m.(*lnrpc.ChangePasswordRequest)
			if !ok {
				return nil, er.New("Argument is not a ChangePasswordRequest")
			}
			meta, err := c.withMetaServer()
			if err != nil {
				return nil, err
			}
			resp, errr := meta.ChangePassword(context.TODO(), req)
			if errr != nil {
				return nil, er.E(errr)
			}
			return resp, nil
		},

		getHelpInfo: pkthelp.MetaService_ChangePassword,
	},
	//Wallet balance
	//requires unlocked wallet -> access to rpcServer
	{
		category:    categoryWallet,
		description: "Compute and display the wallet's current balance",

		path: "/wallet/balance",
		req:  nil,
		res:  (*lnrpc.WalletBalanceResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {
			if server, err := c.withRpcServer(); server != nil {
				if l, err := server.WalletBalance(context.TODO(), nil); err != nil {
					return nil, er.E(err)
				} else {
					return l, nil
				}
			} else {
				return nil, err
			}
		},

		getHelpInfo: pkthelp.Lightning_WalletBalance,
	},
	//Wallet transactions
	//requires unlocked wallet -> access to rpcServer
	{
		category:    subCategoryTransaction,
		description: "List transactions from the wallet",

		path: "/wallet/transaction/query",
		req:  (*lnrpc.GetTransactionsRequest)(nil),
		res:  (*lnrpc.TransactionDetails)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {
			req, ok := m.(*lnrpc.GetTransactionsRequest)
			if !ok {
				return nil, er.New("Argument is not a GetTransactionsRequest")
			}
			if server, err := c.withRpcServer(); server != nil {
				if l, err := server.GetTransactions(context.TODO(), req); err != nil {
					return nil, er.E(err)
				} else {
					return l, nil
				}
			} else {
				return nil, err
			}
		},

		getHelpInfo: pkthelp.Lightning_GetTransactions,
	},
	//New wallet address
	//requires unlocked wallet -> access to rpcServer
	{
		category:    subCategoryAddress,
		description: "Generates a new address",

		path: "/wallet/address/create",
		req:  (*lnrpc.GetNewAddressRequest)(nil),
		res:  (*lnrpc.GetNewAddressResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {
			req, ok := m.(*lnrpc.GetNewAddressRequest)
			if !ok {
				return nil, er.New("Argument is not a GetNewAddressRequest")
			}
			if server, err := c.withRpcServer(); server != nil {
				if l, err := server.GetNewAddress(context.TODO(), req); err != nil {
					return nil, er.E(err)
				} else {
					return l, nil
				}
			} else {
				return nil, err
			}
		},

		getHelpInfo: pkthelp.Lightning_GetNewAddress,
	},
	//GetAddressBalances
	{
		category:    subCategoryAddress,
		description: "Compute and display balances for each address in the wallet",

		path: "/wallet/address/balances",
		req:  (*lnrpc.GetAddressBalancesRequest)(nil),
		res:  (*lnrpc.GetAddressBalancesResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {
			req, ok := m.(*lnrpc.GetAddressBalancesRequest)
			if !ok {
				return nil, er.New("Argument is not a GetAddressBalancesRequest")
			}
			if server, err := c.withRpcServer(); server != nil {
				if l, err := server.GetAddressBalances(context.TODO(), req); err != nil {
					return nil, er.E(err)
				} else {
					return l, nil
				}
			} else {
				return nil, err
			}
		},

		getHelpInfo: pkthelp.Lightning_GetAddressBalances,
	},
	//Sendfrom
	{
		category:    subCategoryTransaction,
		description: "Authors, signs, and sends a transaction that outputs some amount to a payment address",

		path: "/wallet/transaction/sendfrom",
		req:  (*lnrpc.SendFromRequest)(nil),
		res:  (*lnrpc.SendFromResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {
			req, ok := m.(*lnrpc.SendFromRequest)
			if !ok {
				return nil, er.New("Argument is not a SendFromRequest")
			}
			if server, err := c.withRpcServer(); server != nil {
				if l, err := server.SendFrom(context.TODO(), req); err != nil {
					return nil, er.E(err)
				} else {
					return l, nil
				}
			} else {
				return nil, err
			}
		},

		getHelpInfo: pkthelp.Lightning_SendFrom,
	},
	//GetWalletSeed
	{
		category:    categoryWallet,
		description: "Get the wallet seed words for this wallet",

		path: "/wallet/seed",
		req:  nil,
		res:  (*lnrpc.GetWalletSeedResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {
			if server, err := c.withRpcServer(); server != nil {
				if l, err := server.GetWalletSeed(context.TODO(), nil); err != nil {
					return nil, er.E(err)
				} else {
					return l, nil
				}
			} else {
				return nil, err
			}
		},

		getHelpInfo: pkthelp.Lightning_GetWalletSeed,
	},
	//GetTransaction
	{
		category:    subCategoryTransaction,
		description: "Returns a JSON object with details regarding a transaction relevant to this wallet",

		path: "/wallet/transaction",
		req:  (*lnrpc.GetTransactionRequest)(nil),
		res:  (*lnrpc.GetTransactionResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {
			req, ok := m.(*lnrpc.GetTransactionRequest)
			if !ok {
				return nil, er.New("Argument is not a GetTransactionRequest")
			}
			if server, err := c.withRpcServer(); server != nil {
				if l, err := server.GetTransaction(context.TODO(), req); err != nil {
					return nil, er.E(err)
				} else {
					return l, nil
				}
			} else {
				return nil, err
			}
		},

		getHelpInfo: pkthelp.Lightning_GetTransaction,
	},
	//Resync
	{
		category:    subCategoryUnspent,
		description: "Scan over the chain to find any transactions which may not have been recorded in the wallet's database",

		path: "/wallet/unspent/resync",
		req:  (*lnrpc.ReSyncChainRequest)(nil),
		res:  nil,
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {
			req, ok := m.(*lnrpc.ReSyncChainRequest)
			if !ok {
				return nil, er.New("Argument is not a ReSyncChainRequest")
			}
			if server, err := c.withRpcServer(); server != nil {
				if l, err := server.ReSync(context.TODO(), req); err != nil {
					return nil, er.E(err)
				} else {
					return l, nil
				}
			} else {
				return nil, err
			}
		},

		getHelpInfo: pkthelp.Lightning_ReSync,
	},
	//StopResync
	{
		category:    subCategoryUnspent,
		description: "Stop a re-synchronization job before it's completion",

		path: "/wallet/unspent/stopresync",
		req:  nil,
		res:  (*lnrpc.StopReSyncResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {
			if server, err := c.withRpcServer(); server != nil {
				if l, err := server.StopReSync(context.TODO(), nil); err != nil {
					return nil, er.E(err)
				} else {
					return l, nil
				}
			} else {
				return nil, err
			}
		},

		getHelpInfo: pkthelp.Lightning_StopReSync,
	},
	//	GenSeed service
	{
		category:    subCategorySeed,
		description: "Create a secret seed",

		path: "/util/seed/create",
		req:  (*lnrpc.GenSeedRequest)(nil),
		res:  (*lnrpc.GenSeedResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			genSeedReq, ok := m.(*lnrpc.GenSeedRequest)
			if !ok {
				return nil, er.New("Argument is not a GenSeedRequest")
			}

			//	generate a new seed
			cc, errr := c.withUnlocker()
			if cc != nil {
				var genSeedResp *lnrpc.GenSeedResponse

				genSeedResp, err := cc.GenSeed(context.TODO(), genSeedReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return genSeedResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.WalletUnlocker_GenSeed,
	},
	//	TODO: Change Passphrase service
	{
		category:    subCategorySeed,
		description: "Alter the passphrase which is used to encrypt a wallet seed",

		path: "/util/seed/changepassphrase",
		req:  nil,
		res:  nil,
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			return nil, nil
		},

		getHelpInfo: pkthelp.WalletUnlocker_GenSeed,
	},
	//	service debug level
	{
		category:    categoryMeta,
		description: "Set the debug level",

		path: "/meta/debuglevel",
		req:  (*lnrpc.DebugLevelRequest)(nil),
		res:  (*lnrpc.DebugLevelResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			debugLevelReq, ok := m.(*lnrpc.DebugLevelRequest)
			if !ok {
				return nil, er.New("Argument is not a DebugLevelRequest")
			}

			//	set Lightning debug level
			cc, errr := c.withRpcServer()
			if cc != nil {
				var debugLevelResp *lnrpc.DebugLevelResponse

				debugLevelResp, err := cc.DebugLevel(context.TODO(), debugLevelReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return debugLevelResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_DebugLevel,
	},
	//	service to stop the pld daemon
	{
		category:    categoryMeta,
		description: "Stop and shutdown the daemon",

		path: "/meta/stop",
		req:  nil,
		res:  nil,
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	invoke Lightning stop daemon command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var stopResp *lnrpc.StopResponse

				stopResp, err := cc.StopDaemon(context.TODO(), nil)
				if err != nil {
					return nil, er.E(err)
				} else {
					return stopResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_StopDaemon,
	},
	//	service daemon version
	{
		category:    categoryMeta,
		description: "Display pldctl and pld version info",

		path: "/meta/version",
		req:  nil,
		res:  (*verrpc.Version)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get daemon version
			cc, errr := c.withVerRPCServer()
			if cc != nil {
				var versionResp *verrpc.Version

				versionResp, err := cc.GetVersion(context.TODO(), nil)
				if err != nil {
					return nil, er.E(err)
				} else {
					return versionResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Versioner_GetVersion,
	},
	//	TODO: service openchannel
	{
		category:    subcategoryChannel,
		description: "Open a channel to a node or an existing peer",

		path: "/lightning/channel/open",
		req:  (*lnrpc.OpenChannelRequest)(nil),
		res:  (*lnrpc.ChannelPoint)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			openChannelReq, ok := m.(*lnrpc.OpenChannelRequest)
			if !ok {
				return nil, er.New("Argument is not a OpenChannelRequest")
			}

			//	open a channel
			cc, errr := c.withRpcServer()
			if cc != nil {
				var openChannelResp *lnrpc.ChannelPoint

				openChannelResp, err := cc.OpenChannelSync(context.TODO(), openChannelReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return openChannelResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_OpenChannel,
	},
	//	TODO: service closechannel
	{
		category:    subcategoryChannel,
		description: "Close an existing channel",

		path: "/lightning/channel/close",
		req:  (*lnrpc.CloseChannelRequest)(nil),
		res:  nil,
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			closeChannelReq, ok := m.(*lnrpc.CloseChannelRequest)
			if !ok {
				return nil, er.New("Argument is not a CloseChannelRequest")
			}

			//	close a channel
			cc, errr := c.withRpcServer()
			if cc != nil {
				err := cc.CloseChannel(closeChannelReq, nil)
				if err != nil {
					return nil, er.E(err)
				} else {
					return nil, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_CloseChannel,
	},
	//	TODO: service closeallchannels
	//	check with Dimitris because the CloseAllChannels calls listChannels and then close one by one. This is done in the client ide (pldctl)
	/*
		{
			category:    categoryChannels,
			description: "Close all existing channels",

			path: "/channels/closeall",
			req:  nil,
			res:  (*lnrpc.GetRecoveryInfoResponse)(nil),
			f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

				//	close all channels
				cc, errr := c.withRpcServer()
				if cc != nil {
					var recoveryInfo *lnrpc.GetRecoveryInfoResponse

					recoveryInfo, err := cc.GetRecoveryInfo(context.TODO(), nil)
					if err != nil {
						return nil, er.E(err)
					} else {
						return recoveryInfo, nil
					}
				} else {
					return nil, errr
				}
			},
		},
	*/
	//	service abandonchannel
	{
		category:    subcategoryChannel,
		description: "Abandons an existing channel",

		path: "/lightning/channel/abandon",
		req:  (*lnrpc.AbandonChannelRequest)(nil),
		res:  (*lnrpc.AbandonChannelResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			abandonChannelReq, ok := m.(*lnrpc.AbandonChannelRequest)
			if !ok {
				return nil, er.New("Argument is not a AbandonChannelRequest")
			}

			//	abandon a channel
			cc, errr := c.withRpcServer()
			if cc != nil {
				var abandonChannelResp *lnrpc.AbandonChannelResponse

				abandonChannelResp, err := cc.AbandonChannel(context.TODO(), abandonChannelReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return abandonChannelResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_AbandonChannel,
	},
	//	service channelbalance
	{
		category:    subcategoryChannel,
		description: "Returns the sum of the total available channel balance across all open channels",

		path: "/lightning/channel/balance",
		req:  nil,
		res:  (*lnrpc.ChannelBalanceResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the channel balance
			cc, errr := c.withRpcServer()
			if cc != nil {
				var channelBalanceResp *lnrpc.ChannelBalanceResponse

				channelBalanceResp, err := cc.ChannelBalance(context.TODO(), nil)
				if err != nil {
					return nil, er.E(err)
				} else {
					return channelBalanceResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_ChannelBalance,
	},
	//	service pendingchannels
	{
		category:    subcategoryChannel,
		description: "Display information pertaining to pending channels",

		path: "/lightning/channel/pending",
		req:  nil,
		res:  (*lnrpc.PendingChannelsResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get pending channels info
			cc, errr := c.withRpcServer()
			if cc != nil {
				var pendingChannelsResp *lnrpc.PendingChannelsResponse

				pendingChannelsResp, err := cc.PendingChannels(context.TODO(), nil)
				if err != nil {
					return nil, er.E(err)
				} else {
					return pendingChannelsResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_PendingChannels,
	},
	//	service listchannels
	{
		category:    subcategoryChannel,
		description: "List all open channels",

		path: "/lightning/channel",
		req:  (*lnrpc.ListChannelsRequest)(nil),
		res:  (*lnrpc.ListChannelsResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			listChannelsReq, ok := m.(*lnrpc.ListChannelsRequest)
			if !ok {
				return nil, er.New("Argument is not a ListChannelsRequest")
			}

			//	get a list of channels
			cc, errr := c.withRpcServer()
			if cc != nil {
				var listChannelsResp *lnrpc.ListChannelsResponse

				listChannelsResp, err := cc.ListChannels(context.TODO(), listChannelsReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return listChannelsResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_ListChannels,
	},
	//	service closedchannels
	{
		category:    subcategoryChannel,
		description: "List all closed channels",

		path: "/lightning/channel/closed",
		req:  (*lnrpc.ClosedChannelsRequest)(nil),
		res:  (*lnrpc.ClosedChannelsResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			closeChannelReq, ok := m.(*lnrpc.ClosedChannelsRequest)
			if !ok {
				return nil, er.New("Argument is not a ClosedChannelRequest")
			}

			//	get a list of all closed channels
			cc, errr := c.withRpcServer()
			if cc != nil {
				var closedChannelsResp *lnrpc.ClosedChannelsResponse

				closedChannelsResp, err := cc.ClosedChannels(context.TODO(), closeChannelReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return closedChannelsResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_ClosedChannels,
	},
	//	service getnetworkinfo
	{
		category:    subcategoryChannel,
		description: "Get statistical information about the current state of the network",

		path: "/lightning/channel/networkinfo",
		req:  nil,
		res:  (*lnrpc.NetworkInfo)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get network info
			cc, errr := c.withRpcServer()
			if cc != nil {
				var networkInfoResp *lnrpc.NetworkInfo

				networkInfoResp, err := cc.GetNetworkInfo(context.TODO(), nil)
				if err != nil {
					return nil, er.E(err)
				} else {
					return networkInfoResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_GetNetworkInfo,
	},
	//	service feereport
	{
		category:    subcategoryChannel,
		description: "Display the current fee policies of all active channels",

		path: "/lightning/channel/feereport",
		req:  nil,
		res:  (*lnrpc.FeeReportResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get fee report
			cc, errr := c.withRpcServer()
			if cc != nil {
				var feeReportResp *lnrpc.FeeReportResponse

				feeReportResp, err := cc.FeeReport(context.TODO(), nil)
				if err != nil {
					return nil, er.E(err)
				} else {
					return feeReportResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_FeeReport,
	},
	//	service updatechanpolicy
	{
		category:    subcategoryChannel,
		description: "Update the channel policy for all channels, or a single channel",

		path: "/lightning/channel/policy",
		req:  (*lnrpc.PolicyUpdateRequest)(nil),
		res:  (*lnrpc.PolicyUpdateResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			policyUpdateReq, ok := m.(*lnrpc.PolicyUpdateRequest)
			if !ok {
				return nil, er.New("Argument is not a ClosedChannelRequest")
			}

			//	invoke Lightning update chan policy command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var policyUpdateResp *lnrpc.PolicyUpdateResponse

				policyUpdateResp, err := cc.UpdateChannelPolicy(context.TODO(), policyUpdateReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return policyUpdateResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_FeeReport,
	},
	//	service exportchanbackup
	{
		category:    subSubCategoryBackup,
		description: "Obtain a static channel back up for a selected channels, or all known channels",

		path: "/lightning/channel/backup/export",
		req:  (*lnrpc.ExportChannelBackupRequest)(nil),
		res:  (*lnrpc.ChannelBackup)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			exportChannelBackupReq, ok := m.(*lnrpc.ExportChannelBackupRequest)
			if !ok {
				return nil, er.New("Argument is not a ExportChannelBackupRequest")
			}

			//	invoke Lightning export chan backup command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var channelBackupResp *lnrpc.ChannelBackup

				channelBackupResp, err := cc.ExportChannelBackup(context.TODO(), exportChannelBackupReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return channelBackupResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_ExportChannelBackup,
	},
	//	service verifychanbackup
	{
		category:    subSubCategoryBackup,
		description: "Verify an existing channel backup",

		path: "/lightning/channel/backup/verify",
		req:  (*lnrpc.ChanBackupSnapshot)(nil),
		res:  (*lnrpc.VerifyChanBackupResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			chanBackupSnapshotReq, ok := m.(*lnrpc.ChanBackupSnapshot)
			if !ok {
				return nil, er.New("Argument is not a ChanBackupSnapshot")
			}

			//	invoke Lightning verify chan backup command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var verifyChanBackupResp *lnrpc.VerifyChanBackupResponse

				verifyChanBackupResp, err := cc.VerifyChanBackup(context.TODO(), chanBackupSnapshotReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return verifyChanBackupResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_VerifyChanBackup,
	},
	//	service restorechanbackup
	{
		category:    subSubCategoryBackup,
		description: "Restore an existing single or multi-channel static channel backup",

		path: "/lightning/channel/backup/restore",
		req:  (*lnrpc.RestoreChanBackupRequest)(nil),
		res:  (*lnrpc.RestoreBackupResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			restoreChanBackupReq, ok := m.(*lnrpc.RestoreChanBackupRequest)
			if !ok {
				return nil, er.New("Argument is not a RestoreChanBackupRequest")
			}

			//	invoke Lightning restore chan backup command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var restoreBackupResp *lnrpc.RestoreBackupResponse

				restoreBackupResp, err := cc.RestoreChannelBackups(context.TODO(), restoreChanBackupReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return restoreBackupResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_RestoreChannelBackups,
	},
	//	service describegraph
	{
		category:    subCategoryGraph,
		description: "Describe the network graph",

		path: "/lightning/graph",
		req:  (*lnrpc.ChannelGraphRequest)(nil),
		res:  (*lnrpc.ChannelGraph)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			channelGraphReq, ok := m.(*lnrpc.ChannelGraphRequest)
			if !ok {
				return nil, er.New("Argument is not a ChannelGraphRequest")
			}

			//	get graph description info
			cc, errr := c.withRpcServer()
			if cc != nil {
				var channelGraphResp *lnrpc.ChannelGraph

				channelGraphResp, err := cc.DescribeGraph(context.TODO(), channelGraphReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return channelGraphResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_DescribeGraph,
	},
	//	service getnodemetrics
	{
		category:    subCategoryGraph,
		description: "Get node metrics",

		path: "/lightning/graph/nodemetrics",
		req:  (*lnrpc.NodeMetricsRequest)(nil),
		res:  (*lnrpc.NodeMetricsResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			nodeMetricsReq, ok := m.(*lnrpc.NodeMetricsRequest)
			if !ok {
				return nil, er.New("Argument is not a NodeMetricsRequest")
			}

			//	get node metrics info
			cc, errr := c.withRpcServer()
			if cc != nil {
				var nodeMetricsResp *lnrpc.NodeMetricsResponse

				nodeMetricsResp, err := cc.GetNodeMetrics(context.TODO(), nodeMetricsReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return nodeMetricsResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_GetNodeMetrics,
	},
	//	service getchaninfo
	{
		category:    subCategoryGraph,
		description: "Get the state of a channel",

		path: "/lightning/graph/channel",
		req:  (*lnrpc.ChanInfoRequest)(nil),
		res:  (*lnrpc.ChannelEdge)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			chanInfoReq, ok := m.(*lnrpc.ChanInfoRequest)
			if !ok {
				return nil, er.New("Argument is not a ChanInfoRequest")
			}

			//	get chan info
			cc, errr := c.withRpcServer()
			if cc != nil {
				var channelEdgeResp *lnrpc.ChannelEdge

				channelEdgeResp, err := cc.GetChanInfo(context.TODO(), chanInfoReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return channelEdgeResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_GetChanInfo,
	},
	//	service getnodeinfo
	{
		category:    subCategoryGraph,
		description: "Get information on a specific node",

		path: "/lightning/graph/nodeinfo",
		req:  (*lnrpc.NodeInfoRequest)(nil),
		res:  (*lnrpc.NodeInfo)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			nodeInfoReq, ok := m.(*lnrpc.NodeInfoRequest)
			if !ok {
				return nil, er.New("Argument is not a NodeInfoRequest")
			}

			//	get node info
			cc, errr := c.withRpcServer()
			if cc != nil {
				var nodeInfoResp *lnrpc.NodeInfo

				nodeInfoResp, err := cc.GetNodeInfo(context.TODO(), nodeInfoReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return nodeInfoResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_GetNodeInfo,
	},
	//	service addinvoice
	{
		category:    subCategoryInvoice,
		description: "Add a new invoice",

		path: "/lightning/invoice/create",
		req:  (*lnrpc.Invoice)(nil),
		res:  (*lnrpc.AddInvoiceResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			invoiceReq, ok := m.(*lnrpc.Invoice)
			if !ok {
				return nil, er.New("Argument is not a Invoice")
			}

			//	add an invoice
			cc, errr := c.withRpcServer()
			if cc != nil {
				var addInvoiceResp *lnrpc.AddInvoiceResponse

				addInvoiceResp, err := cc.AddInvoice(context.TODO(), invoiceReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return addInvoiceResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_AddInvoice,
	},
	//	service lookupinvoice
	{
		category:    subCategoryInvoice,
		description: "Lookup an existing invoice by its payment hash",

		path: "/lightning/invoice/lookup",
		req:  (*lnrpc.PaymentHash)(nil),
		res:  (*lnrpc.Invoice)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			paymentHashReq, ok := m.(*lnrpc.PaymentHash)
			if !ok {
				return nil, er.New("Argument is not a PaymentHash")
			}

			//	lookup an invoice
			cc, errr := c.withRpcServer()
			if cc != nil {
				var InvoiceResp *lnrpc.Invoice

				InvoiceResp, err := cc.LookupInvoice(context.TODO(), paymentHashReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return InvoiceResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_LookupInvoice,
	},
	//	service listinvoices
	{
		category:    subCategoryInvoice,
		description: "List all invoices currently stored within the database. Any active debug invoices are ignored",

		path: "/lightning/invoice",
		req:  (*lnrpc.ListInvoiceRequest)(nil),
		res:  (*lnrpc.ListInvoiceResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			listInvoiceReq, ok := m.(*lnrpc.ListInvoiceRequest)
			if !ok {
				return nil, er.New("Argument is not a ListInvoiceRequest")
			}

			//	list all invoices
			cc, errr := c.withRpcServer()
			if cc != nil {
				var listInvoiceResp *lnrpc.ListInvoiceResponse

				listInvoiceResp, err := cc.ListInvoices(context.TODO(), listInvoiceReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return listInvoiceResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_ListInvoices,
	},
	//	service decodepayreq
	{
		category:    subCategoryInvoice,
		description: "Decode a payment request",

		path: "/lightning/invoice/decodepayreq", // move to /util/payreq/decode
		req:  (*lnrpc.PayReqString)(nil),
		res:  (*lnrpc.PayReq)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			payReqStringReq, ok := m.(*lnrpc.PayReqString)
			if !ok {
				return nil, er.New("Argument is not a PayReqString")
			}

			//	decode payment request
			cc, errr := c.withRpcServer()
			if cc != nil {
				var payReqResp *lnrpc.PayReq

				payReqResp, err := cc.DecodePayReq(context.TODO(), payReqStringReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return payReqResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_DecodePayReq,
	},
	//	service estimatefee
	{
		category:    categoryNeutrino,
		description: "Get fee estimates for sending bitcoin on-chain to multiple addresses",

		path: "/neutrino/estimatefee",
		req:  (*lnrpc.EstimateFeeRequest)(nil),
		res:  (*lnrpc.EstimateFeeResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			estimateFeeReq, ok := m.(*lnrpc.EstimateFeeRequest)
			if !ok {
				return nil, er.New("Argument is not a EstimateFeeRequest")
			}

			//	get estimate fee info
			cc, errr := c.withRpcServer()
			if cc != nil {
				var estimateFeeResp *lnrpc.EstimateFeeResponse

				estimateFeeResp, err := cc.EstimateFee(context.TODO(), estimateFeeReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return estimateFeeResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_EstimateFee,
	},
	//	service sendmany
	{
		category:    subCategoryTransaction,
		description: "Send bitcoin on-chain to multiple addresses",

		path: "/wallet/transaction/sendmany",
		req:  (*lnrpc.SendManyRequest)(nil),
		res:  (*lnrpc.SendManyResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			sendManyReq, ok := m.(*lnrpc.SendManyRequest)
			if !ok {
				return nil, er.New("Argument is not a SendManyRequest")
			}

			//	send coins to many addresses
			cc, errr := c.withRpcServer()
			if cc != nil {
				var sendManyResp *lnrpc.SendManyResponse

				sendManyResp, err := cc.SendMany(context.TODO(), sendManyReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return sendManyResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_SendMany,
	},
	//	service sendcoins
	{
		category:    subCategoryTransaction,
		description: "Send bitcoin on-chain to an address",

		path: "/wallet/transaction/sendcoins",
		req:  (*lnrpc.SendCoinsRequest)(nil),
		res:  (*lnrpc.SendCoinsResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			sendCoinsReq, ok := m.(*lnrpc.SendCoinsRequest)
			if !ok {
				return nil, er.New("Argument is not a SendCoinsRequest")
			}

			//	send coins to one addresses
			cc, errr := c.withRpcServer()
			if cc != nil {
				var sendCoinsResp *lnrpc.SendCoinsResponse

				sendCoinsResp, err := cc.SendCoins(context.TODO(), sendCoinsReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return sendCoinsResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_SendCoins,
	},
	//	service listunspent
	{
		category:    subCategoryUnspent,
		description: "List utxos available for spending",

		path: "/wallet/unspent",
		req:  (*lnrpc.ListUnspentRequest)(nil),
		res:  (*lnrpc.ListUnspentResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			listUnspentReq, ok := m.(*lnrpc.ListUnspentRequest)
			if !ok {
				return nil, er.New("Argument is not a ListUnspentRequest")
			}

			//	get a list of available utxos
			cc, errr := c.withRpcServer()
			if cc != nil {
				var listUnspentResp *lnrpc.ListUnspentResponse

				listUnspentResp, err := cc.ListUnspent(context.TODO(), listUnspentReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return listUnspentResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_ListUnspent,
	},
	//	service setnetworkstewardvote
	{
		category:    subCategoryNetworkStewardVote,
		description: "Configure the wallet to vote for a network steward when making payments (note: payments to segwit addresses cannot vote)",

		path: "/wallet/networkstewardvote/set",
		req:  (*lnrpc.SetNetworkStewardVoteRequest)(nil),
		res:  nil,
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			setNetworkStewardVoteReq, ok := m.(*lnrpc.SetNetworkStewardVoteRequest)
			if !ok {
				return nil, er.New("Argument is not a SetNetworkStewardVoteRequest")
			}

			//	set network steward vote
			cc, errr := c.withRpcServer()
			if cc != nil {
				var setNetworkStewardVoteResp *lnrpc.SetNetworkStewardVoteResponse

				setNetworkStewardVoteResp, err := cc.SetNetworkStewardVote(context.TODO(), setNetworkStewardVoteReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return setNetworkStewardVoteResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_SetNetworkStewardVote,
	},
	//	service getnetworkstewardvote
	{
		category:    subCategoryNetworkStewardVote,
		description: "Find out how the wallet is currently configured to vote in a network steward election",

		path: "/wallet/networkstewardvote",
		req:  nil,
		res:  (*lnrpc.GetNetworkStewardVoteResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get network steward vote
			cc, errr := c.withRpcServer()
			if cc != nil {
				var getNetworkStewardVoteResp *lnrpc.GetNetworkStewardVoteResponse

				getNetworkStewardVoteResp, err := cc.GetNetworkStewardVote(context.TODO(), nil)
				if err != nil {
					return nil, er.E(err)
				} else {
					return getNetworkStewardVoteResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_GetNetworkStewardVote,
	},
	//	service bcasttransaction
	{
		category:    categoryNeutrino,
		description: "Broadcast a transaction onchain",

		path: "/neutrino/bcasttransaction",
		req:  (*lnrpc.BcastTransactionRequest)(nil),
		res:  (*lnrpc.BcastTransactionResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			bcastTransactionReq, ok := m.(*lnrpc.BcastTransactionRequest)
			if !ok {
				return nil, er.New("Argument is not a BcastTransactionRequest")
			}

			//	invoke Lightning broadcast transaction in chain command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var bcastTransactionResp *lnrpc.BcastTransactionResponse

				bcastTransactionResp, err := cc.BcastTransaction(context.TODO(), bcastTransactionReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return bcastTransactionResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_BcastTransaction,
	},
	//	service sendpayment
	{
		category:    subCategoryPayment,
		description: "Send a payment over lightning",

		path: "/lightning/payment/send",
		req:  (*lnrpc.SendRequest)(nil),
		res:  (*lnrpc.SendResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			sendReq, ok := m.(*lnrpc.SendRequest)
			if !ok {
				return nil, er.New("Argument is not a SendRequest")
			}

			//	invoke Lightning send payment command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var sendResp *lnrpc.SendResponse

				sendResp, err := cc.SendPaymentSync(context.TODO(), sendReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return sendResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_SendPaymentSync,
	},
	//	TODO: service payinvoice
	//	uses a stream Router_SendPaymentV2Server to send the payment updates - how to do it with RESP endpoints ?
	{
		category:    subCategoryPayment,
		description: "Pay an invoice over lightning",

		path: "/lightning/payment/payinvoice",
		req:  (*routerrpc.SendPaymentRequest)(nil),
		res:  nil,
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			sendPaymentReq, ok := m.(*routerrpc.SendPaymentRequest)
			if !ok {
				return nil, er.New("Argument is not a SendPaymentRequest")
			}

			//	invoke Lightning send payment command
			cc, errr := c.withRouterServer()
			if cc != nil {
				err := cc.SendPaymentV2(sendPaymentReq, nil)
				if err != nil {
					return nil, er.E(err)
				} else {
					return nil, nil
				}
			} else {
				return nil, errr
			}
		},
	},
	//	service sendtoroute
	{
		category:    subCategoryPayment,
		description: "Send a payment over a predefined route",

		path: "/lightning/payment/sendtoroute",
		req:  (*lnrpc.SendToRouteRequest)(nil),
		res:  (*lnrpc.SendResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			sendToRouteReq, ok := m.(*lnrpc.SendToRouteRequest)
			if !ok {
				return nil, er.New("Argument is not a SendToRouteRequest")
			}

			//	invoke Lightning send to route command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var sendResp *lnrpc.SendResponse

				sendResp, err := cc.SendToRouteSync(context.TODO(), sendToRouteReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return sendResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_SendToRouteSync,
	},
	//	service listpayments
	{
		category:    subCategoryPayment,
		description: "List all outgoing payments",

		path: "/lightning/payment",
		req:  (*lnrpc.ListPaymentsRequest)(nil),
		res:  (*lnrpc.ListPaymentsResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			listPaymentsReq, ok := m.(*lnrpc.ListPaymentsRequest)
			if !ok {
				return nil, er.New("Argument is not a ListPaymentsRequest")
			}

			//	invoke Lightning list payments command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var listPaymentsResp *lnrpc.ListPaymentsResponse

				listPaymentsResp, err := cc.ListPayments(context.TODO(), listPaymentsReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return listPaymentsResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_ListPayments,
	},
	//	service queryroutes
	{
		category:    subCategoryPayment,
		description: "Query a route to a destination",

		path: "/lightning/payment/queryroutes",
		req:  (*lnrpc.QueryRoutesRequest)(nil),
		res:  (*lnrpc.QueryRoutesResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			queryRoutesReq, ok := m.(*lnrpc.QueryRoutesRequest)
			if !ok {
				return nil, er.New("Argument is not a QueryRoutesRequest")
			}

			//	invoke Lightning query routes command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var queryRoutesResp *lnrpc.QueryRoutesResponse

				queryRoutesResp, err := cc.QueryRoutes(context.TODO(), queryRoutesReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return queryRoutesResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_QueryRoutes,
	},
	//	service fwdinghistory
	{
		category:    subCategoryPayment,
		description: "Query the history of all forwarded HTLCs",

		path: "/lightning/payment/fwdinghistory",
		req:  (*lnrpc.ForwardingHistoryRequest)(nil),
		res:  (*lnrpc.ForwardingHistoryResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			forwardingHistoryReq, ok := m.(*lnrpc.ForwardingHistoryRequest)
			if !ok {
				return nil, er.New("Argument is not a ForwardingHistoryRequest")
			}

			//	invoke Lightning forwarding history command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var forwardingHistoryResp *lnrpc.ForwardingHistoryResponse

				forwardingHistoryResp, err := cc.ForwardingHistory(context.TODO(), forwardingHistoryReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return forwardingHistoryResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_ForwardingHistory,
	},
	//	TODO: service trackpayment
	//	uses a stream Router_SendPaymentV2Server to send the payment updates - how to do it with RESP endpoints ?
	{
		category:    subCategoryPayment,
		description: "Track progress of an existing payment",

		path: "/lightning/payment/track",
		req:  (*routerrpc.TrackPaymentRequest)(nil),
		res:  nil,
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			trackPaymentReq, ok := m.(*routerrpc.TrackPaymentRequest)
			if !ok {
				return nil, er.New("Argument is not a TrackPaymentRequest")
			}

			//	invoke Lightning send payment command
			cc, errr := c.withRouterServer()
			if cc != nil {
				err := cc.TrackPaymentV2(trackPaymentReq, nil)
				if err != nil {
					return nil, er.E(err)
				} else {
					return nil, nil
				}
			} else {
				return nil, errr
			}
		},
	},
	//	service querymc
	{
		category:    subCategoryPayment,
		description: "Query the internal mission control state",

		path: "/lightning/payment/querymc",
		req:  nil,
		res:  (*routerrpc.QueryMissionControlResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	query the mission control status
			cc, errr := c.withRouterServer()
			if cc != nil {
				var queryMissionControlResp *routerrpc.QueryMissionControlResponse

				queryMissionControlResp, err := cc.QueryMissionControl(context.TODO(), nil)
				if err != nil {
					return nil, er.E(err)
				} else {
					return queryMissionControlResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Router_QueryMissionControl,
	},
	//	service queryprob
	{
		category:    subCategoryPayment,
		description: "Estimate a success probability",

		path: "/lightning/payment/queryprob",
		req:  (*routerrpc.QueryProbabilityRequest)(nil),
		res:  (*routerrpc.QueryProbabilityResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			queryProbabilityReq, ok := m.(*routerrpc.QueryProbabilityRequest)
			if !ok {
				return nil, er.New("Argument is not a QueryProbabilityRequest")
			}

			//	invoke the probability service
			cc, errr := c.withRouterServer()
			if cc != nil {
				var queryProbabilityResp *routerrpc.QueryProbabilityResponse

				queryProbabilityResp, err := cc.QueryProbability(context.TODO(), queryProbabilityReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return queryProbabilityResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Router_QueryProbability,
	},
	//	service resetmc
	{
		category:    subCategoryPayment,
		description: "Reset internal mission control state",

		path: "/lightning/payment/resetmc",
		req:  nil,
		res:  nil,
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	invoke reset mission controle service
			cc, errr := c.withRouterServer()
			if cc != nil {
				var resetMissionControlResp *routerrpc.ResetMissionControlResponse

				resetMissionControlResp, err := cc.ResetMissionControl(context.TODO(), nil)
				if err != nil {
					return nil, er.E(err)
				} else {
					return resetMissionControlResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Router_ResetMissionControl,
	},
	//	service buildroute
	{
		category:    subCategoryPayment,
		description: "Build a route from a list of hop pubkeys",

		path: "/lightning/payment/buildroute",
		req:  (*routerrpc.BuildRouteRequest)(nil),
		res:  (*routerrpc.BuildRouteResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			buildRouteReq, ok := m.(*routerrpc.BuildRouteRequest)
			if !ok {
				return nil, er.New("Argument is not a BuildRouteRequest")
			}

			//	invoke reset mission controle service
			cc, errr := c.withRouterServer()
			if cc != nil {
				var buildRouteResp *routerrpc.BuildRouteResponse

				buildRouteResp, err := cc.BuildRoute(context.TODO(), buildRouteReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return buildRouteResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Router_BuildRoute,
	},
	//	service connect
	{
		category:    subCategoryPeer,
		description: "Connect to a remote pld peer",

		path: "/lightning/peer/connect",
		req:  (*lnrpc.ConnectPeerRequest)(nil),
		res:  nil,
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			connectPeerReq, ok := m.(*lnrpc.ConnectPeerRequest)
			if !ok {
				return nil, er.New("Argument is not a ConnectPeerRequest")
			}

			//	invoke Lightning connect peer command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var connectPeerResp *lnrpc.ConnectPeerResponse

				connectPeerResp, err := cc.ConnectPeer(context.TODO(), connectPeerReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return connectPeerResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_ConnectPeer,
	},
	//	service disconnect
	{
		category:    subCategoryPeer,
		description: "Disconnect a remote pld peer identified by public key",

		path: "/lightning/peer/disconnect",
		req:  (*lnrpc.DisconnectPeerRequest)(nil),
		res:  nil,
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			disconnectPeerReq, ok := m.(*lnrpc.DisconnectPeerRequest)
			if !ok {
				return nil, er.New("Argument is not a DisconnectPeerRequest")
			}

			//	invoke Lightning disconnect peer command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var disconnectPeerResp *lnrpc.DisconnectPeerResponse

				disconnectPeerResp, err := cc.DisconnectPeer(context.TODO(), disconnectPeerReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return disconnectPeerResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_DisconnectPeer,
	},
	//	service listpeers
	{
		category:    subCategoryPeer,
		description: "List all active, currently connected peers",

		path: "/lightning/peer",
		req:  nil,
		res:  (*lnrpc.ListPeersResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			listPeersReq := &lnrpc.ListPeersRequest{
				LatestError: true,
			}

			//	invoke Lightning list peers command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var listPeersResp *lnrpc.ListPeersResponse

				listPeersResp, err := cc.ListPeers(context.TODO(), listPeersReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return listPeersResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_ListPeers,
	},
	//	service signmessage
	{
		category:    subCategoryAddress,
		description: "Signs a message using the private key of a payment address",

		path: "/wallet/address/signmessage",
		req:  (*lnrpc.SignMessageRequest)(nil),
		res:  (*lnrpc.SignMessageResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			signMessageReq, ok := m.(*lnrpc.SignMessageRequest)
			if !ok {
				return nil, er.New("Argument is not a SignMessageRequest")
			}

			//	invoke wallet sign message command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var signMessageResp *lnrpc.SignMessageResponse

				signMessageResp, err := cc.SignMessage(context.TODO(), signMessageReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return signMessageResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Signer_SignMessage,
	},
	//	service getsecret
	{
		category:    categoryWallet,
		description: "Get a secret seed",

		path: "/wallet/getsecret",
		req:  (*lnrpc.GetSecretRequest)(nil),
		res:  (*lnrpc.GetSecretResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			getSecretReq, ok := m.(*lnrpc.GetSecretRequest)
			if !ok {
				return nil, er.New("Argument is not a GetSecretRequest")
			}

			//	invoke wallet get secret command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var getSecretResp *lnrpc.GetSecretResponse

				getSecretResp, err := cc.GetSecret(context.TODO(), getSecretReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return getSecretResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_GetSecret,
	},
	//	service importprivkey
	{
		category:    subCategoryAddress,
		description: "Imports a WIF-encoded private key to the 'imported' account",

		path: "/wallet/address/import",
		req:  (*lnrpc.ImportPrivKeyRequest)(nil),
		res:  (*lnrpc.ImportPrivKeyResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			importPrivKeyReq, ok := m.(*lnrpc.ImportPrivKeyRequest)
			if !ok {
				return nil, er.New("Argument is not a ImportPrivKeyRequest")
			}

			//	invoke wallet import private key command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var importPrivKeyResp *lnrpc.ImportPrivKeyResponse

				importPrivKeyResp, err := cc.ImportPrivKey(context.TODO(), importPrivKeyReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return importPrivKeyResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_ImportPrivKey,
	},
	//	service listlockunspent
	{
		category:    subSubCategoryLock,
		description: "Returns a JSON array of outpoints marked as locked (with lockunspent) for this wallet session",

		path: "/wallet/unspent/lock",
		req:  nil,
		res:  (*lnrpc.ListLockUnspentResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	invoke wallet list lock unspent command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var listLockUnspentResp *lnrpc.ListLockUnspentResponse

				listLockUnspentResp, err := cc.ListLockUnspent(context.TODO(), nil)
				if err != nil {
					return nil, er.E(err)
				} else {
					return listLockUnspentResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_ListLockUnspent,
	},
	//	service lockunspent
	{
		category:    subSubCategoryLock,
		description: "Locks or unlocks an unspent output",

		path: "/wallet/unspent/lock/create", // TODO: /wallet/unspent/lock/delete
		req:  (*lnrpc.LockUnspentRequest)(nil),
		res:  (*lnrpc.LockUnspentResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			lockUnspentReq, ok := m.(*lnrpc.LockUnspentRequest)
			if !ok {
				return nil, er.New("Argument is not a LockUnspentRequest")
			}

			//	invoke wallet lock unspent command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var lockUnspentResp *lnrpc.LockUnspentResponse

				lockUnspentResp, err := cc.LockUnspent(context.TODO(), lockUnspentReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return lockUnspentResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_LockUnspent,
	},
	//	service createtransaction
	{
		category:    subCategoryTransaction,
		description: "Create a transaction but do not send it to the chain",

		path: "/wallet/transaction/create",
		req:  (*lnrpc.CreateTransactionRequest)(nil),
		res:  (*lnrpc.CreateTransactionResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			createTransactionReq, ok := m.(*lnrpc.CreateTransactionRequest)
			if !ok {
				return nil, er.New("Argument is not a CreateTransactionRequest")
			}

			//	invoke wallet create transaction command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var createTransactionResp *lnrpc.CreateTransactionResponse

				createTransactionResp, err := cc.CreateTransaction(context.TODO(), createTransactionReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return createTransactionResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_CreateTransaction,
	},
	//	service dumpprivkey
	{
		category:    subCategoryAddress,
		description: "Returns the private key in WIF encoding that controls some wallet address",

		path: "/wallet/address/dumpprivkey",
		req:  (*lnrpc.DumpPrivKeyRequest)(nil),
		res:  (*lnrpc.DumpPrivKeyResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			dumpPrivKeyReq, ok := m.(*lnrpc.DumpPrivKeyRequest)
			if !ok {
				return nil, er.New("Argument is not a DumpPrivKeyRequest")
			}

			//	invoke wallet dump private key command
			cc, errr := c.withRpcServer()
			if cc != nil {
				var dumpPrivKeyResp *lnrpc.DumpPrivKeyResponse

				dumpPrivKeyResp, err := cc.DumpPrivKey(context.TODO(), dumpPrivKeyReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return dumpPrivKeyResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.Lightning_DumpPrivKey,
	},
	//	service wtclient: Add
	{
		category:    categoryWatchtower,
		description: "Register a watchtower to use for future sessions/backups",

		path: "/wtclient/tower/create",
		req:  (*wtclientrpc.AddTowerRequest)(nil),
		res:  nil,
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			addTowerReq, ok := m.(*wtclientrpc.AddTowerRequest)
			if !ok {
				return nil, er.New("Argument is not a AddTowerRequest")
			}

			//	invoke wallet get transactions command
			cc, errr := c.withWatchTowerClient()

			if cc != nil {
				var addTowerResp *wtclientrpc.AddTowerResponse

				addTowerResp, err := cc.AddTower(context.TODO(), addTowerReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return addTowerResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.WatchtowerClient_AddTower,
	},
	//	service wtclient: Remove
	{
		category:    categoryWatchtower,
		description: "Remove a watchtower to prevent its use for future sessions/backups",

		path: "/wtclient/tower/remove",
		req:  (*wtclientrpc.RemoveTowerRequest)(nil),
		res:  nil,
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			removeTowerReq, ok := m.(*wtclientrpc.RemoveTowerRequest)
			if !ok {
				return nil, er.New("Argument is not a RemoveTowerRequest")
			}

			//	invoke wallet get transactions command
			cc, errr := c.withWatchTowerClient()

			if cc != nil {
				var removeTowerResp *wtclientrpc.RemoveTowerResponse

				removeTowerResp, err := cc.RemoveTower(context.TODO(), removeTowerReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return removeTowerResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.WatchtowerClient_RemoveTower,
	},
	//	service wtclient: Towers
	{
		category:    categoryWatchtower,
		description: "Display information about all registered watchtowers",

		path: "/wtclient/tower",
		req:  (*wtclientrpc.ListTowersRequest)(nil),
		res:  (*wtclientrpc.ListTowersResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			listTowersReq, ok := m.(*wtclientrpc.ListTowersRequest)
			if !ok {
				return nil, er.New("Argument is not a ListTowersRequest")
			}

			//	invoke wallet get transactions command
			cc, errr := c.withWatchTowerClient()

			if cc != nil {
				var listTowersResp *wtclientrpc.ListTowersResponse

				listTowersResp, err := cc.ListTowers(context.TODO(), listTowersReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return listTowersResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.WatchtowerClient_ListTowers,
	},
	//	service wtclient: Tower
	{
		category:    categoryWatchtower,
		description: "Display information about a specific registered watchtower",

		path: "/wtclient/tower/getinfo",
		req:  (*wtclientrpc.GetTowerInfoRequest)(nil),
		res:  (*wtclientrpc.Tower)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			getTowerInfoReq, ok := m.(*wtclientrpc.GetTowerInfoRequest)
			if !ok {
				return nil, er.New("Argument is not a GetTowerInfoRequest")
			}

			//	invoke wallet get transactions command
			cc, errr := c.withWatchTowerClient()

			if cc != nil {
				var towerResp *wtclientrpc.Tower

				towerResp, err := cc.GetTowerInfo(context.TODO(), getTowerInfoReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return towerResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.WatchtowerClient_GetTowerInfo,
	},
	//	service wtclient: stats
	{
		category:    categoryWatchtower,
		description: "Display the session stats of the watchtower client",

		path: "/wtclient/tower/stats",
		req:  nil,
		res:  (*wtclientrpc.StatsResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			statsReq, ok := m.(*wtclientrpc.StatsRequest)
			if !ok {
				return nil, er.New("Argument is not a StatsRequest")
			}

			//	invoke wallet get transactions command
			cc, errr := c.withWatchTowerClient()

			if cc != nil {
				var statsResp *wtclientrpc.StatsResponse

				statsResp, err := cc.Stats(context.TODO(), statsReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return statsResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.WatchtowerClient_Stats,
	},
	//	service wtclient: policy
	{
		category:    categoryWatchtower,
		description: "Display the active watchtower client policy configuration",

		path: "/wtclient/tower/policy",
		req:  (*wtclientrpc.PolicyRequest)(nil),
		res:  (*wtclientrpc.PolicyResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get the request payload
			policyReq, ok := m.(*wtclientrpc.PolicyRequest)
			if !ok {
				return nil, er.New("Argument is not a PolicyRequest")
			}

			//	invoke wallet get transactions command
			cc, errr := c.withWatchTowerClient()

			if cc != nil {
				var policyResp *wtclientrpc.PolicyResponse

				policyResp, err := cc.Policy(context.TODO(), policyReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return policyResp, nil
				}
			} else {
				return nil, errr
			}
		},

		getHelpInfo: pkthelp.WatchtowerClient_Policy,
	},
}

type RpcContext struct {
	MaybeCC               *chainreg.ChainControl
	MaybeNeutrino         *neutrino.ChainService
	MaybeWallet           *wallet.Wallet
	MaybeRpcServer        lnrpc.LightningServer
	MaybeWalletUnlocker   *walletunlocker.UnlockerService
	MaybeMetaService      lnrpc.MetaServiceServer
	MaybeVerRPCServer     verrpc.VersionerServer
	MaybeRouterServer     routerrpc.RouterServer
	MaybeWatchTowerClient wtclientrpc.WatchtowerClientClient
}

func with(thing interface{}, name string) er.R {
	if thing == nil {
		return er.Errorf("Could not call function because [%s] is not yet ready", name)
	}
	return nil
}
func (c *RpcContext) withCC() (*chainreg.ChainControl, er.R) {
	return c.MaybeCC, with(c.MaybeCC, "ChainController")
}
func (c *RpcContext) withNeutrino() (*neutrino.ChainService, er.R) {
	return c.MaybeNeutrino, with(c.MaybeNeutrino, "Neutrino")
}
func (c *RpcContext) withWallet() (*wallet.Wallet, er.R) {
	return c.MaybeWallet, with(c.MaybeWallet, "Wallet")
}
func (c *RpcContext) withRpcServer() (lnrpc.LightningServer, er.R) {
	return c.MaybeRpcServer, with(c.MaybeRpcServer, "LightningServer")
}
func (c *RpcContext) withUnlocker() (*walletunlocker.UnlockerService, er.R) {
	return c.MaybeWalletUnlocker, with(c.MaybeWalletUnlocker, "UnlockerService")
}
func (c *RpcContext) withMetaServer() (lnrpc.MetaServiceServer, er.R) {
	return c.MaybeMetaService, with(c.MaybeMetaService, "MetaServiceService")
}
func (c *RpcContext) withVerRPCServer() (verrpc.VersionerServer, er.R) {
	return c.MaybeVerRPCServer, with(c.MaybeVerRPCServer, "VersionerService")
}
func (c *RpcContext) withRouterServer() (routerrpc.RouterServer, er.R) {
	return c.MaybeRouterServer, with(c.MaybeRouterServer, "RouterServer")
}

func (c *RpcContext) withWatchTowerClient() (wtclientrpc.WatchtowerClientClient, er.R) {
	return c.MaybeWatchTowerClient, with(c.MaybeWatchTowerClient, "WatchTowerClient")
}

type SimpleHandler struct {
	rf RpcFunc
	c  *RpcContext
}

func unmarshal1(r *http.Request, m proto.Message, isJson bool) er.R {
	if b, err := io.ReadAll(r.Body); err != nil {
		return er.E(err)
	} else if isJson {
		// Use jsoniter for unmarshaling because it is far more forgiving
		if err := jsoniter.Unmarshal(b, m); err != nil {
			return er.E(err)
		}
	} else if err := proto.Unmarshal(b, m); err != nil {
		return er.E(err)
	}
	return nil
}

func unmarshal(r *http.Request, m proto.Message, isJson bool) er.R {
	if isJson {
		if err := jsonpb.Unmarshal(r.Body, m); err != nil {
			return er.E(err)
		}
	} else {
		if b, err := io.ReadAll(r.Body); err != nil {
			return er.E(err)
		} else if err := proto.Unmarshal(b, m); err != nil {
			return er.E(err)
		}
	}
	return nil
}
func marshal(w http.ResponseWriter, m proto.Message, isJson bool) er.R {
	if m == nil {
		return nil
	}
	if isJson {
		marshaler := jsonpb.Marshaler{
			OrigName:     false,
			EnumsAsInts:  false,
			EmitDefaults: true,
			Indent:       "\t",
		}
		if s, err := marshaler.MarshalToString(m); err != nil {
			return er.E(err)
		} else if _, err := io.WriteString(w, s); err != nil {
			return er.E(err)
		}
	} else {
		if b, err := proto.Marshal(m); err != nil {
			return er.E(err)
		} else if _, err := w.Write(b); err != nil {
			return er.E(err)
		}
	}
	return nil
}

func (s *SimpleHandler) ServeHttpOrErr(w http.ResponseWriter, r *http.Request, isJson bool) er.R {
	var req proto.Message

	//	check if the URI is for command help
	if r.RequestURI == URI_prefix+helpURI_prefix+s.rf.path {

		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return er.New("405 - Request should be a GET because the help endpoint requires no input")
		}
		err := marshalHelp(w, s.rf.getHelpInfo())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
		return nil
	}

	//	command URI handler
	if s.rf.req != nil {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return er.New("405 - Request should be a POST because the endpoint requires input")
		}
		req1 := reflect.New(reflect.TypeOf(s.rf.req).Elem())
		if r, ok := req1.Interface().(proto.Message); !ok {
			panic("elem is not a proto.Message")
		} else {
			req = r
		}
		if err := unmarshal(r, req, isJson); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return err
		}
	} else if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return er.New("405 - Request should be a GET because the endpoint requires no input")
	}
	if res, err := s.rf.f(s.c, req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	} else if err := marshal(w, res, isJson); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}
	return nil
}

func (s *SimpleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ct := strings.ToLower(r.Header.Get("Content-Type"))
	isJson := strings.Contains(ct, "application/json")
	if !isJson && !strings.Contains(ct, "application/protobuf") {
		if r.Method == "GET" {
			isJson = true
		} else {
			w.Header().Set("Connection", "close")
			w.Header().Set("Content-Type", "text/plain")
			http.Error(w, "415 - Invalid content type, must be json or protobuf", http.StatusUnsupportedMediaType)
			return
		}
	}
	if err := s.ServeHttpOrErr(w, r, isJson); err != nil {
		if err = marshal(w, &lnrpc.RestError{
			Message: err.Message(),
			Stack:   err.Stack(),
		}, isJson); err != nil {
			log.Errorf("Error replying to request for [%s] from [%s] - error sending error, giving up: [%s]",
				r.RequestURI, r.RemoteAddr, err)
		}
	}
}

func RestHandlers(c *RpcContext) *mux.Router {
	r := mux.NewRouter()
	for _, rf := range rpcFunctions {
		r.Handle(URI_prefix+rf.path, &SimpleHandler{c: c, rf: rf})
		r.Handle(URI_prefix+helpURI_prefix+rf.path, &SimpleHandler{c: c, rf: rf})
	}

	//	add a handler for endpoint not found (404)
	r.NotFoundHandler = http.HandlerFunc(func(httpResponse http.ResponseWriter, r *http.Request) {
		httpResponse.Header().Set("Content-Type", "text/plain")
		http.Error(httpResponse, "404 - invalid endpoint: for help on all endpoints go to /api/v1 URI", http.StatusNotFound)
	})

	return r
}
