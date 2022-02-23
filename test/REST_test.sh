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

    if [ "${HTTP_METHOD}" == "GET" ]
    then
        JSON_OUTPUT=$( curl "${PLD_REST_SERVER}${URI}" 2>> ${REST_ERRORS_FILE} )
    elif [ "${HTTP_METHOD}" == "POST" ]
    then
        JSON_OUTPUT=$( curl -H "Content-Type: application/json" -X POST -d "${PAYLOAD}" "${PLD_REST_SERVER}${URI}" 2>> ${REST_ERRORS_FILE} )
    else
        echo "error: invalid HTTP method \"${HTTP_METHOD}\""
        return 1
    fi

    if [ $? -eq 0 ]
    then
        echo ">>> ${COMMAND} ${ARGUMENTS}: command successfully executed"
    else
        echo "error: fail attempting to run command \"${COMMAND} ${ARGUMENTS}\": $?"
        return 1
    fi
}

#   splash screen
echo ">>>>> Testing pld REST endpoints"
echo

#   check if curl is available
output=$( which curl 2> /dev/null )
if [ $? -ne 0 ]
then
    exit "error: 'curl' is required to run this script"
fi

#   check if jq is available
output=$( which jq 2> /dev/null )
if [ $? -ne 0 ]
then
    exit "error: 'jq' is required to run this script"
fi

#   test commands to get info about the running pld daemon
executeCommand 'getinfo' 'GET' '/api/v1/meta/getinfo'
if [ $? -eq 0 ]
then
    echo -e "\t#neutrino peers: $( echo ${JSON_OUTPUT} | jq '.neutrino.peers | length' )"
fi

executeCommand 'getrecoveryinfo' 'GET' '/api/v1/meta/getrecoveryinfo'
if [ $? -eq 0 ]
then
    echo -e "\trecovery mode: $( echo ${JSON_OUTPUT} | jq '.recoveryMode' )"
fi

executeCommand 'debuglevel' 'POST' '/api/v1/debuglevel' '{ "show": true, "level_spec": "debug" }'

executeCommand 'version' 'GET' '/api/v2/versioner/version'
if [ $? -eq 0 ]
then
    echo -e "${JSON_OUTPUT}"
fi

#   test commands to manage channels
#executeCommand 'openchannel'
#executeCommand 'closechannel'
#executeCommand 'closeallchannels'

FUNDING_TXID="12345678900"
OUTPUT_INDEX="123"

executeCommand 'abandonchannel' 'POST' "/api/v1/channels/abandon/{$FUNDING_TXID}/{$OUTPUT_INDEX}" '{  }'
if [ $? -eq 0 ]
then
    echo -e "${JSON_OUTPUT}"
fi

executeCommand 'channelbalance' 'GET' '/api/v1/channelbalance'
if [ $? -eq 0 ]
then
    echo -e "\tchannel balance: $( echo ${JSON_OUTPUT} | jq '.balance' )"
fi

executeCommand 'pendingchannels' 'GET' '/api/v1/pendingchannels'
if [ $? -eq 0 ]
then
    echo -e "\tlimbo balance: $( echo ${JSON_OUTPUT} | jq '.totalLimboBalance' )"
fi

executeCommand 'listchannels' 'POST' '/api/v1/channels' '{  }'
if [ $? -eq 0 ]
then
    echo -e "\t#open channels: $( echo ${JSON_OUTPUT} | jq '.channels | length' )"
fi

executeCommand 'closedchannels' 'POST' '/api/v1/channels/closed' '{  }'
if [ $? -eq 0 ]
then
    echo -e "\t#closed channels: $( echo ${JSON_OUTPUT} | jq '.channels | length' )"
fi

executeCommand 'getnetworkinfo' 'GET' '/api/v1/graph/info'
if [ $? -eq 0 ]
then
    echo -e "\t#nodes: $( echo ${JSON_OUTPUT} | jq '.numNodes' )"
    echo -e "\t#channels: $( echo ${JSON_OUTPUT} | jq '.numChannels' )"
fi

executeCommand 'feereport' 'GET' '/api/v1/fees'
if [ $? -eq 0 ]
then
    echo -e "\tweek fee sum: $( echo ${JSON_OUTPUT} | jq '.weekFeeSum' )"
fi

executeCommand 'updatechanpolicy' 'POST' '/api/v1/chanpolicy' '{ "baseFeeMsat": 10, "feeRate": 10, "timeLockDelta": 20, "maxHtlcMsat": 30, "minHtlcMsat": 1, "minHtlcMsatSpecified": false }'
if [ $? -eq 0 ]
then
    echo -e "${JSON_OUTPUT}"
fi

executeCommand 'exportchanbackup' 'POST' "/api/v1/channels/backup/{$FUNDING_TXID}/{$OUTPUT_INDEX}" "{ \"chanPoint\": { \"outputIndex\": \"{$OUTPUT_INDEX}\" } }"
if [ $? -eq 0 ]
then
    echo -e "${JSON_OUTPUT}"
    MULTI_BACKUP=$( echo ${JSON_OUTPUT} | jq '.multi_chan_backup.multi_chan_backup' | tr --delete '"' )
    echo -e "\tmulti backup: ${MULTI_BACKUP}"
fi

executeCommand 'verifychanbackup' 'POST' '/api/v1/channels/backup/verify' "{ \"multiChanBackup\": {  } }"
if [ $? -eq 0 ]
then
    echo -e "${JSON_OUTPUT}"
fi

executeCommand 'restorechanbackup' 'POST' '/api/v1/channels/backup/restore' '{ "backup": true }'
if [ $? -eq 0 ]
then
    echo -e "${JSON_OUTPUT}"
fi

#   test commands to stop pld daemon
executeCommand 'stop' 'GET' '/api/v1/stop'

rm -rf ${REST_ERRORS_FILE}
