package billing

import (
	"testing"
)

func TestRegex(t *testing.T) {
	f := "billing-2019-01-18.json"
	actual := extractDate(f)
	expected := "2019-01-18"
	if actual != expected {
		t.Errorf("actual %v\nwant %v", actual, expected)
	}
}
