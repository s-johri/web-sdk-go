package wrappers

import (
	"github.com/payu-india/web-sdk-go/http"
	"github.com/payu-india/web-sdk-go/utils"
	"net/url"
)

// ExpireInvoice expires an existing invoice through the expire_invoice command.
func ExpireInvoice(creds utils.Creds, apiEndPoint string, var1 string) (map[string]interface{}, error) {
	command := "expire_invoice"
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