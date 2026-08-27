package wrappers

import (
	"html/template"
	"strings"
	"testing"

	"github.com/payu-india/web-sdk-go/utils"
)

const testEndpoint = "https://test.payu.in/_payment"

func testCreds() utils.Creds {
	return utils.Creds{Key: "testkey", Salt: "testsalt"}
}

func validParams() map[string]interface{} {
	return map[string]interface{}{
		"txnid":       "txn001",
		"amount":      "100.00",
		"productinfo": "Book",
		"firstname":   "Alice",
		"email":       "alice@example.com",
		"phone":       "9999999999",
		"surl":        "https://merchant.example/success",
		"furl":        "https://merchant.example/failure",
	}
}

// formBody returns the markup between the opening form tag and </form>, which
// is where injected markup would land.
func formBody(t *testing.T, form string) string {
	t.Helper()
	body, _, found := strings.Cut(form, "</form>")
	if !found {
		t.Fatalf("form has no closing tag: %s", form)
	}
	return body
}

// An attacker controlled value must never add elements to the form or escape
// its attribute. The structural assertion is deliberate: counting elements
// catches any breakout, whatever the payload looks like.
func TestGeneratePaymentFormEscapesValues(t *testing.T) {
	payloads := []string{
		`"`,
		`" onfocus="alert(1)`,
		`Book"><script>alert(1)</script>`,
		`Book"><input name="surl" value="https://attacker.example`,
		`'`,
		`<b>bold</b>`,
		`a & b`,
		"line\nbreak",
		"“fancy quotes”",
		`</form>`,
	}

	clean, err := GeneratePaymentForm(testCreds(), testEndpoint, validParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantInputs := strings.Count(formBody(t, clean), "<input")

	for _, payload := range payloads {
		t.Run(payload, func(t *testing.T) {
			params := validParams()
			params["productinfo"] = payload

			form, err := GeneratePaymentForm(testCreds(), testEndpoint, params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			body := formBody(t, form)

			if got := strings.Count(body, "<input"); got != wantInputs {
				t.Errorf("payload changed the input count from %d to %d\nform: %s", wantInputs, got, form)
			}
			if strings.Contains(body, "<script") {
				t.Errorf("payload injected a script element\nform: %s", form)
			}
			// Assert the exact escaping rather than merely the absence of the
			// raw payload. Characters that are not HTML-special, such as
			// curly quotes, are correctly emitted verbatim: they cannot close
			// an ASCII-quoted attribute, and a product name should render as
			// the merchant typed it.
			wantValue := template.HTMLEscapeString(payload)
			if !strings.Contains(body, `value="`+wantValue+`"`) {
				t.Errorf("productinfo was not rendered as the escaped payload %q\nform: %s", wantValue, form)
			}
		})
	}
}

// A parameter name is an identifier, not free text. Rejecting anything else
// closes the key injection route at its source.
func TestGeneratePaymentFormRejectsInvalidParameterNames(t *testing.T) {
	names := []string{
		`udf1" value="x"><input name="surl`,
		`name with spaces`,
		`quote"`,
		`<tag>`,
		`semi;colon`,
		``,
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			params := validParams()
			params[name] = "value"

			form, err := GeneratePaymentForm(testCreds(), testEndpoint, params)
			if err == nil {
				t.Errorf("expected an error for parameter name %q, got form: %s", name, form)
			}
			if form != "" {
				t.Errorf("expected empty output on error, got: %s", form)
			}
		})
	}
}

func TestGeneratePaymentFormAcceptsValidParameterNames(t *testing.T) {
	for _, name := range []string{"udf1", "address_1", "sub-merchant", "api.version", "UDF5"} {
		t.Run(name, func(t *testing.T) {
			params := validParams()
			params[name] = "value"
			if _, err := GeneratePaymentForm(testCreds(), testEndpoint, params); err != nil {
				t.Errorf("valid name %q was rejected: %v", name, err)
			}
		})
	}
}

