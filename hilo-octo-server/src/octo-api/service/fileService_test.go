package service

import (
	"math/rand"
	"testing"
)

func TestMakeObjectHash(t *testing.T) {
	rand.Seed(0)
	hash := makeObjectHash()
	if hash != "cKDuHq" {
		t.Fatal("hash:", hash)
	}
}
