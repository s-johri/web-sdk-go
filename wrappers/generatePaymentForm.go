package wrappers

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"sort"

	"github.com/payu-india/web-sdk-go/utils"
)

// formField is one hidden input in the generated payment form.
type formField struct {
	Name  string
	Value string
}

// formData is the template input for the payment form.
type formData struct {
	Action string
	Fields []formField
}

// parameterNamePattern accepts the characters PayU uses in parameter names.
// A name is an identifier rather than free text, so anything else is rejected
// instead of escaped.
var parameterNamePattern = regexp.MustCompile(`^[A-Za-z0-9_.\-]+$`)

// paymentFormTemplate renders the auto submitting payment form.
//
// html/template applies context aware escaping to every interpolated value, so
// a parameter value cannot close its attribute and inject markup.
var paymentFormTemplate = template.Must(template.New("payuPaymentForm").Parse(
	`<form name="payment_post" id="payment_post" action="{{.Action}}" method="post">` +
		`{{range .Fields}}<input hidden type="text" name="{{.Name}}" value="{{.Value}}"/>{{end}}` +
		`</form>` +
		`<script type="text/javascript"> window.onload=function(){document.forms['payment_post'].submit();}</script>`))

// GeneratePaymentForm builds an HTML form that posts a payment request to PayU
// and submits itself on load.
//
// params carries the PayU payment parameters. The merchant key and the request
// hash are added by this function; any key or hash in params is ignored. The
// caller's map is not modified.
func GeneratePaymentForm(creds utils.Creds, apiEndPoint string, params map[string]interface{}) (string, error) {
	mandatoryParams := []string{"txnid", "amount", "productinfo", "firstname", "email", "phone", "surl", "furl"}
	for _, paramName := range mandatoryParams {
		if _, ok := params[paramName]; !ok {
			return "", fmt.Errorf("missing mandatory parameter %q", paramName)
		}
	}

	names := make([]string, 0, len(params))
	for name := range params {
		// The SDK always supplies these two, so a caller supplied value is
		// dropped rather than emitted twice.
		if name == "key" || name == "hash" {
			continue
		}
		if !parameterNamePattern.MatchString(name) {
			return "", fmt.Errorf("invalid parameter name %q: names may contain only letters, digits, underscore, dot and hyphen", name)
		}
		names = append(names, name)
	}
	// Sorting makes the output reproducible, which random map order does not.
	sort.Strings(names)

	data := formData{
		Action: apiEndPoint,
		Fields: make([]formField, 0, len(names)+2),
	}
	for _, name := range names {
		data.Fields = append(data.Fields, formField{
			Name:  name,
			Value: utils.FormatValue(params[name]),
		})
	}
	data.Fields = append(data.Fields,
		formField{Name: "key", Value: creds.Key},
		formField{Name: "hash", Value: utils.PaymentHasher(creds, params)},
	)

	var rendered bytes.Buffer
	if err := paymentFormTemplate.Execute(&rendered, data); err != nil {
		return "", err
	}
	return rendered.String(), nil
}
