#!  /usr/bin/bash

################################################################################
#   smoke tests for pld REST endpoints
################################################################################

export  PLD_REST_SERVER='http://localhost:8080'
export  REST_ERRORS_FILE='./rest.err'
export  JSON_OUTPUT=''
export  VERBOSE='true'

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
        echo -e "\t#${TITLE}: ${RESULT}"
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

#   test commands to get info about the running pld daemon
executeCommand 'getinfo' 'GET' '/api/v1/meta/getinfo'
showCommandResult 'neutrino peers' '.neutrino.peers | length'

executeCommand 'getrecoveryinfo' 'GET' '/api/v1/meta/getrecoveryinfo'
showCommandResult 'recovery mode' '.recoveryMode'

executeCommand 'debuglevel' 'POST' '/api/v1/meta/debuglevel' '{ "show": true, "level_spec": "debug" }'

executeCommand 'version' 'GET' '/api/v1/meta/version'
showCommandResult 'result' ''

#   test commands to manage channels
#executeCommand 'openchannel'
#showCommandResult 'result' ''

#executeCommand 'closechannel'
#showCommandResult 'result' ''

#executeCommand 'closeallchannels'
#showCommandResult 'result' ''

FUNDING_TXID="12345678900"
OUTPUT_INDEX="123"

executeCommand 'abandonchannel' 'POST' '/api/v1/channel/abandon' '{  }'
showCommandResult 'result' ''

executeCommand 'channelbalance' 'GET' '/api/v1/channel/balance'
showCommandResult 'channel balance' '.balance'

executeCommand 'pendingchannels' 'GET' '/api/v1/channel/pending'
showCommandResult 'limbo balance' '.totalLimboBalance'

executeCommand 'listchannels' 'POST' '/api/v1/channel' '{  }'
showCommandResult 'open channels' '.channels | length'

executeCommand 'closedchannels' 'POST' '/api/v1/channel/closed' '{  }'
showCommandResult 'closed channels' '.channels | length'

executeCommand 'getnetworkinfo' 'GET' '/api/v1/channel/networkinfo'
showCommandResult 'nodes' '.numNodes'
showCommandResult 'channels' '.numChannels'

executeCommand 'feereport' 'GET' '/api/v1/channel/feereport'
showCommandResult 'week fee sum' '.weekFeeSum'

executeCommand 'updatechanpolicy' 'POST' '/api/v1/channel/policy' '{ "baseFeeMsat": 10, "feeRate": 10, "timeLockDelta": 20, "maxHtlcMsat": 30, "minHtlcMsat": 1, "minHtlcMsatSpecified": false }'
showCommandResult 'result' ''

executeCommand 'exportchanbackup' 'POST' '/api/v1/channel/backup/export' "{ \"chanPoint\": { \"outputIndex\": \"${OUTPUT_INDEX}\" } }"
if [ $? -eq 0 ]
then
    echo -e "${JSON_OUTPUT}"
    MULTI_BACKUP=$( echo ${JSON_OUTPUT} | jq '.multi_chan_backup.multi_chan_backup' | tr --delete '"' )
    echo -e "\tmulti backup: ${MULTI_BACKUP}"
fi

executeCommand 'verifychanbackup' 'POST' '/api/v1/channel/backup/verify' "{ \"multiChanBackup\": {  } }"
showCommandResult 'result' ''

executeCommand 'restorechanbackup' 'POST' '/api/v1/channel/backup/restore' '{ "backup": true }'
showCommandResult 'result' ''

#   test commands to get graph info
executeCommand 'describegraph' 'POST' '/api/v1/graph' '{ "includeUnannounced": true }'
showCommandResult 'last update' '.nodes | .[0] | .lastUpdate '
PUBLIC_KEY="$( getCommandResult '.nodes | .[0] | .pubKey ' | tr -d '\"' )"

executeCommand 'getnodemetrics' 'POST' '/api/v1/graph/nodemetrics' '{ "types": [ 0, 1 ] }'
showCommandResult 'betweenness centrality' '.betweennessCentrality'

CHAN_ID="123"
executeCommand 'getchaninfo' 'POST' '/api/v1/graph/channel' "{ \"chanId\": ${CHAN_ID} }"
showCommandResult 'result' ''

executeCommand 'getnodeinfo' 'POST' '/api/v1/graph/nodeinfo' "{ \"pubKey\": \"${PUBLIC_KEY}\", \"includeChannels\": true }"
showCommandResult 'last update' '.lastUpdate'

#   test commands to manage invoices
executeCommand 'addinvoice' 'POST' '/api/v1/invoice/add' '{ "memo": "xpto", "value": 10, "expiry": 3600 }'
showCommandResult 'rHash' '.rHash'
showCommandResult 'payment request' '.paymentRequest'
RHASH="$( getCommandResult '.rHash' | tr -d '\"' )"
PAYREQ="$( getCommandResult '.paymentRequest' | tr -d '\"' )"

