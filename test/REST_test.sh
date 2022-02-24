#!  /usr/bin/bash

################################################################################
#   smoke tests for pld REST endpoints
################################################################################

export  PLD_REST_SERVER='http://localhost:8080'
export  REST_ERRORS_FILE="./rest.err"
export  JSON_OUTPUT=""
export  TARGET_WALLET="pkt1q07ly7r47ss4drsvt2zq9zkcstksrq2dap3x0yw"

#   use curl to execute a command
executeCommand() {
    local COMMAND="${1}"
    local HTTP_METHOD="${2}"
    local URI="${3}"
    local PAYLOAD="${4}"

    GREEN='\033[0;32m'
    RED='\033[0;31m'
    NOCOLOR='\033[0m'

    if [ "${HTTP_METHOD}" == "GET" ]
    then
        JSON_OUTPUT=$( curl "${PLD_REST_SERVER}${URI}" 2>> ${REST_ERRORS_FILE} )
    elif [ "${HTTP_METHOD}" == "POST" ]
    then
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

executeCommand 'debuglevel' 'POST' '/api/v1/debuglevel' '{ "show": true, "level_spec": "debug" }'

executeCommand 'version' 'GET' '/api/v2/versioner/version'
showCommandResult 'result' ''

#   test commands to manage channels
#executeCommand 'openchannel'
#executeCommand 'closechannel'
#executeCommand 'closeallchannels'

FUNDING_TXID="12345678900"
OUTPUT_INDEX="123"

executeCommand 'abandonchannel' 'POST' "/api/v1/channels/abandon/{$FUNDING_TXID}/{$OUTPUT_INDEX}" '{  }'
showCommandResult 'result' ''

executeCommand 'channelbalance' 'GET' '/api/v1/channelbalance'
showCommandResult 'channel balance' '.balance'

executeCommand 'pendingchannels' 'GET' '/api/v1/pendingchannels'
showCommandResult 'limbo balance' '.totalLimboBalance'

executeCommand 'listchannels' 'POST' '/api/v1/channels' '{  }'
showCommandResult 'open channels' '.channels | length'

executeCommand 'closedchannels' 'POST' '/api/v1/channels/closed' '{  }'
showCommandResult 'closed channels' '.channels | length'

executeCommand 'getnetworkinfo' 'GET' '/api/v1/graph/info'
showCommandResult 'nodes' '.numNodes'
showCommandResult 'channels' '.numChannels'

executeCommand 'feereport' 'GET' '/api/v1/fees'
showCommandResult 'week fee sum' '.weekFeeSum'

executeCommand 'updatechanpolicy' 'POST' '/api/v1/chanpolicy' '{ "baseFeeMsat": 10, "feeRate": 10, "timeLockDelta": 20, "maxHtlcMsat": 30, "minHtlcMsat": 1, "minHtlcMsatSpecified": false }'
showCommandResult 'result' ''

executeCommand 'exportchanbackup' 'POST' "/api/v1/channels/backup/{$FUNDING_TXID}/{$OUTPUT_INDEX}" "{ \"chanPoint\": { \"outputIndex\": \"{$OUTPUT_INDEX}\" } }"
if [ $? -eq 0 ]
then
    echo -e "${JSON_OUTPUT}"
    MULTI_BACKUP=$( echo ${JSON_OUTPUT} | jq '.multi_chan_backup.multi_chan_backup' | tr --delete '"' )
    echo -e "\tmulti backup: ${MULTI_BACKUP}"
fi

executeCommand 'verifychanbackup' 'POST' '/api/v1/channels/backup/verify' "{ \"multiChanBackup\": {  } }"
showCommandResult 'result' ''

executeCommand 'restorechanbackup' 'POST' '/api/v1/channels/backup/restore' '{ "backup": true }'
showCommandResult 'result' ''

#   test commands to get graph info
executeCommand 'describegraph' 'POST' '/api/v1/graph' '{ "includeUnannounced": true }'
showCommandResult 'last update' '.nodes | .[0] | .lastUpdate '

executeCommand 'getnodemetrics' 'POST' '/api/v1/graph/nodemetrics' '{ "types": [ 0, 1 ] }'
showCommandResult 'betweenness centrality' '.betweennessCentrality'

CHAN_ID="123"
executeCommand 'getchaninfo' 'POST' "/api/v1/graph/edge/{$CHAN_ID}" "{ \"chanId\": {$CHAN_ID} }"
showCommandResult 'result' ''

PUBLIC_KEY="02c9d02352f3cfb06ad2d296d08098aaa2f0146e087c7cda0edc444a1b6c27905b"
executeCommand 'getnodeinfo' 'POST' '/api/v1/graph/node/{$PUBLIC_KEY}' "{ \"pubKey\": \"{$PUBLIC_KEY}\", \"includeChannels\": true }"
showCommandResult 'result' ''

#   test commands to stop pld daemon
executeCommand 'stop' 'GET' '/api/v1/stop'

rm -rf ${REST_ERRORS_FILE}
