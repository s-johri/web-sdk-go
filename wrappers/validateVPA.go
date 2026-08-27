package wrappers

import (
	"github.com/payu-india/web-sdk-go/http"
	"github.com/payu-india/web-sdk-go/utils"
	"net/url"
)

// ValidateVPA validates a UPI VPA through the validateVPA command.
func ValidateVPA(creds utils.Creds, apiEndPoint string, var1 string) (map[string]interface{}, error) {

	command := "validateVPA"
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