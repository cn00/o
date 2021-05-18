package utils

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	"log"
	"math/rand"
)

func RandSeed() {
	var seed int64
	err := binary.Read(cryptorand.Reader, binary.LittleEndian, &seed)
	if err != nil {
		log.Panicln("Read error", err)
	}
	rand.Seed(seed)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
