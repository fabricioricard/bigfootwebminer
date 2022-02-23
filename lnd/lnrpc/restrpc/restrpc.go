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
	"github.com/pkt-cash/pktd/lnd/walletunlocker"
	"github.com/pkt-cash/pktd/neutrino"
	"github.com/pkt-cash/pktd/pktlog/log"
	"github.com/pkt-cash/pktd/pktwallet/waddrmgr"
	"github.com/pkt-cash/pktd/pktwallet/wallet"
)

type RpcFunc struct {
	path string
	req  proto.Message
	res  proto.Message
	f    func(c *RpcContext, m proto.Message) (proto.Message, er.R)
}

var rpcFunctions []RpcFunc = []RpcFunc{
	//WalletUnlocker: Wallet unlock
	//Will try to unlock the wallet with the password(s) provided
	{
		path: "/api/v1/wallet/unlock",
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
	},
	//WalletUnlocker: Wallet create
	//Will try to create/restore wallet
	{
		path: "/api/v1/wallet/create",
		req:  (*lnrpc.CreateWalletRequest)(nil),
		res:  (*lnrpc.CreateWalletResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {
			req, ok := m.(*lnrpc.CreateWalletRequest)
			if !ok {
				return nil, er.New("Argument is not a CreateWalletRequest")
			}
			u, err := c.withUnlocker()
			if err != nil {
				return nil, err
			}
			resp, errr := u.CreateWallet(context.TODO(), req)
			if errr != nil {
				return nil, er.E(errr)
			}
			return resp, nil
		},
	},
	//MetaService get info
	{
		path: "/api/v1/meta/getinfo",
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
	},
	//MetaService change wallet password
	{
		path: "/api/v1/meta/changepassword",
		req:  (*lnrpc.ChangePasswordRequest)(nil),
		res:  (*lnrpc.ChangePasswordResponse)(nil),
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
	},
	//Wallet balance
	//requires unlocked wallet -> access to rpcServer
	{
		path: "/api/v1/lightning/walletbalance",
		req:  nil,
		res:  (*lnrpc.GetAddressBalancesResponse)(nil),
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
	},
	//Wallet transactions
	//requires unlocked wallet -> access to rpcServer
	{
		path: "/api/v1/lightning/gettransactions",
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
	},
	//New wallet address
	//requires unlocked wallet -> access to rpcServer
	{
		path: "/api/v1/lightning/getnewaddress",
		req:  (*lnrpc.GetNewAddressRequest)(nil),
		res:  (*lnrpc.GetNewAddressResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {
			req, ok := m.(*lnrpc.GetNewAddressRequest)
			if !ok {
				return nil, er.New("Argument is not a GetAddressBalancesRequest")
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
	},
	//GetAddressBalances
	{
		path: "/api/v1/lightning/getaddressbalances",
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
	},
	//Sendfrom
	{
		path: "/api/v1/lightning/sendfrom",
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
	},
	//GetWalletSeed
	{
		path: "/api/v1/lightning/getwalletseed",
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
	},
	//GetTransaction
	{
		path: "/api/v1/lightning/gettransaction",
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
	},
	//Resync
	{
		path: "/api/v1/lightning/resync",
		req:  (*lnrpc.ReSyncChainRequest)(nil),
		res:  (*lnrpc.ReSyncChainResponse)(nil),
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
	},
	//StopResync
	{
		path: "/api/v1/lightning/stopresync",
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
	},
	//	meta service get recovery info
	{
		path: "/api/v1/meta/getrecoveryinfo",
		req:  nil,
		res:  (*lnrpc.GetRecoveryInfoResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get Lightning recovery info
			cc, errr := c.withRpcServer()
			if cc != nil {
				var recoveryInfoRsp *lnrpc.GetRecoveryInfoResponse

				recoveryInfoRsp, err := cc.GetRecoveryInfo(context.TODO(), nil)
				if err != nil {
					return nil, er.E(err)
				} else {
					return recoveryInfoRsp, nil
				}
			} else {
				return nil, errr
			}
		},
	},
	//	service debug level
	{
		path: "/api/v1/debuglevel",
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
				var debugLevelRsp *lnrpc.DebugLevelResponse

				debugLevelRsp, err := cc.DebugLevel(context.TODO(), debugLevelReq)
				if err != nil {
					return nil, er.E(err)
				} else {
					return debugLevelRsp, nil
				}
			} else {
				return nil, errr
			}
		},
	},
	//	service to stop the pld daemon
	{
		path: "/api/v1/stop",
		req:  nil,
		res:  (*lnrpc.StopResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	invoke Lightning stop daemon
			cc, errr := c.withRpcServer()
			if cc != nil {
				var stopResponse *lnrpc.StopResponse

				stopResponse, err := cc.StopDaemon(context.TODO(), nil)
				if err != nil {
					return nil, er.E(err)
				} else {
					return stopResponse, nil
				}
			} else {
				return nil, errr
			}
		},
	},
	//	service daemon version
	{
		path: "/api/v1/version",
		req:  nil,
		res:  (*lnrpc.GetRecoveryInfoResponse)(nil),
		f: func(c *RpcContext, m proto.Message) (proto.Message, er.R) {

			//	get Lightning recovery info
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
}

type RpcContext struct {
	MaybeCC             *chainreg.ChainControl
	MaybeNeutrino       *neutrino.ChainService
	MaybeWallet         *wallet.Wallet
	MaybeRpcServer      lnrpc.LightningServer
	MaybeWalletUnlocker *walletunlocker.UnlockerService
	MaybeMetaService    lnrpc.MetaServiceServer
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
	if s.rf.req != nil {
		if r.Method != "POST" {
			return er.New("Request should be a POST because the endpoint requires input")
		}
		req1 := reflect.New(reflect.TypeOf(s.rf.req).Elem())
		if r, ok := req1.Interface().(proto.Message); !ok {
			panic("elem is not a proto.Message")
		} else {
			req = r
		}
		if err := unmarshal(r, req, isJson); err != nil {
			return err
		}
	} else if r.Method != "GET" {
		return er.New("Request should be a GET because the endpoint requires no input")
	}
	if res, err := s.rf.f(s.c, req); err != nil {
		return err
	} else if err := marshal(w, res, isJson); err != nil {
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
			http.Error(w, "400 - Invalid content type, must be json or protobuf", http.StatusBadRequest)
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
		r.Handle(rf.path, &SimpleHandler{c: c, rf: rf})
	}
	return r
}
