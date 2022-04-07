#!  /usr/bin/bash

################################################################################
#   smoke tests for pld REST endpoints
################################################################################

export  PLD_REST_SERVER='http://localhost:8080'
export  PLD_REST_CONTEXT='/api/v1'
export  REST_ERRORS_FILE='./rest.err'
export  JSON_OUTPUT=''
export  WALLET_PASSPHRASE='w4ll3tP@sswd'
export  VERBOSE='false'

#   color commands
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
LIGHTGRAY='\033[0;37m'
NOCOLOR='\033[0m'

#   use curl to execute a command
executeCommand() {
    local COMMAND="${1}"
    local HTTP_METHOD="${2}"
    local URI="${3}"
    local PAYLOAD="${4}"


    JSON_OUTPUT=""
    HTTP_RESPONSE_CODE=""

    if [ "${HTTP_METHOD}" == "GET" ]
    then
        if [ "${VERBOSE}" == 'true' ]
        then
            echo -e "[trace] ${LIGHTGRAY}curl \"${PLD_REST_SERVER}${PLD_REST_CONTEXT}${URI}\"${NOCOLOR}"
        fi

        OUTPUT=$( curl --write-out '|%{response_code}' "${PLD_REST_SERVER}${PLD_REST_CONTEXT}${URI}" 2>> ${REST_ERRORS_FILE} | tr -d '\n' )
        JSON_OUTPUT=$( echo "${OUTPUT}" | perl -ne 'if( $_ =~ /^(.+)\|(\d{3})$/ ) { print qq/$1\n/; }' )
        HTTP_RESPONSE_CODE=$( echo "${OUTPUT}" | perl -ne 'if( $_ =~ /^(.+)\|(\d{3})$/ ) { print qq/$2\n/; }' )
    elif [ "${HTTP_METHOD}" == "POST" ]
    then
        if [ "${VERBOSE}" == 'true' ]
        then
            echo -e "[trace] ${LIGHTGRAY}curl -H \"Content-Type: application/json\" -X POST -d '${PAYLOAD}' \"${PLD_REST_SERVER}${PLD_REST_CONTEXT}${URI}\"${NOCOLOR}"
        fi

        OUTPUT=$( curl --write-out '|%{response_code}' -H "Content-Type: application/json" -X POST -d "${PAYLOAD}" "${PLD_REST_SERVER}${PLD_REST_CONTEXT}${URI}" 2>> ${REST_ERRORS_FILE} | tr -d '\n' )
        JSON_OUTPUT=$( echo "${OUTPUT}" | perl -ne 'if( $_ =~ /^(.+)\|(\d{3})$/ ) { print qq/$1\n/; }' )
        HTTP_RESPONSE_CODE=$( echo "${OUTPUT}" | perl -ne 'if( $_ =~ /^(.+)\|(\d{3})$/ ) { print qq/$2\n/; }' )
    else
        echo -e "${RED}error: invalid HTTP method \"${HTTP_METHOD}\"${NOCOLOR}"
        return 1
    fi

    if [ "${VERBOSE}" == 'true' ]
    then
        echo -e "[trace] ${LIGHTGRAY}response: ${JSON_OUTPUT}${NOCOLOR}"
    fi

    if [ "${HTTP_RESPONSE_CODE}" == "200" ]
    then
        echo -e ">>> ${CYAN}[${COMMAND}]${NOCOLOR}: ${GREEN}command successfully executed${NOCOLOR}"
    else
        ERROR_MESSAGE=$( echo "${JSON_OUTPUT}" | jq '.message' )
        echo -e ">>> ${CYAN}[${COMMAND}]${NOCOLOR}: ${RED}error running pld command: HTTP response: ${HTTP_RESPONSE_CODE}\n\t${ERROR_MESSAGE}${NOCOLOR}"
        JSON_OUTPUT=''
        return 1
    fi
}

#   use jq to filter results of previously executed command
showCommandResult() {
    local TITLE="${1}"
    local FILTER="${2}"

    if [ ! -z "${JSON_OUTPUT}" ]
    then
        if [ -z "${FILTER}" ]
        then
            RESULT="${JSON_OUTPUT}"
        else
            RESULT=$( echo "${JSON_OUTPUT}" | jq "${FILTER}" )
        fi
        echo -e "    >>> ${TITLE}: ${LIGHTGRAYNOCOLOR}${RESULT}${NOCOLOR}"
    fi
}

#   use jq to filter results of previously executed command
getCommandResult() {
    local FILTER="${1}"

    if [ ! -z "${JSON_OUTPUT}" ]
    then
        if [ ! -z "${FILTER}" ]
        then
            echo -e "$( echo "${JSON_OUTPUT}" | jq "${FILTER}" )"
        fi
    fi
}

