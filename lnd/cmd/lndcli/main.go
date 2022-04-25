////////////////////////////////////////////////////////////////////////////////
//	lndcli/main.go  -  Apr-8-2022  -  aldebap
//
//	Entry point for the pld client using the REST APIs
////////////////////////////////////////////////////////////////////////////////

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	jsoniter "github.com/json-iterator/go"
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

	//	only one argument means the command to be executed have no request payload
	if len(flag.Args()) == 1 {
		err := executeCommand(flag.Args()[0], "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			panic(-1)
		}
		return
	}

	//	two arguments is the command to be executed + the request payload
	if len(flag.Args()) > 1 {
		var command = flag.Args()[0]
		var requestPayload string

		requestPayload, err := formatRequestPayload(command, flag.Args()[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			panic(-1)
		}

		//	[debug] unmarshall the string with the request to marshall it with indentation
		var requestPayloadMap map[string]interface{}

		err = json.Unmarshal([]byte(requestPayload), &requestPayloadMap)
		if err != nil {
		} else {
			prettyRequestPayload, err := json.MarshalIndent(requestPayloadMap, "", "    ")
			if err != nil {
			} else {
				fmt.Fprintf(os.Stdout, "[debug]: request payload: %s\n", string(prettyRequestPayload))
			}
		}

		//	send the request payload to pld
		err = executeCommand(command, requestPayload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			panic(-1)
		}
		return
	}
}

func formatRequestPayload(command string, arguments []string) (string, error) {

	var helpURI = pldServer + URI_prefix + helpURI_prefix + "/" + command

	//	get help from pld
	response, err := http.Get(helpURI)
	if err != nil {
		return "", errors.New("fail getting command help from pld server: " + err.Error())
	}
	defer response.Body.Close()

	responseCommandHelp, err := io.ReadAll(response.Body)
	if err != nil {
		return "", errors.New("fail reading command help message from pld server: %s" + err.Error())
	}

	//	unmarshal the help message
	var commandMethod Method

	err = jsoniter.Unmarshal(responseCommandHelp, &commandMethod)
	if err != nil {
		return "", errors.New("fail unmarshling command help message: %s" + err.Error())
	}

	var requestPayload string

	if len(commandMethod.Request.Fields) > 0 {
		for _, requestField := range commandMethod.Request.Fields {

			formattedField := formatRequestField("", requestField, arguments)
			if len(formattedField) > 0 {
				if len(requestPayload) > 0 {
					requestPayload += ", "
				}
				requestPayload += formattedField
			}
		}
	}
	requestPayload = "{ " + requestPayload + " }"

	return requestPayload, nil
}

func formatRequestField(fieldHierarchy string, requestField *Field, arguments []string) string {

	var formattedField string

	if len(requestField.Type.Fields) == 0 {

		var commandOption string

		if len(fieldHierarchy) == 0 {
			commandOption = "--" + requestField.Name
		} else {
			commandOption = "--" + fieldHierarchy + "." + requestField.Name
		}

		switch requestField.Type.Name {
		case "bool":
			for _, argument := range arguments {
				if argument == commandOption {
					formattedField += "\"" + requestField.Name + "\": true"
				}
			}

		case "[]byte":
			commandOption += "="
			for _, argument := range arguments {
				if strings.HasPrefix(argument, commandOption) {
					formattedField += "\"" + requestField.Name + "\": \"" + argument[len(commandOption):] + "\""
				}
			}

		case "string":
			commandOption += "="
			for _, argument := range arguments {
				if strings.HasPrefix(argument, commandOption) {
					if !requestField.Repeated {
						formattedField += "\"" + requestField.Name + "\": \"" + argument[len(commandOption):] + "\""
					} else {
						var arrayOfStrings string

						for _, stringValue := range strings.Split(argument[len(commandOption):], ":") {
							if len(arrayOfStrings) > 0 {
								arrayOfStrings += ", "
							}
							arrayOfStrings += "\"" + stringValue + "\""
						}

						formattedField += "\"" + requestField.Name + "\": [ " + arrayOfStrings + " ]"
					}
				}
			}

		//	TODO: to make sure that for integer types pld doen't have arrays (Repeated == true)
		case "uint32":
			commandOption += "="
			for _, argument := range arguments {
				if strings.HasPrefix(argument, commandOption) {
					formattedField += "\"" + requestField.Name + "\": " + argument[len(commandOption):]
				}
			}

		case "int32":
			commandOption += "="
			for _, argument := range arguments {
				if strings.HasPrefix(argument, commandOption) {
					formattedField += "\"" + requestField.Name + "\": " + argument[len(commandOption):]
				}
			}

		case "uint64":
			commandOption += "="
			for _, argument := range arguments {
				if strings.HasPrefix(argument, commandOption) {
					formattedField += "\"" + requestField.Name + "\": " + argument[len(commandOption):]
				}
			}

		case "int64":
			commandOption += "="
			for _, argument := range arguments {
				if strings.HasPrefix(argument, commandOption) {
					formattedField += "\"" + requestField.Name + "\": " + argument[len(commandOption):]
				}
			}

		case "float64":
			commandOption += "="
			for _, argument := range arguments {
				if strings.HasPrefix(argument, commandOption) {
					formattedField += "\"" + requestField.Name + "\": " + argument[len(commandOption):]
				}
			}

		//	map some hard coded enums
		case "ENUM_VARIENT":
			for _, argument := range arguments {
				if argument == commandOption {
					formattedField += "\"" + requestField.Name + "\""
				}
			}
		}
	} else {

		var formattedSubFields string

		for _, requestSubField := range requestField.Type.Fields {
			var formattedSubField string

			if len(fieldHierarchy) == 0 {
				formattedSubField = formatRequestField(requestField.Name, requestSubField, arguments)
			} else {
				formattedSubField = formatRequestField(fieldHierarchy+"."+requestField.Name, requestSubField, arguments)
			}

			if len(formattedSubField) > 0 {
				if len(formattedSubFields) > 0 {
					formattedSubFields += ", "
				}
				formattedSubFields += formattedSubField
			}
		}
		if len(formattedSubFields) > 0 {
			if !requestField.Repeated {
				formattedField += "\"" + requestField.Name + "\": { " + formattedSubFields + " }"
			} else {
				formattedField += "\"" + requestField.Name + "\": [ " + formattedSubFields + " ]"
			}
		}
	}

	return formattedField
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
