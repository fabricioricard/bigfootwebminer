////////////////////////////////////////////////////////////////////////////////
//	lndcli/lnd_help.go  -  Apr-12-2022  -  aldebap
//
//	Invoke the pld REST help and show it in a fancy way for the CLI
////////////////////////////////////////////////////////////////////////////////

package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

//	structs from lnd/lnrpc/restrpc/rest.proto
type restMasterHelpResponse struct {
	Name        string                          `json:"name,omitempty"`
	Description []string                        `json:"description,omitempty"`
	Category    map[string]*restCommandCategory `json:"category,omitempty"`
}

type restCommandCategory struct {
	Description []string                        `json:"description,omitempty"`
	Endpoints   map[string]string               `json:"endpoints,omitempty"`
	Subcategory map[string]*restCommandCategory `json:"subcategory,omitempty"`
}

//	structs from lnd/pkthelp/pkthelp.go
type Method struct {
	Name        string   `json:"name,omitempty"`
	Service     string   `json:"service,omitempty"`
	Description []string `json:"description,omitempty"`
	Request     *Type    `json:"request,omitempty"`
	Response    *Type    `json:"response,omitempty"`
}

type Type struct {
	Name        string   `json:"name,omitempty"`
	Description []string `json:"description,omitempty"`
	Fields      []*Field `json:"fields,omitempty"`
}

type Field struct {
	Name        string   `json:"name,omitempty"`
	Description []string `json:"description,omitempty"`
	Repeated    bool     `json:"repeated,omitempty"`
	Type        *Type    `json:"type,omitempty"`
}

func getMasterHelp() error {

	masterHelpURI := pldServer + URI_prefix

	//	get master help from pld
	response, err := http.Get(masterHelpURI)
	if err != nil {
		return errors.New("fail getting master help from pld server: " + err.Error())
	}
	defer response.Body.Close()

	responseMasterHelp, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail reading master help message from pld server: %s", err)
		panic(-1)
	}

	//	unmarshal the master help message
	var masterHelp restMasterHelpResponse

	err = jsoniter.Unmarshal(responseMasterHelp, &masterHelp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail unmarshling master help message: %s", err)
		panic(-1)
	}

	//	show a fancy output for the master help
	fmt.Fprintf(os.Stdout, "NAME:\n    %s\n\n", masterHelp.Name)

	fmt.Fprintf(os.Stdout, "DESCRIPTION:\n")
	for _, line := range masterHelp.Description {
		fmt.Fprintf(os.Stdout, "    %s\n", line)
	}
	fmt.Fprintf(os.Stdout, "\n")

	fmt.Fprintf(os.Stdout, "CATEGORY:\n")
	for categoryName, category := range masterHelp.Category {
		showCategory(categoryName, category, 1)
	}

	return nil
}

func showCategory(name string, category *restCommandCategory, level int) {

	var firstIndent string
	var secondIndent string

	for i := 1; i <= level; i++ {
		firstIndent = firstIndent + "    "
	}
	secondIndent = firstIndent + "    "

	fmt.Fprintf(os.Stdout, "%s%s:\n\n", firstIndent, name)

	fmt.Fprintf(os.Stdout, "%sDESCRIPTION:\n", firstIndent)
	for _, line := range category.Description {
		fmt.Fprintf(os.Stdout, "%s%s\n", secondIndent, line)
	}
	fmt.Fprintf(os.Stdout, "\n")

	if len(category.Endpoints) > 0 {
		fmt.Fprintf(os.Stdout, "%sCOMMANDS:\n", firstIndent)
		for endpoint, description := range category.Endpoints {
			var command = endpoint

			if strings.HasPrefix(command, URI_prefix+"/") {
				command = command[len(URI_prefix)+1:]
			}
			fmt.Fprintf(os.Stdout, "%s%s: %s\n", secondIndent, command, description)
		}
		fmt.Fprintf(os.Stdout, "\n")
	}

	if len(category.Subcategory) > 0 {

		fmt.Fprintf(os.Stdout, "%sSUBCATEGORY:\n", firstIndent)
		for subcategoryName, subcategory := range category.Subcategory {
			showCategory(subcategoryName, subcategory, level+1)
		}
	}
}

func getCommandHelp(commandHelp string) error {

	var helpURI string

	if len(commandHelp) == 0 {
		helpURI = pldServer
	} else {
		helpURI = pldServer + URI_prefix + helpURI_prefix + "/" + commandHelp
	}

	//	get help from pld
	response, err := http.Get(helpURI)
	if err != nil {
		return errors.New("fail getting command help from pld server: " + err.Error())
	}
	defer response.Body.Close()

	responseCommandHelp, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail reading command help message from pld server: %s", err)
		panic(-1)
	}

	//	unmarshal the master help message
	var commandMethod Method

	err = jsoniter.Unmarshal(responseCommandHelp, &commandMethod)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fail unmarshling command help message: %s", err)
		panic(-1)
	}

	//	show a fancy output for the master help
	fmt.Fprintf(os.Stdout, "NAME:\n    %s: %s\n\n", commandMethod.Name, commandHelp)

	fmt.Fprintf(os.Stdout, "DESCRIPTION:\n")
	for _, line := range commandMethod.Description {
		fmt.Fprintf(os.Stdout, "    %s\n", line)
	}
	fmt.Fprintf(os.Stdout, "\n")

	fmt.Fprintf(os.Stdout, "SERVICE:\n    %s\n\n", commandMethod.Service)

	if len(commandMethod.Request.Fields) > 0 {
		fmt.Fprintf(os.Stdout, "OPTIONS:\n")
		for _, requestField := range commandMethod.Request.Fields {
			showField("", requestField)
		}
	}

	return nil
}

func showField(fieldHierarchy string, requestField *Field) {

	if len(requestField.Type.Fields) == 0 {

		var commandOption string

		if len(fieldHierarchy) == 0 {
			commandOption = "--" + requestField.Name
		} else {
			commandOption = "--" + fieldHierarchy + "." + requestField.Name
		}

		switch requestField.Type.Name {
		case "bool":

		case "[]byte":
			commandOption += "=value"
		case "string":
			commandOption += "=value"
		case "uint32":
			commandOption += "=value"
		case "int32":
			commandOption += "=value"
		case "uint64":
			commandOption += "=value"
		case "int64":
			commandOption += "=value"
		}

		if len(requestField.Description) == 0 {
			fmt.Fprintf(os.Stdout, "    %s\n", commandOption)
		} else {
			fmt.Fprintf(os.Stdout, "    %s - %s\n", commandOption, requestField.Description[0])
		}
	} else {

		for _, requestSubField := range requestField.Type.Fields {
			if len(fieldHierarchy) == 0 {
				showField(requestField.Name, requestSubField)
			} else {
				showField(fieldHierarchy+"."+requestField.Name, requestSubField)
			}
		}
	}
}