#   splash screen
echo ">>>>> Testing pld REST endpoints"
echo

#   check if curl is available
OUTPUT=$( which curl 2> /dev/null )
if [ $? -ne 0 ]
then
    exit "error: 'curl' is required to run this script"
fi

#   check if jq is available
OUTPUT=$( which jq 2> /dev/null )
if [ $? -ne 0 ]
then
    exit "error: 'jq' is required to run this script"
fi

#   parse CLI arguments
while [ true ]
do
    ARG=${1}

    if [ -z "${ARG}" ]
    then
        break
    fi

    if [ "${ARG}" == "--verbose" ]
    then
        VERBOSE='true'
    fi

    shift
done

cat > /dev/null <<PLD_HELP
{
   "name": "pld - Lightning Network Daemon REST interface (pld)",
   "description": [
      "General information about PLD"
   ],
   "category": {
      "Lightning": {
         "description": [
            "The Lightning Network component of the wallet"
         ],
         "subcategory": {
            "Channel": {
               "description": [
                  "Management of lightning channels to direct peers of this pld node"
               ],
               "endpoints": {
                  "/api/v1/lightning/channel": "List all open channels",
                  "/api/v1/lightning/channel/abandon": "Abandons an existing channel",
                  "/api/v1/lightning/channel/balance": "Returns the sum of the total available channel balance across all open channels",
                  "/api/v1/lightning/channel/close": "Close an existing channel",
                  "/api/v1/lightning/channel/closed": "List all closed channels",
                  "/api/v1/lightning/channel/feereport": "Display the current fee policies of all active channels",
                  "/api/v1/lightning/channel/networkinfo": "Get statistical information about the current state of the network",
                  "/api/v1/lightning/channel/open": "Open a channel to a node or an existing peer",
                  "/api/v1/lightning/channel/pending": "Display information pertaining to pending channels",
                  "/api/v1/lightning/channel/policy": "Update the channel policy for all channels, or a single channel"
               },
               "subcategory": {
                  "Backup": {
                     "description": [
                        "Backup and recovery of the state of active Lightning channels",
                        "to and from this pld node"
                     ],
                     "endpoints": {
                        "/api/v1/lightning/channel/backup/export": "Obtain a static channel back up for a selected channels, or all known channels",
                        "/api/v1/lightning/channel/backup/restore": "Restore an existing single or multi-channel static channel backup",
                        "/api/v1/lightning/channel/backup/verify": "Verify an existing channel backup"
                     }
                  }
               }
            },
            "Graph": {
               "description": [
                  "Information about the global known Lightning Network"
               ],
               "endpoints": {
                  "/api/v1/lightning/graph": "Describe the network graph",
                  "/api/v1/lightning/graph/channel": "Get the state of a channel",
                  "/api/v1/lightning/graph/nodeinfo": "Get information on a specific node",
                  "/api/v1/lightning/graph/nodemetrics": "Get node metrics"
               }
            },
            "Invoice": {
               "description": [
                  "Management of invoices which are used to request payment over Lightning"
               ],
               "endpoints": {
                  "/api/v1/lightning/invoice": "List all invoices currently stored within the database. Any active debug invoices are ignored",
                  "/api/v1/lightning/invoice/create": "Add a new invoice",
                  "/api/v1/lightning/invoice/decodepayreq": "Decode a payment request",
                  "/api/v1/lightning/invoice/lookup": "Lookup an existing invoice by its payment hash"
               }
            },
            "Payment": {
               "description": [
                  "Lightning network payments which have been made, or have been forwarded, through this node"
               ],
               "endpoints": {
                  "/api/v1/lightning/payment": "List all outgoing payments",
                  "/api/v1/lightning/payment/buildroute": "Build a route from a list of hop pubkeys",
                  "/api/v1/lightning/payment/fwdinghistory": "Query the history of all forwarded HTLCs",
                  "/api/v1/lightning/payment/payinvoice": "Pay an invoice over lightning",
                  "/api/v1/lightning/payment/querymc": "Query the internal mission control state",
                  "/api/v1/lightning/payment/queryprob": "Estimate a success probability",
                  "/api/v1/lightning/payment/queryroutes": "Query a route to a destination",
                  "/api/v1/lightning/payment/resetmc": "Reset internal mission control state",
                  "/api/v1/lightning/payment/send": "Send a payment over lightning",
                  "/api/v1/lightning/payment/sendtoroute": "Send a payment over a predefined route",
                  "/api/v1/lightning/payment/track": "Track progress of an existing payment"
               }
            },
            "Peer": {
               "description": [
                  "Connections to other nodes in the Lightning Network"
               ],
               "endpoints": {
                  "/api/v1/lightning/peer": "List all active, currently connected peers",
                  "/api/v1/lightning/peer/connect": "Connect to a remote pld peer",
                  "/api/v1/lightning/peer/disconnect": "Disconnect a remote pld peer identified by public key"
               }
            }
         }
      },
      "Meta": {
         "description": [
            "API endpoints which are relevant to the entire pld node, not any specific part"
         ],
         "endpoints": {
            "/api/v1/meta/debuglevel": "Set the debug level",
            "/api/v1/meta/getinfo": "Returns basic information related to the active daemon",
            "/api/v1/meta/stop": "Stop and shutdown the daemon",
            "/api/v1/meta/version": "Display pldctl and pld version info"
         }
      },
      "Neutrino": {
         "description": [
            "Management of the Neutrino interface which is used to communicate with the p2p nodes in the network"
         ],
         "endpoints": {
            "/api/v1/neutrino/bcasttransaction": "Broadcast a transaction onchain",
            "/api/v1/neutrino/estimatefee": "Get fee estimates for sending bitcoin on-chain to multiple addresses"
         }
      },
      "Util": {
         "description": [
            "Stateless utility functions which do not affect, not query, the node in any way"
         ],
         "subcategory": {
            "Seed": {
               "description": [
                  "Manipulation of mnemonic seed phrases which represent wallet keys"
               ],
               "endpoints": {
                  "/api/v1/util/seed/changepassphrase": "Alter the passphrase which is used to encrypt a wallet seed",
                  "/api/v1/util/seed/create": "Create a secret seed"
               }
            }
         }
      },
      "Wallet": {
         "description": [
            "APIs for management of on-chain (non-Lightning) payments,",
            "seed export and recovery, and on-chain transaction detection"
         ],
         "endpoints": {
            "/api/v1/wallet/balance": "Compute and display the wallet's current balance",
            "/api/v1/wallet/changepassphrase": "Change an encrypted wallet's password at startup",
            "/api/v1/wallet/checkpassphrase": "Check the wallet's password",
            "/api/v1/wallet/create": "Initialize a wallet when starting lnd for the first time",
            "/api/v1/wallet/getsecret": "Get a secret seed",
            "/api/v1/wallet/seed": "Get the wallet seed words for this wallet",
            "/api/v1/wallet/unlock": "Unlock an encrypted wallet at startup"
         },
         "subcategory": {
            "Address": {
               "description": [
                  "Management of individual wallet addresses"
               ],
               "endpoints": {
                  "/api/v1/wallet/address/balances": "Compute and display balances for each address in the wallet",
                  "/api/v1/wallet/address/create": "Generates a new address",
                  "/api/v1/wallet/address/dumpprivkey": "Returns the private key in WIF encoding that controls some wallet address",
                  "/api/v1/wallet/address/import": "Imports a WIF-encoded private key to the 'imported' account",
                  "/api/v1/wallet/address/signmessage": "Signs a message using the private key of a payment address"
               }
            },
            "Network Steward Vote": {
               "description": [
                  "Control how this wallet votes on PKT Network Steward"
               ],
               "endpoints": {
                  "/api/v1/wallet/networkstewardvote": "Find out how the wallet is currently configured to vote in a network steward election",
                  "/api/v1/wallet/networkstewardvote/set": "Configure the wallet to vote for a network steward when making payments (note: payments to segwit addresses cannot vote)"
               }
            },
            "Transaction": {
               "description": [
                  "Create and manage on-chain transactions with the wallet"
               ],
               "endpoints": {
                  "/api/v1/wallet/transaction": "Returns a JSON object with details regarding a transaction relevant to this wallet",
                  "/api/v1/wallet/transaction/create": "Create a transaction but do not send it to the chain",
                  "/api/v1/wallet/transaction/query": "List transactions from the wallet",
                  "/api/v1/wallet/transaction/sendcoins": "Send bitcoin on-chain to an address",
                  "/api/v1/wallet/transaction/sendfrom": "Authors, signs, and sends a transaction that outputs some amount to a payment address",
                  "/api/v1/wallet/transaction/sendmany": "Send bitcoin on-chain to multiple addresses"
               }
            },
            "Unspent": {
               "description": [
                  "Detected unspent transactions associated with one of our wallet addresses"
               ],
               "endpoints": {
                  "/api/v1/wallet/unspent": "List utxos available for spending",
                  "/api/v1/wallet/unspent/resync": "Scan over the chain to find any transactions which may not have been recorded in the wallet's database",
                  "/api/v1/wallet/unspent/stopresync": "Stop a re-synchronization job before it's completion"
               },
               "subcategory": {
                  "Lock": {
                     "description": [
                        "Manipulation of unspent outputs which are 'locked'",
                        "and therefore will not be used to source funds for any transaction"
                     ],
                     "endpoints": {
                        "/api/v1/wallet/unspent/lock": "Returns a JSON array of outpoints marked as locked (with lockunspent) for this wallet session",
                        "/api/v1/wallet/unspent/lock/create": "Locks or unlocks an unspent output"
                     }
                  }
               }
            }
         }
      },
      "Watchtower": {
         "description": [
            "Interact with the watchtower client"
         ],
         "endpoints": {
            "/api/v1/wtclient/tower": "Display information about all registered watchtowers",
            "/api/v1/wtclient/tower/create": "Register a watchtower to use for future sessions/backups",
            "/api/v1/wtclient/tower/getinfo": "Display information about a specific registered watchtower",
            "/api/v1/wtclient/tower/policy": "Display the active watchtower client policy configuration",
            "/api/v1/wtclient/tower/remove": "Remove a watchtower to prevent its use for future sessions/backups",
            "/api/v1/wtclient/tower/stats": "Display the session stats of the watchtower client"
         }
      }
   }
}
PLD_HELP

