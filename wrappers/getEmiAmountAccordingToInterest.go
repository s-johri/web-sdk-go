package wrappers

import (
	"net/url"
	"github.com/payu-india/web-sdk-go/http"
	"github.com/payu-india/web-sdk-go/utils"
	"strconv"
)

// GetEmiAmountAccordingToInterest returns EMI interest by bank and tenure through the getEmiAmountAccordingToInterest command.
func GetEmiAmountAccordingToInterest(creds utils.Creds, apiEndPoint string, amount float64) (map[string]interface{}, error) {
	command := "getEmiAmountAccordingToInterest"
	// Create the payload
	var1 := strconv.FormatFloat(amount, 'f', 2, 64)
	payload := url.Values{
		"key":     {creds.Key},
		"command": {command},
		"var1":    {var1},
		"hash":    {utils.ApiHasher(creds, utils.ApiStruct{Command: command, Var1: var1})},
	}

	// Send the request and get the response
	response, err := http.PostForm(apiEndPoint, payload)
	if err != nil {
		return nil, err
	}
	// Return the response
	return response, nil
}
