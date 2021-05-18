package service

import (
	"log"
	"math/rand"

	"octo/models"

	"github.com/QualiArts/hilo-octo-proto/go/octo"
)

var fileDao = models.NewFileDao()
var fileUrlDao = &models.FileUrlDao{}
var resourceDao = models.NewResourceDao()
var resourceUrlDao = &models.ResourceUrlDao{}
var versionDao = &models.VersionDao{}
var tagDao = &models.TagDao{}
var gcsDao = &models.GcsDao{}
var bucketDao = &models.BucketDao{}

func getDataState(state int) *octo.Data_State {
	s := octo.Data_State(state)
	switch s {
	case octo.Data_ADD:
	case octo.Data_UPDATE:
	case octo.Data_DELETE:
	default:
		log.Panic("invalid state:", state)
	}
	return s.Enum()
}

func makeObjectHash() string {
	strlen := 6 //TODO config file
	const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