#
#   test commands of "meta" group
#

echo -e ">>> Group ${CYAN}[meta]${NOCOLOR} API endpoints which are relevant to the entire pld node, not any specific part"
echo

executeCommand 'debuglevel' 'POST' '/meta/debuglevel' '{ "show": true, "level_spec": "debug" }'
showCommandResult 'subsystems' '.subSystems'

executeCommand 'getinfo' 'GET' '/meta/getinfo'
showCommandResult '#neutrino peers' '.neutrino.peers | length'

#   this needs be the last of all tests
#executeCommand 'stop' 'GET' '/meta/stop'
#showCommandResult 'result' ''

executeCommand 'version' 'GET' '/meta/version'
showCommandResult 'result' ''

echo "----------"
echo

#
#   test commands of "util/seed" group
#

echo -e ">>> Group ${CYAN}[util/seed]${NOCOLOR} Manipulation of mnemonic seed phrases which represent wallet keys"
echo

executeCommand 'changepassphrase' 'POST' '/util/seed/changepassphrase' '{ "current_seed_passphrase_bin": "cGFzc3dvcmQ=", "current_seed": [ "plastic",  "hollow",  "mansion",  "keep",  "into",  "cloth",  "awesome",  "salmon",  "reopen",  "inner",  "replace",  "dice",  "life",  "example",  "around" ], "new_seed_passphrase": "password" }'
showCommandResult 'new ciphered seed' '.seed'

