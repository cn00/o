module github.com/QualiArts/hilo-octo-server

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/ahmetb/go-linq/v3 v3.2.0 // indirect
	github.com/boj/redistore v0.0.0-20180917114910-cd5dcc76aeff
	github.com/davecgh/go-spew v1.1.1
	github.com/dgrijalva/jwt-go v2.4.1-0.20151222215319-2240de772c17+incompatible
	github.com/dustin/go-broadcast v0.0.0-20140627040055-3bdf6d4a7164
	github.com/garyburd/redigo v0.0.0-20151029235527-6ece6e0a09f2
	github.com/gin-gonic/gin v1.6.3
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.4.2
	github.com/gorilla/context v1.1.1
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/sessions v1.2.0
	github.com/hashicorp/golang-lru v0.5.4
	github.com/kisielk/sqlstruct v0.0.0-20150923205031-648daed35d49
	github.com/manucorporat/sse v0.0.0-20150715184805-fe6ea2c8e398
	github.com/manucorporat/stats v0.0.0-20150531204625-8f2d6ace262e
	github.com/mattn/go-colorable v0.0.0-20150625154642-40e4aedc8fab
	github.com/mattn/go-isatty v0.0.12
	github.com/mattn/go-sqlite3 v1.14.7 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/stretchr/objx v0.1.2-0.20180129172003-8a3f7159479f
	github.com/stretchr/testify v1.5.1
	github.com/ugorji/go-codec-bench v0.0.0-20180131102424-deae4129ac4e
	github.com/wataru420/contrib v0.0.0-20151203083238-682ac49274a3
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	google.golang.org/api v0.28.0
	google.golang.org/cloud v0.0.0-20151221053510-6fdcab499d2c
	gopkg.in/DATA-DOG/go-sqlmock.v1 v1.3.0
	gopkg.in/asn1-ber.v1 v1.0.0-20170511165959-379148ca0225
	gopkg.in/bluesuncorp/validator.v5 v5.10.4-0.20150804013332-98121ac23ff3
	gopkg.in/go-playground/validator.v8 v8.18.2
	gopkg.in/yaml.v2 v2.2.8
	hilo-octo-proto v0.0.0-00010101000000-000000000000
	octo v0.0.0-00010101000000-000000000000
	octo-admin v0.0.0-00010101000000-000000000000 // indirect
	octo-api v0.0.0-00010101000000-000000000000
	octo-cli v0.0.0-00010101000000-000000000000
)

replace octo => ./src/octo

replace octo-api => ./src/octo-api

replace octo-cli => ./src/octo-cli

replace octo-admin => ./src/octo-admin

replace hilo-octo-proto => ./src/hilo-octo-proto
