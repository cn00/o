# octo-api

## ローカルで動作確認する方法

```
$ make
$ make all
$ make api
$ make admin
```

でビルドできます。
VERSION情報をつける場合は、 `make all VERSION=*.*.*` ですがlocalで動かす分には不要です。
ビルド後に各バイナリを起動することで、アクセスできます。

`octo-api`
http://localhost:8080/

`octo-admin`
http://localhost:8081/

動作にはmysqlが必要で、admin.tml の情報を参照して適宜DBを作成してください
テーブルやレコード情報は `./sql`

## GCPに上げる方法

[OCTO Jenkins](https://jenkins.octo-cloud.com/) をご利用ください。