#   this test is meant to fail, since seed creation can only be ordered before wallet's creation
executeCommand 'genseed' 'POST' '/util/seed/create' '{ "seed_passphrase_bin": "cGFzc3dvcmQ=" }'
showCommandResult 'ciphered seed' '.seed'

echo "----------"
echo

#
#   test commands of "Lightning/Channel" group
#

echo -e ">>> Group ${CYAN}[lightning/channel]${NOCOLOR} Management of lightning channels to direct peers of this pld node"
echo

#   fetch a public key for the channels tests
executeCommand 'describegraph' 'POST' '/lightning/graph' '{ "includeUnannounced": true }'
showCommandResult 'public key' '.nodes | .[0] | .pubKey'
PUBLIC_KEY="$( getCommandResult '.nodes | .[0] | .pubKey ' | tr -d '\"' | cut --characters=3- )"

AMOUNT="100000"
executeCommand 'openchannel' 'POST' '/lightning/channel/open' "{ \"node_pubkey\": \"${PUBLIC_KEY}\", \"local_funding_amount\": ${AMOUNT} }"
showCommandResult 'result' ''

CHANNEL_POINT="XPTO"
executeCommand 'closechannel' 'POST' '/lightning/channel/close' "{ \"channel_point\": \"${CHANNEL_POINT}\" }"
showCommandResult 'result' ''

FUNDING_TXID="934095dc4afa8d4b5d43732a96e78e11c0e88defdaab12d946f525e54478938f"
executeCommand 'abandonchannel' 'POST' '/lightning/channel/abandon' "{ \"channelPoint\": { \"funding_txid_str\": \"${FUNDING_TXID}\" } }"
showCommandResult 'result' ''

executeCommand 'channelbalance' 'GET' '/lightning/channel/balance'
showCommandResult 'channel balance' '.balance'

