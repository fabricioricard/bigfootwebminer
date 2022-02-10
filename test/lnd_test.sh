#!  /usr/bin/bash

################################################################################
#   requires the installation of jq
################################################################################

export  PKT_HOME="$( pwd )"
export  PLD="${PKT_HOME}/bin/pld"
export  PLD_OPTIONS=""
export  PLD_OUTPUT_FILE="./pld.out"
export  PLD_PID=
export  PLDCTL="${PKT_HOME}/bin/pldctl"
export  PLDCTL_OPTIONS=""
export  PLDCTL_ERRORS_FILE="./pldctl.err"

#   start pld deamon in background
startPldDeamon() {

    ${PLD} ${PLD_OPTIONS} > ${PLD_OUTPUT_FILE} &

    PLD_PID=$!

    echo ">>> ${PLD} daemon up and running: PID: ${PLD_PID}"

    sleep 5s
}

#   stop pld deamon
stopPldDeamon() {

    executeCommand 'stop'

    kill ${PLD_PID} 2> /dev/null

    sleep 10s
}

#   send a command to create a wallet
createWallet() {

    OUTPUT=$( perl -w ./test/createWallet.pl )

    if [ -z "$( echo ${OUTPUT} | grep 'pld successfully initialized!' )" ]
    then
        kill ${PLD_PID}
        exit "error: fail attempting to run command create wallet"
    fi

    echo ">>> create: command successfully executed"
}

#   send a command to unlock the wallet
unlockWallet() {

    OUTPUT=$( perl -w ./test/unlockWallet.pl )

    if [ -z "$( echo ${OUTPUT} | grep 'lnd successfully unlocked!' )" ]
    then
        kill ${PLD_PID}
        exit "error: fail attempting to run command unlock wallet"
    fi

    echo ">>> unlock: command successfully executed"
}

#   use pldctl to execute a command
executeCommand() {
    local COMMAND="${1}"
    local ARGUMENTS="${2}"

    OUTPUT=$( ${PLDCTL} ${PLDCTL_OPTIONS} ${COMMAND} ${ARGUMENTS} 2>> ${PLDCTL_ERRORS_FILE} )
    if [ $? == 0 ]
    then
        echo ">>> ${COMMAND} ${ARGUMENTS}: command successfully executed"
    else
        echo "error: fail attempting to run command \"${COMMAND} ${ARGUMENTS}\": $?"
    fi
}

#   splash screen
echo ">>>>> Testing pld and pldctl"
echo

#   parse CLI arguments
CREATE_WALLET="false"

while [ true ]
do
    ARG=${1}

    if [ -z "${ARG}" ]
    then
        break
    fi

    if [ "${ARG}" == "--createWallet" ]
    then
        CREATE_WALLET="true"
    fi

    shift
done

#   create wallet when requested
if [ "${PROTOC}" == "true" ]
then
    #   clean things up by removing previous wallet
    rm -rf ~/.lncli ~/.pki ~/.pktd ~/.pktwallet
    rm -rf ${PLD_OUTPUT_FILE}
    rm -rf ${PLDCTL_ERRORS_FILE}

    #   start pld deamon, create a wallet and stop the deamon, because first test is unlock wallet
    startPldDeamon
    createWallet
    stopPldDeamon
fi

#   star pld daemon and test the command to unlock the wallet
startPldDeamon
unlockWallet

#   test commands to get info about the running pld daemon
executeCommand 'getinfo'
executeCommand 'getrecoveryinfo'
#   not working command !
#executeCommand 'version'

#   test commands to deal with profile
executeCommand 'profile' 'add pld_test'
executeCommand 'profile' 'list'
executeCommand 'profile' 'setdefault pld_test'
executeCommand 'profile' 'remove pld_test'

#   remove profile file created by profile commands
rm ~/.lncli/profiles.json

#   test commands to manage channels
executeCommand 'listchannels'
executeCommand 'getnetworkinfo'
executeCommand 'feereport'

#   test commands to deal with the wallet
executeCommand 'newaddress' 'p2wkh'
executeCommand 'walletbalance'
executeCommand 'resync'
executeCommand 'getaddressbalances'
executeCommand 'getwalletseed'
executeCommand 'getsecret'
executeCommand 'getnewaddress'

#   show any eventual error during command execution
if [ -f "${PLDCTL_ERRORS_FILE}" -a $( stat --format='%s' "${PLDCTL_ERRORS_FILE}" ) -gt 0 ]
then
    echo ">>> errors executing some command "
    echo "+++++++++++++++"
    cat ${PLDCTL_ERRORS_FILE}
fi

rm -rf ${PLDCTL_ERRORS_FILE}

#   stop pld daemon
stopPldDeamon