// The central regression test. The value written into the form must be the
// same string the hash was computed over, otherwise PayU rejects the payment.
func TestFormValueMatchesHashedValue(t *testing.T) {
	cases := []struct {
		name   string
		amount interface{}
		want   string
	}{
		{"string amount", "100.00", "100.00"},
		{"float amount", 100.5, "100.5"},
		{"whole float amount", 100.0, "100"},
		{"large float amount", 1000000.0, "1000000"},
		{"int amount", 100, "100"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := utils.FormatValue(tc.amount); got != tc.want {
				t.Fatalf("FormatValue(%#v) = %q, want %q", tc.amount, got, tc.want)
			}

			params := validParams()
			params["amount"] = tc.amount
			form, err := GeneratePaymentForm(testCreds(), testEndpoint, params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			wantField := `name="amount" value="` + tc.want + `"`
			if !strings.Contains(form, wantField) {
				t.Errorf("form is missing %s\nform: %s", wantField, form)
			}
		})
	}
}

func TestGeneratePaymentFormRendersNonStringTypesCleanly(t *testing.T) {
	params := validParams()
	params["udf1"] = true
	params["udf2"] = 42

	form, err := GeneratePaymentForm(testCreds(), testEndpoint, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(form, "%!") {
		t.Errorf("form contains a Go formatting error verb\nform: %s", form)
	}
	if !strings.Contains(form, `name="udf1" value="true"`) {
		t.Errorf("bool udf1 not rendered as true\nform: %s", form)
	}
	if !strings.Contains(form, `name="udf2" value="42"`) {
		t.Errorf("int udf2 not rendered as 42\nform: %s", form)
	}
}

func TestGeneratePaymentFormDoesNotMutateInput(t *testing.T) {
	params := validParams()
	before := len(params)

	if _, err := GeneratePaymentForm(testCreds(), testEndpoint, params); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(params) != before {
		t.Errorf("params grew from %d to %d entries", before, len(params))
	}
	if _, ok := params["hash"]; ok {
		t.Error("hash was written into the caller's map")
	}
	if _, ok := params["key"]; ok {
		t.Error("key was written into the caller's map")
	}
}

func TestGeneratePaymentFormIsDeterministic(t *testing.T) {
	first, err := GeneratePaymentForm(testCreds(), testEndpoint, validParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i := 0; i < 20; i++ {
		next, err := GeneratePaymentForm(testCreds(), testEndpoint, validParams())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if next != first {
			t.Fatalf("output changed between runs:\nfirst: %s\nnext:  %s", first, next)
		}
	}
}

// html/template rewrites script contents in some contexts. The auto submit
// behaviour is the whole point of this function, so pin it exactly.
func TestGeneratePaymentFormKeepsAutoSubmitScript(t *testing.T) {
	form, err := GeneratePaymentForm(testCreds(), testEndpoint, validParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	const want = `<script type="text/javascript"> window.onload=function(){document.forms['payment_post'].submit();}</script>`
	if !strings.Contains(form, want) {
		t.Errorf("auto submit script missing or altered\nform: %s", form)
	}
}

func TestGeneratePaymentFormIncludesKeyAndHash(t *testing.T) {
	params := validParams()
	wantHash := utils.PaymentHasher(testCreds(), params)

	form, err := GeneratePaymentForm(testCreds(), testEndpoint, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(form, `name="key" value="testkey"`) {
		t.Errorf("merchant key missing\nform: %s", form)
	}
	if !strings.Contains(form, `name="hash" value="`+wantHash+`"`) {
		t.Errorf("payment hash missing or wrong\nform: %s", form)
	}
}

// A caller supplied hash or key must not override the values the SDK computes.
func TestGeneratePaymentFormOverridesCallerHashAndKey(t *testing.T) {
	params := validParams()
	params["hash"] = "attacker-supplied"
	params["key"] = "attacker-key"

	form, err := GeneratePaymentForm(testCreds(), testEndpoint, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(form, "attacker-supplied") || strings.Contains(form, "attacker-key") {
		t.Errorf("caller supplied hash or key leaked into the form\nform: %s", form)
	}
	if got := strings.Count(form, `name="hash"`); got != 1 {
		t.Errorf("expected exactly one hash field, got %d\nform: %s", got, form)
	}
	if got := strings.Count(form, `name="key"`); got != 1 {
		t.Errorf("expected exactly one key field, got %d\nform: %s", got, form)
	}
}

func TestGeneratePaymentFormRequiresMandatoryParams(t *testing.T) {
	for _, missing := range []string{"txnid", "amount", "productinfo", "firstname", "email", "phone", "surl", "furl"} {
		t.Run(missing, func(t *testing.T) {
			params := validParams()
			delete(params, missing)
			if _, err := GeneratePaymentForm(testCreds(), testEndpoint, params); err == nil {
				t.Errorf("expected an error when %q is missing", missing)
			}
		})
	}
}
