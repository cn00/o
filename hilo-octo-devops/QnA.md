# OCTO Server Q&A

## GCS関連

### Q. Google Storage Bucket名は自動生成されますか？

- ### A. ファイルアップロードするとき自動生成されます。

### Q. ファイルアップロード時、Bucket名が重複になるとどうなりますか？

- ### A. エラーが発生しファイルアップロードが中断されます。

### Q. Bucket名が重複する場合、回避する方法を教えてください。

- ### A. OCTOのAppの設定でBucket名を変更するか他のVersionを利用してください。

## DB関連

### Q. 作られるConnection数を教えてください。

- ### A. octoはgoのsql driverを利用してます。恐らく、サーバーのリソースが利用できる限り、無制限で利用できるはずです。

### Q. Connection pooling利用可否について教えてください。

- ### A. goのsql driverはconnection poolingを支援しています。

### Q. OCTOのMySQL Masterはなぜ、フェールオーバーを考慮してないですか？

- ### A. Masterは書き込み専用で、slaveだけ生きていればmasterが落ちてもユーザーへの影響がないからです。それでmasterを直ちに復旧させる必要性がなく、手動で復旧すれば良いということです。

### Q. DBサーバーはAutoscaleしている状態ですか？

- ### A. いいえ行なっておりません。

### Q. DBの構成について教えてください。

- ### A. 構成とSpecは下記となります。

  - Common Spec
    - Machine : n2-highmem-2(Process x2, Memory 13G)
    - OS : CentOS 5.6
    - DB : MySQL 5.6
  - Master Spec
    - Disk : hdd 10G
  - Slave Spec
    - Disk : ssd 10G

### Q. DBの容量は10Gで十分ですか？

- ### A. 最初運用では十分だと思います。ただし、プロジェクト運用上、もっと容量が必要な場合、それにに合わせて設定してください。

### Q. Master command COM_REGISTER_SLAVE failed: Access denied for user 'octo'@'xx.xx.%' (using password: YES) (Errno: 1045)というエラーメッセージが表示されSlaveからMasterへアクセスができません。

- ### A. 下記の手順で実行して試してみてください。

  - ① MasterでSlave用ユーザー作成：`GRANT REPLICATION SLAVE ON *.* TO 'slave_user'@'10.0.0.2';`

  - ② SlaveでChange master実行：`CHANGE MASTER TO MASTER_HOST='10.0.0.4', MASTER_PORT=3306, MASTER_USER='slave_user', MASTER_PASSWORD='', MASTER_LOG_FILE='mysql-bin.000000', MASTER_LOG_POS=100;`
  
 ### Q. panic: dbm check failed: tcpmulti: dial failed: dial tcp 10.146.0.2:3306: i/o timeout goroutine 1 [running]: octo/models.Setupというエラーが発生してgkeのpodからからmysqlにアクセスできない場合、
 
 - ### A. GCPのFirewall Ruleを確認してください。(TCP、3306許可など）

## CDN関連

### Q. CDNはどこの会社でも大丈夫でしょうか？

- ### A. はい、大丈夫です。

### Q. AWS Cloud FrontのCache設定を教えてください。

- ### A. [CloudFront Cache Setup](./aws-cdn-cache-setup.png)

### Q. OCTO-APIのCDN設定について教えてください。

- ### A. config.tmlで設定可能です。

  - [cdn].default : [cdn.apps]に設定されてない場合このURLを利用します。

  - [cdn.apps] : AppID = "CDN URL"を設定することでAppIdことにCDNを設定することができます。

## OCTO構成関連

### Q. OCTOの利用方法としてdev -> stg -> prdファイル同期できますか？

- ### A. 全ての環境で同じDBを使う前提で、dev->stg->prd ファイル同期可能です。

### Q. Env設定について教えてください。

- ### A. Envはその環境設定ではなく、Appに属しているVersionの環境を指定しますが、その環境はAndroidとかiOSなどを意味します。[EnvSetup](./setup-env.pdf)

### Q. 負荷に対するスペックを教えてください。

- ### A. [Maxリクエストに耐える構成例](./max-request-server.png)

### Q. OCTOの本番は冗長化されていますか？

- ### A. はい、Blue Green Deploy方式を利用しておりますので、ActiveとStandbyのサーバーが存在しています。

### Q. Blue Green Deployで切り替え方式を教えてください。

- ### A. url_mapにぶら下がっているBackend Serviceを切り替えています。

### Q. 本番のActive/ Standbyサーバーだけで運用できませんか

- ### A. 運用できると思いますが、開発の確認で苦労すると思いますので、オススメはしません。こちらオススメする構成は、SB環境は開発及び確認用で、PRD環境（Active, Standby）は本番用で利用することです。

### Q. GKEのインスタンスサイズについて教えてください。

- ### A. Instance ４(Node ４）、Pod ４です。（※インスタンスのサイズはプロジェクトの状況に合わせて設定してください）

### Q. GKEの設定について教えてください。

- ### A. [GKEフォルダーを参照してください。](/gke)

### Q. 本番構築で気をつけるべきことなど教えてください。

- ### A. 気をつける点は下記となります。
  - SlaveからのSnapshotの作成はおすすめしません。
    (テスト目的のMySQLをサーバーを作る時にSlaveのSnapshotから作成することで、レプリケーションエラーが発生する可能性があります。）
  - 開発と本番環境のbucket名は同一にして設定してください。
  - Master更新の時は、運用側と相談して行ってください。

### Q. GCSにファイルをアップロードしたい場合、どうすればいいですか？

- ### A. OCTO-CLIを利用してください。

### Q. Logはどこに出力されますか？

- ### A. octo-api, octo-adminは下記となります。
  - octo-api：stdoutに出力しています。gcpの場合stdoutに出力すると自動的にStack Drvier Loggingに転送されます。
  - octo-admin：/var/www/octo-admin/octo-admin.log

### Q. octo-adminのconfig.tmlにあるcookie_secretって何ですか？

- ### A. Cookieにsession情報を保存しますが、その情報のkeyになります。

### Q. octo-adminは必須存在ですか？

- ### A. 必須です。

### Q. admin 設定ファイルのoauth_googleは何ですか？ 

- ### A. Google認証の設定です。
oauth_googleの内容はgcpのAPI & Servicesのcredentialsでcredentialを作成すると生成されます。redirect_urlはlocalの場合 http://localhost:8081/google/oauth ですが、localhost:8081の代わりにドメインを設定してください。

### Q. admin 設定ファイルのgcp_projectのlocationについて

- ### A. GCS(Google Cloud Storage）のlocationになります。ASIAに設定している場合は、ASIAに設定してください。

