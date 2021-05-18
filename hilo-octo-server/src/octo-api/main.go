package main

import (
	_ "expvar"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"octo-api/cache"
	"octo-api/config"
	"octo-api/controllers"
	"octo-api/service"
	"octo/models"
	"octo/utils"
	"time"
)

var octoVersion string
var octoRevision string

func main() {

	log.Println("/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/")
	log.Println("")
	log.Println("   ____  ______ ______ ____ ")
	log.Println("  / __ \\/ ____/__  __/ __  \\")
	log.Println(" / /_/ / /___   / / / /_/ /     version: ", octoVersion)
	log.Println(" \\____/_____/  /_/  \\____/     revision: ", octoRevision)
	log.Println("")
	log.Println("/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/_/")

	confFile := flag.String("conf", "config/api.tml", "config toml file")
	flag.Parse()

	utils.RandSeed()

	config.Init(*confFile)
	conf := config.LoadConfig()
	go metrics(conf)
	engine := engine()

	controllers.InitRooter(engine)
	cache.Setup()
	service.Setup(service.Config{
		CacheAppsListAPI: conf.CacheApps.ListAPI,
	})

	if err := models.Setup(models.Config{
		ReadOnly:               conf.Api.ReadOnly,
		DatabaseMasterAddrs:    conf.Database.Master.Addrs,
		DatabaseMasterDbname:   conf.Database.Master.Dbname,
		DatabaseMasterUser:     conf.Database.Master.User,
		DatabaseMasterPassword: conf.Database.Master.Password,
		DatabaseSlaveAddrs:     conf.Database.Slave.Addrs,
		DatabaseSlaveDbname:    conf.Database.Slave.Dbname,
		DatabaseSlaveUser:      conf.Database.Slave.User,
		DatabaseSlavePassword:  conf.Database.Slave.Password,
	}); err != nil {
		panic(err)
	}

	s := &http.Server{
		Addr:              fmt.Sprintf(":%d", conf.Api.Port),
		ReadTimeout:       650 * time.Second,
		ReadHeaderTimeout: 650 * time.Second,
		WriteTimeout:      650 * time.Second,
		Handler:           engine,
	}

	s.ListenAndServe()
	//engine.Run(fmt.Sprintf(":%d", conf.Api.Port))

}

func metrics(conf *config.Config) {
	addr := fmt.Sprintf(":%d", conf.Metrics.Port)
	log.Println(http.ListenAndServe(addr, nil))
}
