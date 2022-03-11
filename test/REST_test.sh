#!  /usr/bin/bash

################################################################################
#   smoke tests for pld REST endpoints
################################################################################

export  PLD_REST_SERVER='http://localhost:8080'
export  REST_ERRORS_FILE='./rest.err'
export  JSON_OUTPUT=''
export  VERBOSE='false'

#   use curl to execute a command
executeCommand() {
    local COMMAND="${1}"
    local HTTP_METHOD="${2}"
    local URI="${3}"
    local PAYLOAD="${4}"

    RED='\033[0;31m'
    GREEN='\033[0;32m'
    LIGHTGRAY='\033[0;37m'
    NOCOLOR='\033[0m'

    if [ "${HTTP_METHOD}" == "GET" ]
    then
        if [ "${VERBOSE}" == 'true' ]
        then
            echo -e "[trace] ${LIGHTGRAY}curl \"${PLD_REST_SERVER}${URI}\"${NOCOLOR}"
        fi

        JSON_OUTPUT=$( curl "${PLD_REST_SERVER}${URI}" 2>> ${REST_ERRORS_FILE} )
    elif [ "${HTTP_METHOD}" == "POST" ]
    then
        if [ "${VERBOSE}" == 'true' ]
        then
            echo -e "[trace] ${LIGHTGRAY}curl -H \"Content-Type: application/json\" -X POST -d '${PAYLOAD}' \"${PLD_REST_SERVER}${URI}\"${NOCOLOR}"
        fi

        JSON_OUTPUT=$( curl -H "Content-Type: application/json" -X POST -d "${PAYLOAD}" "${PLD_REST_SERVER}${URI}" 2>> ${REST_ERRORS_FILE} )
    else
        echo -e "${RED}error: invalid HTTP method \"${HTTP_METHOD}\"${NOCOLOR}"
        return 1
    fi

    if [ $? -eq 0 ]
    then
        echo -e ">>> ${COMMAND}: ${GREEN}command successfully executed${NOCOLOR}"
    else
        echo -e "${RED}error: fail attempting to run command \"${COMMAND} ${ARGUMENTS}\": $?${NOCOLOR}"
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
        echo -e "    >>> ${TITLE}: ${RESULT}"
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

#
#   test commands to get info about the running pld daemon
#

executeCommand 'getinfo' 'GET' '/api/v1/meta/getinfo'
showCommandResult '#neutrino peers' '.neutrino.peers | length'

executeCommand 'getrecoveryinfo' 'GET' '/api/v1/meta/getrecoveryinfo'
showCommandResult 'recovery mode' '.recoveryMode'

executeCommand 'debuglevel' 'POST' '/api/v1/meta/debuglevel' '{ "show": true, "level_spec": "debug" }'
showCommandResult 'subsystems' '.subSystems'

executeCommand 'version' 'GET' '/api/v1/meta/version'
showCommandResult 'result' ''

#
#   fetch a public key for the channels tests
#

executeCommand 'describegraph' 'POST' '/api/v1/lightning/graph' '{ "includeUnannounced": true }'
PUBLIC_KEY="$( getCommandResult '.nodes | .[0] | .pubKey ' | tr -d '\"' | cut --characters=3- )"
echo -e "    >>> using public key: ${PUBLIC_KEY}"

#
#   test commands to manage channels
#

AMOUNT="100000"

executeCommand 'openchannel' 'POST' '/api/v1/channels/open' "{ \"node_pubkey\": \"${PUBLIC_KEY}\", \"local_funding_amount\": ${AMOUNT} }"
showCommandResult 'result' ''

CHANNEL_POINT="XPTO"

executeCommand 'closechannel' 'POST' '/api/v1/channels/close' "{ \"channel_point\": \"${CHANNEL_POINT}\" }"
showCommandResult 'result' ''

#executeCommand 'closeallchannels'
#showCommandResult 'result' ''

FUNDING_TXID="934095dc4afa8d4b5d43732a96e78e11c0e88defdaab12d946f525e54478938f"

executeCommand 'abandonchannel' 'POST' '/api/v1/channels/abandon' "{ \"channelPoint\": { \"funding_txid_str\": \"${FUNDING_TXID}\" } }"
showCommandResult 'result' ''

executeCommand 'channelbalance' 'GET' '/api/v1/channels/balance'
showCommandResult 'channel balance' '.balance'

executeCommand 'pendingchannels' 'GET' '/api/v1/channels/pending'
showCommandResult '#pending open channels' '.pendingOpenChannels | length'
showCommandResult '#pending closing channels' '.pendingClosingChannels | length'
showCommandResult 'limbo balance' '.totalLimboBalance'

executeCommand 'listchannels' 'POST' '/api/v1/channels' '{  }'
showCommandResult '#open channels' '.channels | length'

executeCommand 'closedchannels' 'POST' '/api/v1/channels/closed' '{  }'
showCommandResult '#closed channels' '.channels | length'

executeCommand 'getnetworkinfo' 'GET' '/api/v1/channels/networkinfo'
showCommandResult 'nodes' '.numNodes'
showCommandResult 'channels' '.numChannels'

executeCommand 'feereport' 'GET' '/api/v1/channels/feereport'
showCommandResult '#channel fees' '.channelFees | length'
showCommandResult 'week fee sum' '.weekFeeSum'

executeCommand 'updatechanpolicy' 'POST' '/api/v1/channels/policy' '{ "baseFeeMsat": 10, "feeRate": 10, "timeLockDelta": 20, "maxHtlcMsat": 30, "minHtlcMsat": 1, "minHtlcMsatSpecified": false }'
showCommandResult 'result' ''

executeCommand 'exportchanbackup' 'POST' '/api/v1/channels/backup/export' "{ \"chanPoint\": { \"funding_txid_str\": \"${FUNDING_TXID}\" } }"
showCommandResult 'result' '.multi_chan_backup.multi_chan_backup'

executeCommand 'verifychanbackup' 'POST' '/api/v1/channels/backup/verify' "{ \"singleChanBackups\": { \"chanBackups\": [ { \"chanPoint\": { \"fundingTxidStr\": \"${FUNDING_TXID}\", \"outputIndex\": 1000 }, \"chanBackup\": \"RW5jcnlwdGVkIENoYW4gQmFja3Vw\" } ] } }"
showCommandResult 'result' ''

executeCommand 'restorechanbackup' 'POST' '/api/v1/channels/backup/restore' "{ \"chanBackups\": [ { \"chanPoint\": { \"fundingTxidStr\": \"${FUNDING_TXID}\", \"outputIndex\": 1000 }, \"chanBackup\": \"RW5jcnlwdGVkIENoYW4gQmFja3Vw\" } ] }"
showCommandResult 'result' ''

showCommandResult 'result' ''
echo "++++++++++++++++++++++++++++++++"
exit 0

#
#   test commands to get graph info
#
executeCommand 'describegraph' 'POST' '/api/v1/lightning/graph' '{ "includeUnannounced": true }'
showCommandResult 'last update' '.nodes | .[0] | .lastUpdate '
PUBLIC_KEY="$( getCommandResult '.nodes | .[0] | .pubKey ' | tr -d '\"' )"
echo -e "    >>> public key: ${PUBLIC_KEY}"

executeCommand 'getnodemetrics' 'POST' '/api/v1/graph/nodemetrics' '{ "types": [ 0, 1 ] }'
showCommandResult 'betweenness centrality' '.betweennessCentrality'

CHAN_ID=123
executeCommand 'getchaninfo' 'POST' '/api/v1/graph/channel' "{ \"chanId\": ${CHAN_ID} }"
showCommandResult 'result' ''

executeCommand 'getnodeinfo' 'POST' '/api/v1/graph/nodeinfo' "{ \"pubKey\": \"${PUBLIC_KEY}\", \"includeChannels\": true }"
showCommandResult 'last update' '.node.lastUpdate'
showCommandResult '#channels' '.node.channels | length'

#
#   test commands to manage invoices
#

executeCommand 'addinvoice' 'POST' '/api/v1/invoice/add' '{ "memo": "xpto", "value": 10, "expiry": 3600 }'
RHASH="$( getCommandResult '.rHash' | tr -d '\"' )"
PAYREQ="$( getCommandResult '.paymentRequest' | tr -d '\"' )"
echo -e "    >>> rHash: ${RHASH}"
echo -e "    >>> payment request: ${PAYREQ}"

executeCommand 'lookupinvoice' 'POST' '/api/v1/invoice/lookup' "{ \"rHash\": \"${RHASH}\" }"
showCommandResult 'last update' '.lastUpdate'
showCommandResult 'index' '.addIndex'
showCommandResult 'state' '.state'

executeCommand 'listinvoices' 'POST' '/api/v1/invoice' '{ "indexOffset": 1, "numMaxInvoices": 10 }'
showCommandResult '#invoices' '.invoices | length'

executeCommand 'decodepayreq' 'POST' '/api/v1/invoice/decodepayreq' "{ \"payReq\": \"${PAYREQ}\" }"
showCommandResult 'destination' '.destination'
showCommandResult 'payment Hash' '.paymentHash'
showCommandResult '#satoshis' '.numSatoshis'

#
#   test commands to manage on-chain transactions
#

TARGET_WALLET="pkt1q07ly7r47ss4drsvt2zq9zkcstksrq2dap3x0yw"

executeCommand 'estimatefee' 'POST' '/api/v1/transaction/estimatefee' "{ \"AddrToAmount\": { \"${TARGET_WALLET}\": 100000 } }"
showCommandResult 'fee sat' '.feeSat'

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

executeCommand 'bcasttransaction' 'POST' '/api/v1/transaction/bcast' "{ \"tx\": \"${TXID}\" }"
showCommandResult 'result' ''
#    echo -e "\ttransaction hash: $( echo ${JSON_OUTPUT} | jq '.txn_hash' )"

#
#   test commands to manage payments
#

executeCommand 'sendpayment' 'POST' '/api/v1/payment/send' "{ \"paymentHash\": \"${RHASH}\", \"amt\": 100000, \"dest\": \"${TARGET_WALLET}\" }"
showCommandResult 'result' ''

#executeCommand 'payinvoice' 'lnpkt100u1p3q4r85pp5kecz6ckl97wwe2nnqn6lq5lju30z9sc8uaeacamudxv52kykgdnqdqqcqzpgsp5fa0tpf3j3ecppn3tvmc50n6w7pl6dcs7zvus82splfjs2qevwkxq9qy9qsq4sfdxwzrku87zaphgh6wa3rtc2a8g7rmg6a2dp4myk3qa8c7409sv205xxfsc2n0mzmemcg92ukg7x6q7xlkp5ca9gdwvsqmtpuazccpw25hg9'
#executeCommand 'payinvoice' 'POST' '/api/v1/payment/payinvoice' "{ \"paymentHash\": \"${RHASH}\", \"amt\": 100000, \"dest\": \"${TARGET_WALLET}\" }"
#showCommandResult 'result' ''

executeCommand 'sendtoroute' 'POST' '/api/v1/payment/sendroute' '{ "paymentHash": "02e28f38ad50869fd3f3d75147d69bc637090aa9b5013ee49a65c0dda2bf0ab51e", "route": { "hops": { "chanId": "xpto"} } }'
showCommandResult 'result' ''

executeCommand 'listpayments' 'POST' '/api/v1/payment' '{ "indexOffset": 1, "maxPayments": 10, "includeIncomplete": true }'
showCommandResult '#payments' '.payments | length'

PUBKEY="02e28f38ad50869fd3f3d75147d69bc637090aa9b5013ee49a65c0dda2bf0ab51e"
AMOUNT="100000"
#executeCommand 'queryroutes' 'POST' '/api/v1/payment/queryroutes' '02e28f38ad50869fd3f3d75147d69bc637090aa9b5013ee49a65c0dda2bf0ab51e 1'
executeCommand 'queryroutes' 'POST' '/api/v1/payment/queryroutes' "{  }"
showCommandResult 'result' ''

executeCommand 'fwdinghistory' 'POST' '/api/v1/payment/fwdinghistory' '{ "indexOffset": 0, "numMaxEvents": 25 }'
showCommandResult '#forwarding events' '.forwardingEvents | length'

executeCommand 'querymc' 'GET' '/api/v1/payment/querymc'
showCommandResult 'result' ''
#    echo -e "\t#pairs: $( echo ${JSON_OUTPUT} | jq '.pairs | length' )"

FROM_NODE="01020304"
TO_NODE="02030405"
AMOUNT="100000"
executeCommand 'queryprob' 'POST' '/api/v1/payment/queryprob' "{ \"fromNode\": \"${FROM_NODE}\", \"toNode\": \"${TO_NODE}\", \"amtMsat\": \"${AMOUNT}\" }"
showCommandResult 'result' ''

executeCommand 'resetmc' 'GET' '/api/v1/payment/resetmc'
showCommandResult 'result' ''

executeCommand 'buildroute' 'POST' '/api/v1/payment/buildroute' '{ "amtMsat": 0, "hopPubkeys": [ "01020304", "02030405", "03040506" ] }'
showCommandResult 'result' ''

showCommandResult 'result' ''
echo "++++++++++++++++++++++++++++++++"
exit 0

################################################################################
#   test commands to manage peers
################################################################################
executeCommand 'connect' 'POST' '/api/v1/peer/connect' '{ "addr": { "pubkey": "272648127365482", "host": "192.168.40.1:8080" } }'
showCommandResult 'result' ''

executeCommand 'disconnect' 'POST' '/api/v1/peer/disconnect' '{  }'
showCommandResult 'result' ''

executeCommand 'listpeers' 'GET' '/api/v1/peer'
showCommandResult 'result' ''
showCommandResult '#peers' '.peers | length'

################################################################################
#   test commands to manage the wallet
################################################################################
executeCommand 'newaddress' 'POST' '/api/v1/lightning/getnewaddress' '{  }'
showCommandResult 'result' ''

executeCommand 'walletbalance' 'GET' '/api/v1/lightning/walletbalance'
echo -e "\ttotal balance: $( echo ${JSON_OUTPUT} | jq '.totalBalance' )"

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

executeCommand 'genseed' 'POST' '/api/v1/lightning/genseed' '{ "aezeedPassphrase": "cGFzc3dvcmQ=" }'
showCommandResult 'result message' '.message'
showCommandResult 'enciphered seed' 'encipheredSeed'

executeCommand 'getwalletseed' 'POST' '/api/v1/lightning/getwalletseed' '{  }'
showCommandResult 'result' ''
#    echo -e "\twallet seed: $( echo ${JSON_OUTPUT} | jq '.seed' )"

executeCommand 'getsecret' 'POST' '/api/v1/lightning/getsecret' '{ "name": "Isaac Assimov" }'
showCommandResult 'result' ''
#    echo -e "\tsecret: $( echo ${JSON_OUTPUT} | jq '.secret' )"

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
