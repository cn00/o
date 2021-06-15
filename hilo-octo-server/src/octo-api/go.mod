module octo-api

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/gin-gonic/gin v1.6.3
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.4.1
	github.com/hashicorp/golang-lru v0.5.4
	github.com/mattn/go-sqlite3 v1.14.7 // indirect
	github.com/pkg/errors v0.9.1
	hilo-octo-proto v0.0.0-00010101000000-000000000000
	octo v0.0.0-00010101000000-000000000000
)

replace (
	hilo-octo-proto => ../hilo-octo-proto
	octo => ../octo
	octo-api => ./
)
