package wrappers

import (
	"github.com/payu-india/web-sdk-go/http"
	"github.com/payu-india/web-sdk-go/utils"
	"net/url"
	"strconv"
)

// CheckIsDomestic reports card issuer information for a bin through the check_isDomestic command.
func CheckIsDomestic(creds utils.Creds, apiEndPoint string, bin int) (map[string]interface{}, error) {
	command := "check_isDomestic"
	var1 := strconv.Itoa(bin)
	// Create the payload
	payload := url.Values{
		"key": {creds.Key},
		"command": {command},
		"var1": {var1},
		"hash": {utils.ApiHasher(creds, utils.ApiStruct{Command: command, Var1: var1})},
	}

	// Send the request and get the response
	response, err := http.PostForm(apiEndPoint, payload)
	if err != nil {
		return nil, err
	}

	// Return the response
	return response, nil
}