executeCommand 'lookupinvoice' 'POST' '/api/v1/invoice/lookup' "{ \"rHash\": \"${RHASH}\" }"
showCommandResult 'last update' '.lastUpdate'
showCommandResult 'index' '.addIndex'
showCommandResult 'state' '.state'

#executeCommand 'listinvoices' 'POST' '/api/v1/invoices' '{ "pendingOnly": false, "indexOffset": 1, "numMaxInvoices": 10, "reversed": false }'
executeCommand 'listinvoices' 'POST' '/api/v1/invoice' '{ "indexOffset": 1, "numMaxInvoices": 10 }'
showCommandResult 'result' ''

executeCommand 'decodepayreq' 'POST' '/api/v1/invoice/decodepayreq' "{ \"payReq\": \"${PAYREQ}\" }"
showCommandResult 'result' ''

#   test commands to manage on-chain transactions
export  TARGET_WALLET="pkt1q07ly7r47ss4drsvt2zq9zkcstksrq2dap3x0yw"

executeCommand 'estimatefee' 'POST' '/api/v1/transaction/estimatefee' "{ \"AddrToAmount\": [ { \"${TARGET_WALLET}\": 100000 } ] }"
showCommandResult 'result' ''
#    echo -e "\tfee sat: $( echo ${JSON_OUTPUT} | jq '.fee_sat' )"

executeCommand 'sendmany' 'POST' '/api/v1/transaction/sendmany' "{ \"AddrToAmount\": [ { \"${TARGET_WALLET}\": 100000 } ] }"
showCommandResult 'result' ''
#    echo -e "\ttransaction Id: $( echo ${JSON_OUTPUT} | jq '.txid' )"

executeCommand 'sendcoins' 'POST' '/api/v1/transaction/sendcoins' "{ \"addr\": \"${TARGET_WALLET}\", \"amount\": 10000000 }"
showCommandResult 'result' ''
#    echo -e "\ttransaction Id: $( echo ${JSON_OUTPUT} | jq '.txid' )"

executeCommand 'listunspent' 'POST' '/api/v1/transaction/listunspent' '{ "minConfs": 1, "maxConfs": 100 }'
showCommandResult 'result' ''
#    echo -e "\t#utxos: $( echo ${JSON_OUTPUT} | jq '.utxos | length' )"

executeCommand 'listchaintrns' 'POST' '/api/v1/transaction' '{ "startHeight": 1000000, "endHeight": 1300000 }'
showCommandResult 'result' ''
#    echo -e "\t#transactions: $( echo ${JSON_OUTPUT} | jq '.transactions | length' )"

executeCommand 'setnetworkstewardvote' 'POST' '/api/v1/transaction/setnetworkstewardvote' '{ "voteAgainst": "0", "voteFor": "1" }'
showCommandResult 'result' ''

executeCommand 'getnetworkstewardvote' 'GET' '/api/v1/transaction/getnetworkstewardvote'
showCommandResult 'result' ''
#    echo -e "\tvote against: $( echo ${JSON_OUTPUT} | jq '.vote_against' )"
#    echo -e "\tvote for: $( echo ${JSON_OUTPUT} | jq '.vote_for' )"

executeCommand 'bcasttransaction' 'POST' '/api/v1/transaction/bcast' '{ "tx": "01020304050607080910" }'
showCommandResult 'result' ''
#    echo -e "\ttransaction hash: $( echo ${JSON_OUTPUT} | jq '.txn_hash' )"

#   test commands to manage payments
executeCommand 'sendpayment' 'POST' '/api/v1/payment/send' '{ "paymentHash": "02e28f38ad50869fd3f3d75147d69bc637090aa9b5013ee49a65c0dda2bf0ab51e", "amt": 100000, "dest": "1cc616cdeb96016bf278bfea15d55541d31823986b33c0dab38024cb8eff3791" }'
showCommandResult 'result' ''

#executeCommand 'payinvoice' 'lnpkt100u1p3q4r85pp5kecz6ckl97wwe2nnqn6lq5lju30z9sc8uaeacamudxv52kykgdnqdqqcqzpgsp5fa0tpf3j3ecppn3tvmc50n6w7pl6dcs7zvus82splfjs2qevwkxq9qy9qsq4sfdxwzrku87zaphgh6wa3rtc2a8g7rmg6a2dp4myk3qa8c7409sv205xxfsc2n0mzmemcg92ukg7x6q7xlkp5ca9gdwvsqmtpuazccpw25hg9'

executeCommand 'sendtoroute' 'POST' '/api/v1/payment/sendroute' '{ "paymentHash": "02e28f38ad50869fd3f3d75147d69bc637090aa9b5013ee49a65c0dda2bf0ab51e", "route": { "hops": [ { "chanId": "xpto"} ] } }'
showCommandResult 'result' ''

executeCommand 'listpayments' 'GET' '/api/v1/payment' ''
showCommandResult 'result' ''
#    echo -e "\t#payments: $( echo ${JSON_OUTPUT} | jq '.payments | length' )"

