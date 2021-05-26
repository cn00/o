package main

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

func TestUpload_makeFilePaths(t *testing.T) {
	start := time.Now()
	var pathMap = make(map[string]string)
	var fpath = "../../src/octo-cli/commands"
	var recursion = true
	makeFilePaths(fpath, recursion, pathMap)
	assert.Equal(t, 2, len(pathMap), "Result Count not equal")
	elapsed := time.Since(start)
	log.Printf("TestUpload_makeFilePaths took %s", elapsed)

}
