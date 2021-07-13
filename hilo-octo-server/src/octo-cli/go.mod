module hilo-octo-cli

go 1.14

require (
	//aliyun-oss-go-sdk v0.0.0-00020108000000-000000000000
	cloud.google.com/go/storage v1.10.0
	github.com/BurntSushi/toml v0.3.1
	github.com/aliyun/aliyun-oss-go-sdk v2.1.8+incompatible
	github.com/aliyun/ossutil v1.7.1
	//github.com/cn00/ossutil v0.0.0-00010101000000-ab25a2c16c028f91cce57d26e015d25fca00f37c // indirect
	github.com/alyu/configparser v0.0.0-20191103060215-744e9a66e7bc // indirect
	//github.com/cheekybits/is v0.0.0-20150225183255-68e9c0620927 // indirect
	github.com/codegangsta/cli v1.20.0
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/droundy/goopt v0.0.0-20170604162106-0b8effe182da // indirect
	//github.com/matryer/try v0.0.0-20161228173917-9ac251b645a2 // indirect
	github.com/pkg/errors v0.8.0
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/tencentyun/cos-go-sdk-v5 v0.7.27
	//github.com/stretchr/testify v1.4.0
	//github.com/wataru420/contrib v1.4.0
	golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2
	google.golang.org/api v0.28.0
	gopkg.in/matryer/try.v1 v1.0.0-20150601225556-312d2599e12e
	gopkg.in/yaml.v2 v2.2.2
	hilo-octo-proto v0.0.0-00010101000000-000000000000
	octo-cli v0.0.0-00010101000000-000000000000
)

replace octo-cli => ./

replace hilo-octo-proto => ../hilo-octo-proto

//replace aliyun-oss-go-sdk => ../aliyun-oss-go-sdk