PUBKEY="02e28f38ad50869fd3f3d75147d69bc637090aa9b5013ee49a65c0dda2bf0ab51e"
AMOUNT="100000"
executeCommand 'queryroutes' 'POST' '/api/v1/payment/queryroutes' '02e28f38ad50869fd3f3d75147d69bc637090aa9b5013ee49a65c0dda2bf0ab51e 1'
showCommandResult 'result' ''

executeCommand 'fwdinghistory' 'POST' '/api/v1/payment/fwdinghistory' '{ "indexOffset": 0, "numMaxEvents": 25 }'
showCommandResult 'result' ''
#    echo -e "\t#forwarding events: $( echo ${JSON_OUTPUT} | jq '.forwarding_events | length' )"

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

#   test commands to manage peers
executeCommand 'connect' 'POST' '/api/v1/peer/connect' '{ "addr": { "pubkey": "272648127365482", "host": "192.168.40.1:8080" } }'
showCommandResult 'result' ''

executeCommand 'disconnect' 'POST' '/api/v1/peer/disconnect' '{  }'
showCommandResult 'result' ''

executeCommand 'listpeers' 'GET' '/api/v1/peer'
showCommandResult 'result' ''
showCommandResult '#peers' '.peers | length'

#   test commands to manage the wallet
executeCommand 'newaddress' 'POST' '/api/v1/lightning/getnewaddress' '{  }'
showCommandResult 'result' ''

executeCommand 'walletbalance' 'POST' '/api/v1/lightning/walletbalance' '{  }'
showCommandResult 'result' ''
#    echo -e "\ttotal balance: $( echo ${JSON_OUTPUT} | jq '.total_balance' )"

executeCommand 'getaddressbalances' 'POST' '/api/v1/lightning/getaddressbalances' '{  }'
showCommandResult 'result' ''
#    echo -e "\t#addresses: $( echo ${JSON_OUTPUT} | jq '.addrs | length' )"

#executeCommand 'signmessage' 'pkt1q0tgwuwcg4tmwegmevdfz3g6tw838upqcq8xt8u message'
executeCommand 'signmessage' 'POST' '/api/v1/lightning/signmessage' '{  }'
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

executeCommand 'getsecret' 'POST' '/api/v1/lightning/getsecret' '{  }'
showCommandResult 'result' ''
#    echo -e "\tsecret: $( echo ${JSON_OUTPUT} | jq '.secret' )"

#executeCommand 'importprivkey' 'cVgcgWwQpwzViWmG7dGyvf545ra6AdT4tV29UtQfE8okvPuznFZi'
executeCommand 'importprivkey' 'POST' '/api/v1/lightning/importprivkey' '{  }'
showCommandResult 'result' ''
#    echo -e "\taddress: $( echo ${JSON_OUTPUT} | jq '.address' )"

executeCommand 'listlockunspent' 'POST' '/api/v1/lightning/listlockunspent' '{  }'
showCommandResult 'result' ''
#    echo -e "\t#lock unspent: $( echo ${JSON_OUTPUT} | jq '.locked_unspent | length' )"

executeCommand 'lockunspent' 'POST' '/api/v1/lightning/lockunspent' '{  }' ''
showCommandResult 'result' ''

#executeCommand 'createtransaction' 'pkt1q85n69mzthdxlwutn6dr6f7kwyd9nv8ulasdaqz'
executeCommand 'createtransaction' 'POST' '/api/v1/lightning/createtransaction' '{  }'
showCommandResult 'result' ''
#    echo -e "\ttransaction: $( echo ${JSON_OUTPUT} | jq '.transaction' )"

#executeCommand 'dumpprivkey' 'pkt1q0tgwuwcg4tmwegmevdfz3g6tw838upqcq8xt8u'
executeCommand 'dumpprivkey' 'POST' '/api/v1/lightning/dumpprivkey' '{  }'
showCommandResult 'result' ''
#    echo -e "\tprivate key: $( echo ${JSON_OUTPUT} | jq '.private_key' )"

executeCommand 'getnewaddress' 'POST' '/api/v1/lightning/getnewaddress' '{  }'
showCommandResult 'result' ''
#    echo -e "\taddress: $( echo ${JSON_OUTPUT} | jq '.address' )"

#executeCommand 'gettransaction' '934095dc4afa8d4b5d43732a96e78e11c0e88defdaab12d946f525e54478938f'
executeCommand 'gettransaction' 'POST' '/api/v1/lightning/gettransactions' ''
showCommandResult 'result' ''
#    echo -e "\tamount: $( echo ${JSON_OUTPUT} | jq '.transaction.amount' )"
#    echo -e "\tfee: $( echo ${JSON_OUTPUT} | jq '.transaction.fee' )"

executeCommand 'sendfrom' 'POST' '/api/v1/lightning/sendfrom' '{  }'
showCommandResult 'result' ''

#   test commands to stop pld daemon
executeCommand 'stop' 'GET' '/api/v1/meta/stop'

rm -rf ${REST_ERRORS_FILE}
