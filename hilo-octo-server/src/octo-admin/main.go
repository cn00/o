package main

import (
	"flag"
	"fmt"
	"log"

	"octo-admin/config"
	"octo-admin/controllers"
	"octo-admin/service"
	"octo/models"
	"octo/utils"

	"io"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"net/http"
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

	confFile := flag.String("conf", "config/admin.tml", "config toml file")
	flag.Parse()

	f, _ := os.Create("octo-admin.log")
	gin.DefaultWriter = io.MultiWriter(f)

	utils.RandSeed()

	config.Init(*confFile)
	conf := config.LoadConfig()
	engine := engine()
	controllers.InitRooter(engine)

	if err := models.Setup(models.Config{
		ReadOnly:               conf.Admin.ReadOnly,
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

	service.OauthGoogleSetup(&oauth2.Config{
		ClientID:     conf.OauthGoogle.ClientId,
		ClientSecret: conf.OauthGoogle.ClientSecret,
		RedirectURL:  conf.OauthGoogle.RedirectUrl,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	})

	s := &http.Server{
		Addr:              fmt.Sprintf(":%d", conf.Admin.Port),
		ReadTimeout:       650 * time.Second,
		ReadHeaderTimeout: 650 * time.Second,
		WriteTimeout:      650 * time.Second,
		Handler:           engine,
	}

	s.ListenAndServe()
	//	engine.Run(fmt.Sprintf(":%d", conf.Admin.Port))
}
