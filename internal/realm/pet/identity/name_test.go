package identity

import "testing"

// TestValidateName verifies native validation result classes.
func TestValidateName(t *testing.T) {
	cases := []struct {
		value string
		code  int32
	}{{" Pixel  Dog ", NameApproved}, {"x", NameTooShort}, {"This name is much too long", NameTooLong}, {"bad!", NameInvalidCharacters}}
	for _, test := range cases {
		_, code := ValidateName(test.value)
		if code != test.code {
			t.Fatalf("value=%q code=%d expected=%d", test.value, code, test.code)
		}
	}
}

// TestNormalizeColor verifies hexadecimal normalization.
func TestNormalizeColor(t *testing.T) {
	value, err := NormalizeColor("#aabbcc")
	if err != nil || value != "AABBCC" {
		t.Fatalf("value=%q err=%v", value, err)
	}
}