executeCommand 'pendingchannels' 'GET' '/lightning/channel/pending'
showCommandResult '#pending open channels' '.pendingOpenChannels | length'
showCommandResult '#pending closing channels' '.pendingClosingChannels | length'
showCommandResult 'limbo balance' '.totalLimboBalance'

executeCommand 'listchannels' 'POST' '/lightning/channel' '{  }'
showCommandResult '#open channels' '.channels | length'

executeCommand 'closedchannels' 'POST' '/lightning/channel/closed' '{  }'
showCommandResult '#closed channels' '.channels | length'

executeCommand 'getnetworkinfo' 'GET' '/lightning/channel/networkinfo'
showCommandResult 'nodes' '.numNodes'
showCommandResult 'channels' '.numChannels'

executeCommand 'feereport' 'GET' '/lightning/channel/feereport'
showCommandResult '#channel fees' '.channelFees | length'
showCommandResult 'week fee sum' '.weekFeeSum'

executeCommand 'updatechanpolicy' 'POST' '/lightning/channel/policy' '{ "baseFeeMsat": 10, "feeRate": 10, "timeLockDelta": 20, "maxHtlcMsat": 30, "minHtlcMsat": 1, "minHtlcMsatSpecified": false }'
showCommandResult 'result' ''

echo "----------"
echo

#
#   test commands of "Lightning/Channel/Backup" group
#

echo -e ">>> Group ${CYAN}[lightning/channel/backup]${NOCOLOR} Backup and recovery of the state of active Lightning channels to and from this pld node"
echo

executeCommand 'exportchanbackup' 'POST' '/lightning/channel/backup/export' "{ \"chanPoint\": { \"funding_txid_str\": \"${FUNDING_TXID}\" } }"
showCommandResult 'result' '.multi_chan_backup.multi_chan_backup'

executeCommand 'restorechanbackup' 'POST' '/lightning/channel/backup/restore' "{ \"chanBackups\": [ { \"chanPoint\": { \"fundingTxidStr\": \"${FUNDING_TXID}\", \"outputIndex\": 1000 }, \"chanBackup\": \"RW5jcnlwdGVkIENoYW4gQmFja3Vw\" } ] }"
showCommandResult 'result' ''

executeCommand 'verifychanbackup' 'POST' '/lightning/channel/backup/verify' "{ \"singleChanBackups\": { \"chanBackups\": [ { \"chanPoint\": { \"fundingTxidStr\": \"${FUNDING_TXID}\", \"outputIndex\": 1000 }, \"chanBackup\": \"RW5jcnlwdGVkIENoYW4gQmFja3Vw\" } ] } }"
showCommandResult 'result' ''

echo "----------"
echo

#
#   test commands of "Lightning/Graph" group
#

echo -e ">>> Group ${CYAN}[lightning/graph]${NOCOLOR} Information about the global known Lightning Network"
echo

#   fetch a public key for the graph tests
executeCommand 'describegraph' 'POST' '/lightning/graph' '{ "includeUnannounced": true }'
showCommandResult 'last update' '.nodes | .[0] | .lastUpdate '
showCommandResult 'public key' '.nodes | .[0] | .pubKey'
PUBLIC_KEY="$( getCommandResult '.nodes | .[0] | .pubKey ' | tr -d '\"' )"

executeCommand 'getnodemetrics' 'POST' '/lightning/graph/nodemetrics' '{ "types": [ 0, 1 ] }'
showCommandResult 'betweenness centrality' '.betweennessCentrality'

CHAN_ID=123
executeCommand 'getchaninfo' 'POST' '/lightning/graph/channel' "{ \"chanId\": ${CHAN_ID} }"
showCommandResult 'result' ''

executeCommand 'getnodeinfo' 'POST' '/lightning/graph/nodeinfo' "{ \"pubKey\": \"${PUBLIC_KEY}\", \"includeChannels\": true }"
showCommandResult 'last update' '.node.lastUpdate'
showCommandResult '#channels' '.node.channels | length'

echo "----------"
echo

#
#   test commands of "Lightning/Invoice" group
#

echo -e ">>> Group ${CYAN}[lightning/invoice]${NOCOLOR} Management of invoices which are used to request payment over Lightning"
echo

executeCommand 'addinvoice' 'POST' '/lightning/invoice/create' '{ "memo": "xpto", "value": 10, "expiry": 3600 }'
showCommandResult 'rHash' '.rHash'
showCommandResult 'payment request' '.paymentRequest'
RHASH="$( getCommandResult '.rHash' | tr -d '\"' )"
PAYREQ="$( getCommandResult '.paymentRequest' | tr -d '\"' )"

