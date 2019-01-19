package billing

import (
	"fmt"
	"testing"
)

func TestRegex(t *testing.T) {
	expected := "2019-01-18"
	fileName := fmt.Sprintf("billing-%s.json", expected)
	actual := extractDate(fileName)
	if actual != expected {
		t.Errorf("actual %v\nwant %v", actual, expected)
	}
}
