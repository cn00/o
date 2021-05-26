module octo-admin

go 1.14

require (
       github.com/BurntSushi/toml v0.3.1

       github.com/boj/redistore v0.0.0-20180917114910-cd5dcc76aeff // indirect
       github.com/gin-gonic/gin v1.6.3
       github.com/go-sql-driver/mysql v1.5.0
       github.com/gorilla/sessions v1.2.0 // indirect
       github.com/pkg/errors v0.9.1
       //github.com/wataru420/contrib v0.0.0-20151203073237-6e7c2656fddf
       golang.org/x/net v0.0.0-20200506145744-7e3656a0809f
       golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
       google.golang.org/api v0.23.0
       octo v0.0.0-00010101000000-000000000000
       hilo-octo-proto v0.0.0-00010101000000-000000000000
)

replace octo => ../octo
replace hilo-octo-proto => ../hilo-octo-proto