executeCommand 'lookupinvoice' 'POST' '/lightning/invoice/lookup' "{ \"rHash\": \"${RHASH}\" }"
showCommandResult 'last update' '.lastUpdate'
showCommandResult 'index' '.addIndex'
showCommandResult 'state' '.state'

executeCommand 'listinvoices' 'POST' '/lightning/invoice' '{ "indexOffset": 1, "numMaxInvoices": 10 }'
showCommandResult '#invoices' '.invoices | length'

executeCommand 'decodepayreq' 'POST' '/lightning/invoice/decodepayreq' "{ \"payReq\": \"${PAYREQ}\" }"
showCommandResult 'destination' '.destination'
showCommandResult 'payment Hash' '.paymentHash'
showCommandResult '#satoshis' '.numSatoshis'

echo "----------"
echo

#
#   test commands of "Lightning/Payment" group
#

echo -e ">>> Group ${CYAN}[lightning/payment]${NOCOLOR} Lightning network payments which have been made, or have been forwarded, through this node"
echo

executeCommand 'sendpayment' 'POST' '/lightning/payment/send' "{ \"paymentHash\": \"${RHASH}\", \"amt\": 100000, \"dest\": \"${TARGET_WALLET}\" }"
showCommandResult 'result' ''

executeCommand 'payinvoice' 'POST' '/lightning/payment/payinvoice' "{ \"paymentHash\": \"${RHASH}\", \"amt\": 100000, \"dest\": \"${TARGET_WALLET}\" }"
showCommandResult 'result' ''

executeCommand 'sendtoroute' 'POST' '/lightning/payment/sendtoroute' '{ "paymentHash": "02e28f38ad50869fd3f3d75147d69bc637090aa9b5013ee49a65c0dda2bf0ab51e", "route": { "hops": { "chanId": "xpto"} } }'
showCommandResult 'result' ''

executeCommand 'listpayments' 'POST' '/lightning/payment' '{ "indexOffset": 1, "maxPayments": 10, "includeIncomplete": true }'
showCommandResult '#payments' '.payments | length'

executeCommand 'trackpayment' 'POST' '/lightning/payment/track' '{ "indexOffset": 1, "maxPayments": 10, "includeIncomplete": true }'
showCommandResult '#payments' '.payments | length'

executeCommand 'queryroutes' 'POST' '/lightning/payment/queryroutes' "{  }"
showCommandResult 'result' ''

executeCommand 'fwdinghistory' 'POST' '/lightning/payment/fwdinghistory' '{ "indexOffset": 0, "numMaxEvents": 25 }'
showCommandResult '#forwarding events' '.forwardingEvents | length'

executeCommand 'querymc' 'GET' '/lightning/payment/querymc'
showCommandResult 'result' ''

FROM_NODE="01020304"
TO_NODE="02030405"
AMOUNT="100000"
executeCommand 'queryprob' 'POST' '/lightning/payment/queryprob' "{ \"fromNode\": \"${FROM_NODE}\", \"toNode\": \"${TO_NODE}\", \"amtMsat\": \"${AMOUNT}\" }"
showCommandResult 'result' ''

executeCommand 'resetmc' 'GET' '/lightning/payment/resetmc'
showCommandResult 'result' ''

executeCommand 'buildroute' 'POST' '/lightning/payment/buildroute' '{ "amtMsat": 0, "hopPubkeys": [ "01020304", "02030405", "03040506" ] }'
showCommandResult 'result' ''

echo "----------"
echo

#
#   test commands of "Lightning/Peer" group
#

echo -e ">>> Group ${CYAN}[lightning/peer]${NOCOLOR} Connections to other nodes in the Lightning Network"
echo

executeCommand 'connect' 'POST' '/lightning/peer/connect' "{ \"addr\": { \"pubkey\": \"${PUBLIC_KEY}\", \"host\": \"192.168.40.1:8080\" } }"
showCommandResult 'result' ''

executeCommand 'disconnect' 'POST' '/lightning/peer/disconnect' "{ \"pubkey\": \"${PUBLIC_KEY}\" }"
showCommandResult 'result' ''

executeCommand 'listpeers' 'GET' '/lightning/peer'
showCommandResult 'result' ''
showCommandResult '#peers' '.peers | length'

echo "----------"
echo

#
#   test commands of "Neutrino" group
#

echo -e ">>> Group ${CYAN}[neutrino]${NOCOLOR} Management of the Neutrino interface which is used to communicate with the p2p nodes in the network"
echo

