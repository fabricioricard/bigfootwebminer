#!  /usr/bin/bash

################################################################################
#   smoke tests for pld REST endpoints
################################################################################

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
        JSON_OUTPUT=$( curl ${URI} 2>> ${REST_ERRORS_FILE} )
    elif [ "${HTTP_METHOD}" == "POST" ]
    then
        JSON_OUTPUT=$( curl -H "Content-Type: application/json" -X POST -d "${PAYLOAD}" ${URI} 2>> ${REST_ERRORS_FILE} )
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
executeCommand 'getinfo' 'GET' 'http://localhost:8080/api/v1/meta/getinfo'
if [ $? -eq 0 ]
then
    echo -e "${JSON_OUTPUT}"
    echo -e "\t#neutrino peers: $( echo ${JSON_OUTPUT} | jq '.neutrino.peers | length' )"
fi

executeCommand 'getrecoveryinfo' 'GET' 'http://localhost:8080/api/v1/meta/getrecoveryinfo'
if [ $? -eq 0 ]
then
    echo -e "${JSON_OUTPUT}"
    echo -e "\trecovery mode: $( echo ${JSON_OUTPUT} | jq '.recovery_mode' )"
fi

executeCommand 'debuglevel' 'POST' 'http://localhost:8080/api/v1/debuglevel' '{ "show": true, "level_spec": "debug" }'
if [ $? -eq 0 ]
then
    echo -e "${JSON_OUTPUT}"
fi

executeCommand 'stop' 'GET' 'http://localhost:8080/api/v1/stop'
if [ $? -eq 0 ]
then
    echo -e "${JSON_OUTPUT}"
fi

rm -rf ${REST_ERRORS_FILE}
