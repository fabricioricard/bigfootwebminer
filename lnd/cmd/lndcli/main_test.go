package main

import (
	"testing"
)

type testScenario struct {
	name            string
	command         string
	arguments       []string
	expectedPayload string
}

func TestFormatRequestPayload(t *testing.T) {

	pldServer = "http://localhost:8080"

	var testCases []testScenario = []testScenario{
		//	test commands of "meta" group
		{
			name:            "debuglevel CLI options",
			command:         "meta/debuglevel",
			arguments:       []string{"--show", "--level_spec=debug"},
			expectedPayload: `{ "show": true, "level_spec": "debug" }`,
		},
		{
			name:            "getinfo CLI options",
			command:         "meta/getinfo",
			arguments:       []string{},
			expectedPayload: `{  }`,
		},
		{
			name:            "version CLI options",
			command:         "meta/version",
			arguments:       []string{},
			expectedPayload: `{  }`,
		},
		//	test commands of "util/seed" group
		{
			name:    "changepassphrase CLI options",
			command: "util/seed/changepassphrase",
			arguments: []string{"--current_seed_passphrase_bin=cGFzc3dvcmQ=",
				"--current_seed=plastic:hollow:mansion:keep:into:cloth:awesome:salmon:reopen:inner:replace:dice:life:example:around",
				"--new_seed_passphrase=password",
			},
			expectedPayload: `{ "current_seed_passphrase_bin": "cGFzc3dvcmQ=", ` +
				`"current_seed": [ "plastic", "hollow", "mansion", "keep", "into", "cloth", "awesome", "salmon", "reopen", "inner", "replace", "dice", "life", "example", "around" ], ` +
				`"new_seed_passphrase": "password" }`,
		},
		{
			name:            "genseed CLI options",
			command:         "util/seed/create",
			arguments:       []string{"--seed_passphrase_bin=cGFzc3dvcmQ="},
			expectedPayload: `{ "seed_passphrase_bin": "cGFzc3dvcmQ=" }`,
		},
		//	test commands of "Lightning/Channel" group
		{
			name:    "openchannel CLI options",
			command: "lightning/channel/open",
			arguments: []string{"--node_pubkey=a0a1a2a3a4a5a6a7a8a9aa=", "--local_funding_amount=100000", "--push_sat=500000", "--target_conf=50",
				"sat_per_byte=250", "--private", "--min_htlc_msat=2048", "--remote_csv_delay=256", "--min_confs=5", "--close_address=C70534dd",
				"--amt=2050750", "--funding_txid_bytes=f0f1f2f3f4f5f6f7f8f9ff", "--output_index=282748", "--raw_key_bytes=1011121314151617181910",
				"--key_family=25", "--key_index=70", "--remote_key=e0e2e3e4e5e6e7e8e9ee", "--pending_chain_id=632d129", "--thaw_height=16274648",
				"--pending_chan_id=374649828379", "--base_psbt=3754629387292", "--no_publish", "--remote_max_value_in_flight_msat=100500",
				"--remote_max_htlcs=287", "--remote_max_csv=209",
			},
			expectedPayload: `{ "node_pubkey": "a0a1a2a3a4a5a6a7a8a9aa=", "local_funding_amount": 100000, "push_sat": 500000, "target_conf": 50, ` +
				`"private": true, "min_htlc_msat": 2048, "remote_csv_delay": 256, "min_confs": 5, "close_address": "C70534dd", ` +
				`"funding_shim": { "chan_point_shim": { "amt": 2050750, "chan_point": { "funding_txid_bytes": "f0f1f2f3f4f5f6f7f8f9ff", "output_index": 282748 }, "local_key": { "raw_key_bytes": "1011121314151617181910", ` +
				`"key_loc": { "key_family": 25, "key_index": 70 } }, "remote_key": "e0e2e3e4e5e6e7e8e9ee", "pending_chan_id": "374649828379", "thaw_height": 16274648 }, ` +
				`"psbt_shim": { "pending_chan_id": "374649828379", "base_psbt": "3754629387292", "no_publish": true } }, "remote_max_value_in_flight_msat": 100500, ` +
				`"remote_max_htlcs": 287 }`,
		},
		{
			name:            "closechannel CLI options",
			command:         "lightning/channel/close",
			arguments:       []string{"--funding_txid_str=a0a1a2a3a4a5a6a7a8a9aa", "--force", "--delivery_address=d0d1d2d3d4d5d6d7d8d9dd"},
			expectedPayload: `{ "channel_point": { "funding_txid_str": "a0a1a2a3a4a5a6a7a8a9aa" }, "force": true, "delivery_address": "d0d1d2d3d4d5d6d7d8d9dd" }`,
		},
		{
			name:            "abandonchannel CLI options",
			command:         "lightning/channel/abandon",
			arguments:       []string{"--funding_txid_str=a0a1a2a3a4a5a6a7a8a9aa", "--output_index=282748", "--pending_funding_shim_only"},
			expectedPayload: `{ "channel_point": { "funding_txid_str": "a0a1a2a3a4a5a6a7a8a9aa", "output_index": 282748 }, "pending_funding_shim_only": true }`,
		},
		{
			name:            "channelbalance CLI options",
			command:         "lightning/channel/balance",
			arguments:       []string{},
			expectedPayload: `{  }`,
		},
		{
			name:            "pendingchannels CLI options",
			command:         "lightning/channel/pending",
			arguments:       []string{},
			expectedPayload: `{  }`,
		},
		{
			name:            "listchannels CLI options",
			command:         "/lightning/channel",
			arguments:       []string{"--active_only", "--public_only", "--peer=a0a1a2a3a4a5a6a7a8a9aa="},
			expectedPayload: `{ "active_only": true, "public_only": true, "peer": "a0a1a2a3a4a5a6a7a8a9aa=" }`,
		},
		{
			name:            "closedchannels CLI options",
			command:         "/lightning/channel/closed",
			arguments:       []string{"--local_force", "--abandoned"},
			expectedPayload: `{ "local_force": true, "abandoned": true }`,
		},
		{
			name:            "getnetworkinfo CLI options",
			command:         "/lightning/channel/networkinfo",
			arguments:       []string{},
			expectedPayload: `{  }`,
		},
		{
			name:            "feereport CLI options",
			command:         "/lightning/channel/feereport",
			arguments:       []string{},
			expectedPayload: `{  }`,
		},
		{
			name:            "updatechanpolicy CLI options",
			command:         "/lightning/channel/policy",
			arguments:       []string{"--global", "--funding_txid_str=a0a1a2a3a4a5a6a7a8a9aa", "--output_index=282748", "--base_fee_msat=10", "--fee_rate=100"},
			expectedPayload: `{ "global": true, "chan_point": { "funding_txid_str": "a0a1a2a3a4a5a6a7a8a9aa", "output_index": 282748 }, "base_fee_msat": 10, "fee_rate": 100 }`,
		},
		//	test commands of "Lightning/Channel/Backup" group
		{
			name:            "exportchanbackup CLI options",
			command:         "/lightning/channel/backup/export",
			arguments:       []string{"--funding_txid_str=a0a1a2a3a4a5a6a7a8a9aa", "--output_index=282748"},
			expectedPayload: `{ "chan_point": { "funding_txid_str": "a0a1a2a3a4a5a6a7a8a9aa", "output_index": 282748 } }`,
		},
		{
			name:            "restorechanbackup CLI options",
			command:         "/lightning/channel/backup/restore",
			arguments:       []string{"--funding_txid_str=a0a1a2a3a4a5a6a7a8a9aa", "--output_index=282748", "--chan_backup=RW5jcnlwdGVkIENoYW4gQmFja3Vw"},
			expectedPayload: `{ "chan_backups": { "chan_backups": [ "chan_point": { "funding_txid_str": "a0a1a2a3a4a5a6a7a8a9aa", "output_index": 282748 }, "chan_backup": "RW5jcnlwdGVkIENoYW4gQmFja3Vw" ] } }`,
		},
		//	TODO: need to manage how to deal with fields with same name !
		/*
			{
				name:            "verifychanbackup CLI options",
				command:         "/lightning/channel/backup/verify",
				arguments:       []string{"--funding_txid_str=a0a1a2a3a4a5a6a7a8a9aa", "--output_index=282748", "--chanBackup=RW5jcnlwdGVkIENoYW4gQmFja3Vw"},
				expectedPayload: `{ "chan_backups": [ "chan_point": { "funding_txid_str": "a0a1a2a3a4a5a6a7a8a9aa", "output_index": 282748 }, "chanBackup": "RW5jcnlwdGVkIENoYW4gQmFja3Vw" ] }`,
			},
		*/
		//	test commands of "Lightning/Graph" group
		{
			name:            "describegraph CLI options",
			command:         "/lightning/graph",
			arguments:       []string{"--include_unannounced"},
			expectedPayload: `{ "include_unannounced": true }`,
		},
		{
			name:            "getnodemetrics CLI options",
			command:         "/lightning/graph/nodemetrics",
			arguments:       []string{"--UNKNOWN", "--BETWEENNESS_CENTRALITY"},
			expectedPayload: `{ "types": [ 0, 1 ] }`,
		},
		{
			name:            "getchaninfo CLI options",
			command:         "/lightning/graph/channel",
			arguments:       []string{"--chan_id=123456"},
			expectedPayload: `{ "chan_id": 123456 }`,
		},
		{
			name:            "getnodeinfo CLI options",
			command:         "/lightning/graph/nodeinfo",
			arguments:       []string{"--pub_key=a0a1a2a3a4a5a6a7a8a9aa", "--include_channels"},
			expectedPayload: `{ "pub_key": "a0a1a2a3a4a5a6a7a8a9aa", "include_channels": true }`,
		},
		//	test commands of "Lightning/Invoice" group

		//	TODO: need to manage how to deal with fields with same name !
		/*
			{
				name:            "addinvoice CLI options",
				command:         "/lightning/invoice/create",
				arguments:       []string{"--memo=xpto", "--r_preimage=0123456789abcdef0123456789abcdef", "--r_hash=00112233445566778899", "--value=10", "--expiry=3600"},
				expectedPayload: `{ "memo": "xpto", "r_preimage": "0123456789abcdef0123456789abcdef", "r_hash": "00112233445566778899", "value": 10, "expiry": 3600 }`,
			},
		*/
		{
			name:            "lookupinvoice CLI options",
			command:         "/lightning/invoice/lookup",
			arguments:       []string{"--r_hash=00112233445566778899"},
			expectedPayload: `{ "r_hash": "00112233445566778899" }`,
		},
		{
			name:            "listinvoices CLI options",
			command:         "/lightning/invoice",
			arguments:       []string{"--index_offset=1", "--num_max_invoices=10", "--reversed"},
			expectedPayload: `{ "index_offset": 1, "num_max_invoices": 10, "reversed": true }`,
		},
		{
			name:            "decodepayreq CLI options",
			command:         "/lightning/invoice/decodepayreq",
			arguments:       []string{"--pay_req=123456"},
			expectedPayload: `{ "pay_req": "123456" }`,
		},
		//	test commands of "Lightning/Payment" group
		{
			name:            "sendpayment CLI options",
			command:         "/lightning/payment/send",
			arguments:       []string{"--dest=010203040506070809", "--payment_hash=02e28f38ad50869fd3f3d75147d69bc637090aa9b5013ee49a65c0dda2bf0ab51e", "--amt=100000"},
			expectedPayload: `{ "dest": "010203040506070809", "amt": 100000, "payment_hash": "02e28f38ad50869fd3f3d75147d69bc637090aa9b5013ee49a65c0dda2bf0ab51e" }`,
		},
		/*
			{
				name:            "payinvoice CLI options",
				command:         "/lightning/payment/payinvoice",
				arguments:       []string{""},
				expectedPayload: `{  }`,
			},
		*/
		{
			name:            "sendtoroute CLI options",
			command:         "/lightning/payment/sendtoroute",
			arguments:       []string{""},
			expectedPayload: `{  }`,
		},
		{
			name:            "listpayments CLI options",
			command:         "/lightning/payment",
			arguments:       []string{""},
			expectedPayload: `{  }`,
		},
		/*
			{
				name:            "trackpayment CLI options",
				command:         "/lightning/payment/track",
				arguments:       []string{""},
				expectedPayload: `{  }`,
			},
		*/
		{
			name:            "queryroutes CLI options",
			command:         "/lightning/payment/queryroutes",
			arguments:       []string{""},
			expectedPayload: `{  }`,
		},
		{
			name:            "fwdinghistory CLI options",
			command:         "/lightning/payment/fwdinghistory",
			arguments:       []string{""},
			expectedPayload: `{  }`,
		},
		{
			name:            "querymc CLI options",
			command:         "/lightning/payment/querymc",
			arguments:       []string{""},
			expectedPayload: `{  }`,
		},
		{
			name:            "queryprob CLI options",
			command:         "/lightning/payment/queryprob",
			arguments:       []string{""},
			expectedPayload: `{  }`,
		},
		{
			name:            "resetmc CLI options",
			command:         "/lightning/payment/resetmc",
			arguments:       []string{""},
			expectedPayload: `{  }`,
		},
		{
			name:            "buildrout CLI options",
			command:         "/lightning/payment/buildroute",
			arguments:       []string{""},
			expectedPayload: `{  }`,
		},
		/*
			executeCommand '' 'POST' '/lightning/payment' "{ \"paymentHash\": \"${RHASH}\", \"amt\": 100000, \"dest\": \"${TARGET_WALLET}\" }"
			executeCommand '' 'POST' '/lightning/payment' '{ "paymentHash": "02e28f38ad50869fd3f3d75147d69bc637090aa9b5013ee49a65c0dda2bf0ab51e", "route": { "hops": { "chanId": "xpto"} } }'
			executeCommand '' 'POST' '/lightning/payment' '{ "indexOffset": 1, "maxPayments": 10, "includeIncomplete": true }'
			executeCommand '' 'POST' '/lightning/payment' '{ "indexOffset": 1, "maxPayments": 10, "includeIncomplete": true }'
			executeCommand '' 'POST' '/lightning/payment' "{  }"
			executeCommand '' 'POST' '/lightning/payment' '{ "indexOffset": 0, "numMaxEvents": 25 }'
			executeCommand '' 'GET' '/lightning/payment'
			executeCommand '' 'POST' '/lightning/payment' "{ \"fromNode\": \"${FROM_NODE}\", \"toNode\": \"${TO_NODE}\", \"amtMsat\": \"${AMOUNT}\" }"
			executeCommand '' 'GET' '/lightning/payment'
			executeCommand '' 'POST' '/lightning/payment' '{ "amtMsat": 0, "hopPubkeys": [ "01020304", "02030405", "03040506" ] }'

			executeCommand 'connect' 'POST' '/lightning/peer/connect' "{ \"addr\": { \"pubkey\": \"${PUBLIC_KEY}\", \"host\": \"192.168.40.1:8080\" } }"
			executeCommand 'disconnect' 'POST' '/lightning/peer/disconnect' "{ \"pubkey\": \"${PUBLIC_KEY}\" }"
			executeCommand 'listpeers' 'GET' '/lightning/peer'

			executeCommand 'bcasttransaction' 'POST' '/neutrino/bcasttransaction' "{ \"tx\": \"${TXID}\" }"
			executeCommand 'estimatefee' 'POST' '/neutrino/estimatefee' "{ \"AddrToAmount\": { \"${TARGET_WALLET}\": 100000 } }"

			executeCommand 'walletbalance' 'GET' '/wallet/balance'
			executeCommand 'changePassphrase' 'POST' '/wallet/changepassphrase' "{ \"current_passphrase\": \"${PASSPHRASE}\", \"new_passphrase\": \"${NEW_PASSPHRASE}\" }"
			executeCommand 'checkPassphrase' 'POST' '/wallet/checkpassphrase' "{ \"wallet_passphrase\": \"${WALLET_PASSPHRASE}\" }"
			executeCommand 'create_wallet' 'POST' '/wallet/create' "{ \"wallet_passphrase\": \"${PASSPHRASE}\", \"wallet_seed\": ${WALLET_SEED}, \"seed_passphrase_bin\": \"${SEED_PASSPHRASE}\" }"
			executeCommand 'getsecret' 'POST' '/wallet/getsecret' '{ "name": "Isaac Assimov" }'
			executeCommand 'getwalletseed' 'GET' '/wallet/seed'
			executeCommand 'unlock_wallet' 'POST' '/wallet/unlock' "{ \"wallet_passphrase\": \"${WALLET_PASSPHRASE}\" }"

			executeCommand 'getaddressbalances' 'POST' '/wallet/address/balances' '{  }'
			executeCommand 'newaddress' 'POST' '/wallet/address/create' '{  }'
			executeCommand 'dumpprivkey' 'POST' '/wallet/address/dumpprivkey' '{ "address": "pkt1q85n69mzthdxlwutn6dr6f7kwyd9nv8ulasdaqz" }'
			executeCommand 'importprivkey' 'POST' '/wallet/address/import' "{ \"privateKey\": \"${WALLET_PRIVKEY}xx\", \"rescan\": true }"
			executeCommand 'signmessage' 'POST' '/wallet/address/signmessage' '{ "msg": "testing pld REST endpoints" }'

			executeCommand 'getnetworkstewardvote' 'GET' '/wallet/networkstewardvote'
			executeCommand 'setnetworkstewardvote' 'POST' '/wallet/networkstewardvote/set' '{ "voteAgainst": "0", "voteFor": "1" }'

			executeCommand 'gettransaction' 'POST' '/wallet/transaction' ''
			executeCommand 'createtransaction' 'POST' '/wallet/transaction/create' "{ \"toAddress\": \"${TARGET_WALLET}\", \"amount\": ${AMOUNT} }"
			executeCommand 'query' 'POST' '/wallet/transaction/query' '{  }'
			executeCommand 'sendcoins' 'POST' '/wallet/transaction/sendcoins' "{ \"addr\": \"${TARGET_WALLET}\", \"amount\": ${AMOUNT} }"
			executeCommand 'sendfrom' 'POST' '/wallet/transaction/sendfrom' '{  }'
			executeCommand 'sendmany' 'POST' '/wallet/transaction/sendmany' "{ \"AddrToAmount\": { \"${TARGET_WALLET}\": ${AMOUNT} } }"

			executeCommand 'listunspent' 'POST' '/wallet/unspent' '{ "minConfs": 1, "maxConfs": 100 }'
			executeCommand 'resync' 'POST' '/wallet/unspent/resync' '{  }'
			executeCommand 'stopresync' 'GET' '/wallet/unspent/stopresync' ''

			executeCommand 'listlockunspent' 'GET' '/wallet/unspent/lock'
			executeCommand 'lockunspent' 'POST' '/wallet/unspent/lock/create' '{ "lockname": "secure vault", "unlock": false, "transactions": [ { "txid": "934095dc4afa8d4b5d43732a96e78e11c0e88defdaab12d946f525e54478938f" } ] }' ''

			executeCommand 'show' 'POST' '/wtclient/tower' '{  }'
			executeCommand 'show' 'POST' '/wtclient/tower/create' '{  }'
			executeCommand 'show' 'POST' '/wtclient/tower/getinfo' '{  }'
			executeCommand 'show' 'POST' '/wtclient/tower/policy' '{  }'
			executeCommand 'show' 'POST' '/wtclient/tower/remove' '{  }'
			executeCommand 'show' 'POST' '/wtclient/tower/stats' '{  }'
		*/
	}

	for _, testCase := range testCases {

		t.Logf(">>> Test the conversion of all CLI arguments into REST JSon request payload")

		t.Run(testCase.name, func(t *testing.T) {
			want := testCase.expectedPayload
			got, err := formatRequestPayload(testCase.command, testCase.arguments)
			if err != nil {
				t.Errorf("Unexpected error formatting the payload: %s", err)
			}
			if want != got {
				t.Errorf("Error formatting the payload: got '%s', want '%s'", got, want)
			}
		})
	}
}