#   fetch a transaction ID for the neutrino tests
TARGET_WALLET="pkt1q07ly7r47ss4drsvt2zq9zkcstksrq2dap3x0yw"
executeCommand 'sendmany' 'POST' '/api/v1/transaction/sendmany' "{ \"AddrToAmount\": { \"${TARGET_WALLET}\": 100000 } }"
showCommandResult 'result' ''
showCommandResult 'transaction ID' '.txid'
TXID="$( getCommandResult '.txid' | tr -d '\"' )"

executeCommand 'bcasttransaction' 'POST' '/neutrino/bcasttransaction' "{ \"tx\": \"${TXID}\" }"
showCommandResult 'result' ''
#    echo -e "\ttransaction hash: $( echo ${JSON_OUTPUT} | jq '.txn_hash' )"

executeCommand 'estimatefee' 'POST' '/neutrino/estimatefee' "{ \"AddrToAmount\": { \"${TARGET_WALLET}\": 100000 } }"
showCommandResult 'fee sat' '.feeSat'

echo "----------"
echo

#
#   test commands of "Wallet" group
#

echo -e ">>> Group ${CYAN}[wallet]${NOCOLOR} APIs for management of on-chain (non-Lightning) payments, seed export and recovery, and on-chain transaction detection"
echo

executeCommand 'walletbalance' 'GET' '/wallet/balance'
echo -e "\ttotal balance: $( echo ${JSON_OUTPUT} | jq '.totalBalance' )"

#   we don't want to change the wallet's passphrase here, since this test is being made by REST_createWalletTest.sh script
#executeCommand 'changePassphrase' 'POST' '/wallet/changepassphrase' "{ \"current_passphrase\": \"${PASSPHRASE}\", \"new_passphrase\": \"${NEW_PASSPHRASE}\" }"
#showCommandResult 'result' ''

WALLET_PASSPHRASE='w4ll3tP@sswd'
executeCommand 'checkPassphrase' 'POST' '/wallet/checkpassphrase' "{ \"wallet_passphrase\": \"${WALLET_PASSPHRASE}\" }"
showCommandResult 'result' '.validPassphrase'

#   this test is meant to fail, since wallet is supposed to have been created already
WALLET_SEED='[ "plastic",  "hollow",  "mansion",  "keep",  "into",  "cloth",  "awesome",  "salmon",  "reopen",  "inner",  "replace",  "dice",  "life",  "example",  "around" ]'
SEED_PASSPHRASE='cGFzc3dvcmQ='
executeCommand 'create_wallet' 'POST' '/wallet/create' "{ \"wallet_passphrase\": \"${PASSPHRASE}\", \"wallet_seed\": ${WALLET_SEED}, \"seed_passphrase_bin\": \"${SEED_PASSPHRASE}\" }"
showCommandResult 'result' ''

executeCommand 'getsecret' 'POST' '/wallet/getsecret' '{ "name": "Isaac Assimov" }'
showCommandResult 'result' '.secret'

executeCommand 'getwalletseed' 'GET' '/wallet/seed'
showCommandResult 'result' '.seed'

#   we don't want to unlock the wallet here, since this test is being made by REST_createWalletTest.sh script
#executeCommand 'unlock_wallet' 'POST' '/wallet/unlock' "{ \"wallet_passphrase\": \"${WALLET_PASSPHRASE}\" }"
#showCommandResult 'result' ''

echo "----------"
echo

showCommandResult 'result' ''
echo "++++++++++++++++++++++++++++++++"
exit 0



#
#   test commands to manage payments
#

executeCommand 'sendmany' 'POST' '/api/v1/transaction/sendmany' "{ \"AddrToAmount\": { \"${TARGET_WALLET}\": 100000 } }"
TXID="$( getCommandResult '.txid' | tr -d '\"' )"
echo -e "    >>> transaction ID: ${TXID}"

executeCommand 'sendcoins' 'POST' '/api/v1/transaction/sendcoins' "{ \"addr\": \"${TARGET_WALLET}\", \"amount\": 10000000 }"
showCommandResult 'transaction ID' '.txid'

executeCommand 'listunspent' 'POST' '/api/v1/transaction/listunspent' '{ "minConfs": 1, "maxConfs": 100 }'
showCommandResult '#utxos' '.utxos | length'

executeCommand 'listchaintrns' 'POST' '/api/v1/transaction' '{ "startHeight": 1000000, "endHeight": 1347381 }'
showCommandResult '#transactions' '.transactions | length'

executeCommand 'setnetworkstewardvote' 'POST' '/api/v1/transaction/setnetworkstewardvote' '{ "voteAgainst": "0", "voteFor": "1" }'
showCommandResult 'result' ''

