package wrappers

import (
	"github.com/payu-india/web-sdk-go/http"
	"github.com/payu-india/web-sdk-go/utils"
	"net/url"
)

// VerifyPayment returns the details of a transaction through the verify_payment command.
func VerifyPayment(creds utils.Creds, apiEndPoint string, var1 string) (map[string]interface{}, error) {
	command := "verify_payment"
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