////////////////////////////////////////////////////////////////////////////////
//	lndcli/main.go  -  Apr-8-2022  -  aldebap
//
//	Entry point for the pld client using the REST APIs
////////////////////////////////////////////////////////////////////////////////

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	URI_prefix     = "/api/v1"
	helpURI_prefix = "/help"
)

var (
	pldServer string
)

func main() {
	var help bool

	//	parse command line arguments
	flag.StringVar(&pldServer, "pld_server", "http://localhost:8080", "set the pld server URL")
	flag.BoolVar(&help, "help", false, "get help on a specific command")

	flag.Parse()

	//	if a protocol is missing from pld_server, assume HTTP as default
	if !strings.HasPrefix(pldServer, "http://") && !strings.HasPrefix(pldServer, "https://") {
		pldServer = "http://" + pldServer
	}

	//	check if the user wants a command help
	if help {
		var err error

		if len(flag.Args()) == 0 {
			err = getMasterHelp()
		} else if len(flag.Args()) == 1 {
			err = getCommandHelp(flag.Args()[0])
		} else {
			fmt.Fprintf(os.Stderr, "error: unexpected arguments for help on command %s", flag.Args()[0])
			panic(-1)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			panic(-1)
		}
		return
	}

	//	only on argument is the command to be executed
	if len(flag.Args()) == 1 {
		err := executeCommand(flag.Args()[0], "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			panic(-1)
		}
		return
	}

	//	two arguments is the command to be executed + the request payload
	if len(flag.Args()) == 2 {
		err := executeCommand(flag.Args()[0], flag.Args()[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			panic(-1)
		}
		return
	}
}

func executeCommand(command string, payload string) error {

	var response *http.Response
	var err error

	commandURI := pldServer + URI_prefix + "/" + command

	//	if there's no payload, use HTTP GET method to invoke pld command, otherwise use POST method
	if len(payload) == 0 {
		response, err = http.Get(commandURI)
		if err != nil {
			return errors.New("fail executing pld command: " + err.Error())
		}
	} else {
		response, err = http.Post(commandURI, "application/json", strings.NewReader(payload))
		if err != nil {
			return errors.New("fail executing pld command: " + err.Error())
		}
	}
	defer response.Body.Close()

	responsePayload, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail reading command response payload from pld server: %s", err)
		panic(-1)
	}
	fmt.Fprintf(os.Stdout, "%s\n", responsePayload)

	return nil
}
