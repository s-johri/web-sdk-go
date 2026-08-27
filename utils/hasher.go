package utils

import (
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
)

func HashString(rawString string) string {
	sha := sha512.New()
	sha.Write([]byte(rawString))
	return hex.EncodeToString(sha.Sum(nil))
}

// PaymentHasher builds the SHA-512 request hash for a payment.
//
// Layout:
//
//	key|txnid|amount|productinfo|firstname|email|udf1|udf2|udf3|udf4|udf5||||||salt
//
// When additionalCharges is present it is appended after the salt.
func PaymentHasher(credes Creds, params map[string]interface{}) string {
	var udfFields string
	for i := 1; i <= 5; i++ {
		udfFields += FormatValue(params[fmt.Sprintf("udf%d", i)]) + "|"
	}

	hashString := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|||||%s",
		credes.Key,
		FormatValue(params["txnid"]),
		FormatValue(params["amount"]),
		FormatValue(params["productinfo"]),
		FormatValue(params["firstname"]),
		FormatValue(params["email"]),
		udfFields,
		credes.Salt,
	)

	if _, ok := params["additionalCharges"]; ok {
		hashString += "|" + FormatValue(params["additionalCharges"])
	}
	return HashString(hashString)
}

// ReverseHasher builds the SHA-512 response hash that PayU returns, so a
// merchant can verify a webhook or a redirect.
//
// Layout:
//
//	salt|status||||||udf5|udf4|udf3|udf2|udf1|email|firstname|productinfo|amount|txnid|key
//
// When additionalCharges is present it is prepended before the salt.
func ReverseHasher(credes Creds, params map[string]interface{}) (string, error) {
	mandatoryParams := []string{"status", "txnid", "amount", "productinfo", "firstname"}
	for _, paramName := range mandatoryParams {
		if _, ok := params[paramName]; !ok {
			return "", fmt.Errorf("missing mandatory parameter %q", paramName)
		}
	}

	hashString := fmt.Sprintf("%s|%s||||||%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s",
		credes.Salt,
		FormatValue(params["status"]),
		FormatValue(params["udf5"]),
		FormatValue(params["udf4"]),
		FormatValue(params["udf3"]),
		FormatValue(params["udf2"]),
		FormatValue(params["udf1"]),
		FormatValue(params["email"]),
		FormatValue(params["firstname"]),
		FormatValue(params["productinfo"]),
		FormatValue(params["amount"]),
		FormatValue(params["txnid"]),
		credes.Key,
	)

	if _, ok := params["additionalCharges"]; ok {
		hashString = FormatValue(params["additionalCharges"]) + "|" + hashString
	}
	return HashString(hashString), nil
}

// CheckReversehash verifies the hash PayU returned with a payment response.
//
// The comparison is constant time, so a caller cannot learn the expected hash
// byte by byte from response timing. It is also case insensitive, because the
// hash is hex and different PayU surfaces differ in case.
func CheckReversehash(credes Creds, params map[string]interface{}) (bool, error) {
	expected, err := ReverseHasher(credes, params)
	if err != nil {
		return false, err
	}

	received, ok := params["hash"].(string)
	if !ok {
		return false, fmt.Errorf("missing mandatory parameter %q", "hash")
	}

	expectedBytes := []byte(strings.ToLower(expected))
	receivedBytes := []byte(strings.ToLower(received))
	return subtle.ConstantTimeCompare(expectedBytes, receivedBytes) == 1, nil
}

func ApiHasher(cred Creds, params ApiStruct) string {
	rawString := fmt.Sprintf("%v|%v|%v|%v", cred.Key, params.Command, params.Var1, cred.Salt)
	return HashString(rawString)
}
