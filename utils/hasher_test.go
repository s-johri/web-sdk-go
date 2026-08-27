package utils

import (
	"testing"
)

func testCreds() Creds {
	return Creds{Key: "testkey", Salt: "testsalt"}
}

func baseParams() map[string]interface{} {
	return map[string]interface{}{
		"txnid":       "txn001",
		"amount":      "100.00",
		"productinfo": "Book",
		"firstname":   "Alice",
		"email":       "alice@example.com",
	}
}

func TestPaymentHasherGoldenVectors(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(map[string]interface{})
		want   string
	}{
		{
			name:   "no udf fields",
			mutate: func(p map[string]interface{}) {},
			want:   "748d3ec9c31f35b1ad28a334e470d16f6d0d4cb81269949cc5f8d4ed71bde9901e1c8f7f1a5609dd2dadfeb0674a7673899bf70f570b9fb7bbfc16b6a6387ecc",
		},
		{
			name:   "udf1 and udf3 set",
			mutate: func(p map[string]interface{}) { p["udf1"] = "a"; p["udf3"] = "c" },
			want:   "4cbdc13d5865a6dd38d3de81ec7fbcd70d22558b109c111e349c87c966f3715afab08f0ca917e8594f26c5b7dc13a59fde7609b7c99b1e5c7a1cb8b679321e09",
		},
		{
			name:   "additional charges appended after the salt",
			mutate: func(p map[string]interface{}) { p["additionalCharges"] = "10.00" },
			want:   "1e741fe852c0a62f6f59fc32739ba127dd93ed490b43153abd84d46a018956031a6e4079d4c7bba86fa92badd257c5d38c14ca21963889aed0ef6ebc97c35029",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			params := baseParams()
			tc.mutate(params)
			if got := PaymentHasher(testCreds(), params); got != tc.want {
				t.Errorf("PaymentHasher = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestReverseHasherGoldenVectors(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(map[string]interface{})
		want   string
	}{
		{
			name:   "no udf fields",
			mutate: func(p map[string]interface{}) {},
			want:   "ddcfea305f9cb630268c12af1dab392b539332631723d5b030cddd968c0785cebdc197d2b129433be107788e368bc70c6c2d5c216586e90c7d779c9f448fee26",
		},
		{
			name:   "udf1 and udf3 set",
			mutate: func(p map[string]interface{}) { p["udf1"] = "a"; p["udf3"] = "c" },
			want:   "b305498cd4f1d151581e49891d508ac752e60da33b27b153c94c88c8992e2dc9c736bb17de42fce87e1f9efa0ae079717cd1e594a67834779847a80b128feda4",
		},
		{
			name:   "additional charges prepended before the salt",
			mutate: func(p map[string]interface{}) { p["additionalCharges"] = "10.00" },
			want:   "817cb9a27779cbc6cfe0e4b4af9f543471a3cc927363ba289df144ffc4587e4abc5b63858f8fb806b17d597c1c89e83c1d7c15ac0aeb591def1cea95dbba040b",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			params := baseParams()
			params["status"] = "success"
			tc.mutate(params)
			got, err := ReverseHasher(testCreds(), params)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("ReverseHasher = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestReverseHasherRequiresMandatoryParams(t *testing.T) {
	params := baseParams()
	params["status"] = "success"
	delete(params, "txnid")
	if _, err := ReverseHasher(testCreds(), params); err == nil {
		t.Error("expected an error when txnid is missing")
	}
}

// A non string udf must hash identically in both directions. Before FormatValue
// the forward hash used %v while the reverse hash used a string only type
// assertion, so an integer udf silently became the empty string in reverse.
func TestHashersAgreeOnNonStringUdf(t *testing.T) {
	forward := baseParams()
	forward["udf1"] = 7
	forwardHash := PaymentHasher(testCreds(), forward)

	equivalent := baseParams()
	equivalent["udf1"] = "7"
	if got := PaymentHasher(testCreds(), equivalent); got != forwardHash {
		t.Errorf("PaymentHasher treats int 7 and string \"7\" differently:\n int:    %s\n string: %s", forwardHash, got)
	}

	reverse := baseParams()
	reverse["status"] = "success"
	reverse["udf1"] = 7
	reverseHash, err := ReverseHasher(testCreds(), reverse)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	equivalentReverse := baseParams()
	equivalentReverse["status"] = "success"
	equivalentReverse["udf1"] = "7"
	want, err := ReverseHasher(testCreds(), equivalentReverse)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reverseHash != want {
		t.Errorf("ReverseHasher treats int 7 and string \"7\" differently:\n int:    %s\n string: %s", reverseHash, want)
	}
}

// A missing optional field must hash as the empty string, not as the literal
// "<nil>" that fmt %v produces for a nil interface value.
func TestHashersTreatMissingFieldsAsEmpty(t *testing.T) {
	withMissing := baseParams()
	delete(withMissing, "email")
	got := PaymentHasher(testCreds(), withMissing)

	withEmpty := baseParams()
	withEmpty["email"] = ""
	want := PaymentHasher(testCreds(), withEmpty)

	if got != want {
		t.Errorf("a missing email does not hash as an empty email:\n missing: %s\n empty:   %s", got, want)
	}
}