executeCommand 'getnetworkstewardvote' 'GET' '/api/v1/transaction/getnetworkstewardvote'
showCommandResult 'vote against' '.voteAgainst'
showCommandResult 'vote for' '.voteFor'

echo "----------"
echo

showCommandResult 'result' ''
echo "++++++++++++++++++++++++++++++++"
exit 0

################################################################################
#   test commands to manage the wallet
################################################################################
executeCommand 'newaddress' 'POST' '/api/v1/lightning/getnewaddress' '{  }'
showCommandResult 'result' ''

executeCommand 'getaddressbalances' 'POST' '/api/v1/wallet/addresses/balances' '{  }'
echo -e "\t#addresses: $( echo ${JSON_OUTPUT} | jq '.addrs | length' )"

#executeCommand 'signmessage' 'pkt1q0tgwuwcg4tmwegmevdfz3g6tw838upqcq8xt8u message'
executeCommand 'signmessage' 'POST' '/api/v1/lightning/signmessage' '{ "msg": "testing pld REST endpoints" }'
showCommandResult 'result' ''
#    echo -e "\tsignature: $( echo ${JSON_OUTPUT} | jq '.signature' )"

executeCommand 'resync' 'POST' '/api/v1/lightning/resync' '{  }'
showCommandResult 'result' ''

executeCommand 'stopresync' 'GET' '/api/v1/lightning/stopresync' ''
showCommandResult 'result' ''
#    echo -e "\tstop sync: $( echo ${JSON_OUTPUT} | jq '.value' )"

executeCommand 'importprivkey' 'POST' '/api/v1/lightning/importprivkey' '{ "privateKey": "cVgcgWwQpwzViWmG7dGyvf545ra6AdT4tV29UtQfE8okvPuznFZi", "rescan": true }'
showCommandResult 'result' ''
#    echo -e "\taddress: $( echo ${JSON_OUTPUT} | jq '.address' )"

executeCommand 'listlockunspent' 'GET' '/api/v1/lightning/listlockunspent'
showCommandResult 'result' ''
#    echo -e "\t#lock unspent: $( echo ${JSON_OUTPUT} | jq '.locked_unspent | length' )"

executeCommand 'lockunspent' 'POST' '/api/v1/lightning/lockunspent' '{ "lockname": "secure vault", "unlock": false, "transactions": [ { "txid": "934095dc4afa8d4b5d43732a96e78e11c0e88defdaab12d946f525e54478938f" } ] }' ''
showCommandResult 'result' ''

executeCommand 'createtransaction' 'POST' '/api/v1/lightning/createtransaction' '{ "toAddress": "pkt1q85n69mzthdxlwutn6dr6f7kwyd9nv8ulasdaqz", "amount": 100000 }'
showCommandResult 'result' ''
#    echo -e "\ttransaction: $( echo ${JSON_OUTPUT} | jq '.transaction' )"

#executeCommand 'dumpprivkey' 'pkt1q0tgwuwcg4tmwegmevdfz3g6tw838upqcq8xt8u'
executeCommand 'dumpprivkey' 'POST' '/api/v1/lightning/dumpprivkey' '{ "address": "pkt1q85n69mzthdxlwutn6dr6f7kwyd9nv8ulasdaqz" }'
showCommandResult 'result' ''
#    echo -e "\tprivate key: $( echo ${JSON_OUTPUT} | jq '.private_key' )"

#executeCommand 'getnewaddress' 'POST' '/api/v1/lightning/getnewaddress' '{  }'
#showCommandResult 'result' ''
#    echo -e "\taddress: $( echo ${JSON_OUTPUT} | jq '.address' )"

#executeCommand 'gettransaction' '934095dc4afa8d4b5d43732a96e78e11c0e88defdaab12d946f525e54478938f'
executeCommand 'gettransaction' 'POST' '/api/v1/lightning/gettransaction' ''
showCommandResult 'result' ''
#    echo -e "\tamount: $( echo ${JSON_OUTPUT} | jq '.transaction.amount' )"
#    echo -e "\tfee: $( echo ${JSON_OUTPUT} | jq '.transaction.fee' )"

executeCommand 'gettransactions' 'POST' '/api/v1/lightning/gettransactions' ''
showCommandResult 'result' ''

executeCommand 'sendfrom' 'POST' '/api/v1/lightning/sendfrom' '{  }'
showCommandResult 'result' ''

################################################################################
#   test commands to manage watch tower
################################################################################

#executeCommand 'wtclient' 'towers'

#   test commands to stop pld daemon
executeCommand 'stop' 'GET' '/api/v1/meta/stop'

rm -rf ${REST_ERRORS_FILE}
