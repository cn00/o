package controllers

import (
	"math"
	"testing"
)

func TestParseCliVersion(t *testing.T) {
	cliVersion, err := parseCliVersion("v1.2")
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(cliVersion-1.2) > math.SmallestNonzeroFloat64 {
		t.Fatal("wrong cliVersion:", cliVersion)
	}
}
