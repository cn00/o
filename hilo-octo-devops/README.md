# octo-devops

## GCPでテストクラスタ環境を構築する手順

### この手順はGCP(Google Cloud Platform)上で作業することを前提とします。

### 1. GCP上クラスタ作成

- 完全新規の場合、octo-gcp-cluster.shを利用してください。(完全新規でなくても利用できます。)
- octo-gcp-cluster.shを利用する前にgcloud sdkのインストールとgcloud認証は必須です。
  - gcloud sdk : <https://cloud.google.com/sdk/?hl=ja>
  - 認証：<https://cloud.google.com/sdk/gcloud/reference/auth/login>
- octo-gcp-cluster.shを利用することでGCP上、下記のことが作成されます。
  - ネットワーク新規作成（以下の作業はすべてこのネットワークを利用します。）
  - GKE cluster作成（ZONEにNODEを二つ、追加ZONEにNODEを二つ、合計４つのVM Instance 生成します。）
  - MySQL用VM Instance作成
  - Firewall作成(外部から接続するため tcp22 portなど開ける)
  - HealthCheck作成（30420でチェックします。）
  - HTTP LB作成（HTTPS LBが必要であれば、このLBを参照して別当作成してください）
    - 作成したgke clusterのbackendサービス作成（portは30420を利用）
    - Frontend 作成（静的IPは自動生成）
- 利用方法。（DBマシンタイプはn1-highmem-2を推奨します。）

  ```~bash
    ./octo-gcp-cluster.sh ネットワーク名 インスタンス名のPrefix マシンタイプ　DBマシンタイプ ゾーン 追加ゾーン 地域
  ```

- 利用例

  ```~bash
    ./octo-gcp-cluster.sh temp-octo-n test-octo-app-001 n1-standard-1 n1-highmem-2 asia-northeast1-a asia-northeast1-c asia-northeast1
  ```

- クラスタ作成後、クラスタ情報取得

  ```~bash
  gcloud container clusters describe クラスタ名 --zone ZONE名
  gcloud container clusters get-credentials クラスタ名 --zone ZONE名
  ```

### 2. MySQL作業

- Ansible Task :
  https://github.com/QualiArts/hilo-octo-devops/tree/master/ansible/db

  - Ansible Taskを実行することで下記、a、bの作業が自動で行われます。

    a. VMInstance上、MySQLインストール作業
      ```~bash
            yum -y install http://dev.mysql.com/get/mysql-community-release-el6-5.noarch.rpm`
            yum info mysql-community-server
            yum -y install mysql-community-server
      ```

    b. DB作成
    - CREATE DATABASE octo DEFAULT CHARACTER SET utf8;

    c. ユーザー作成とAPPからの接続許可
    - GRANT ALL PRIVILEGES ON octo.* TO octo@'xx.%' IDENTIFIED BY 'hilo';

### 3. OCTO用config.tmlファイル作成

- addrsはMySQL VM Instanceの内部IPを指定してください。
- cache_appsにはAppIDをしていしてください。(コンマを使うことで複数指定できます。)
- cdn.appsはCDNを利用するAppIDとCDN URLを設定してください。

    ```~tml
    [api]
    port = 8080
    read_only = false
    minimum_cli_version = 0.0

    [database.master]
    addrs = "xxx.xxx.xxx.xxx:3306"
    dbname = "octo"
    user = "octo"
    password = "hilo"

    [database.slave]
    addrs = "xxx.xxx.xxx.xxx:3306"
    dbname = "octo"
    user = "octo"
    password = "hilo"

    [metrics]
    port = 2000

    [cache_apps]
    list_api = [ Set Your AppId ]

    [cdn]
    default = "https://storage.googleapis.com"

    [cdn.apps]
    Set Your AppId = "Set Your CDN URL"
    # ex)
    # 1 = "https://storage.googleapis.com"
    ```

### 4. Dockerイメージ作成

- OCTOのdockerImageを作成しGoogleContainerRegistryにアップロードします。
- この工程は通常はJenkinsで行うのが良いです
- octo-protoのレポジトリにssh公開鍵を登録してください。
- hilo-octo-serverをクローン
- octo-apiビルド

  ```~bash
    cd hilo-octo-server
    echo "Go build start"
    go get -u github.com/constabulary/gb/...
    eval $(ssh-agent)
    ssh-add ~/.ssh/octo-proto
    gb vendor restore
    ssh-agent -k
    gb build octo-api
  ```

- docker build

  ```~bash
    GOOS=linux GOARCH=386 docker build -t asia.gcr.io/${PROJECT_ID}/octo-api:${VERSION} .
  ```

- push to docker

  ```~bash
    echo "Push to Docker repo"
    gcloud docker -- push asia.gcr.io/${PROJECT_ID}/octo-api:${VERSION}
  ```

### 5. Kubernetes周り作業

yamlファイルは[gke](gke/)フォルダを参照してください。

#### a. Config Map作成(２で作成したconfig.tmlを使います。）

  ```~bash
    kubectl create configmap octo-config --from-file=config-api.tml=config.tml -o yaml
  ```

#### b. deployment.ymlを作成(下記は作成例)

  ```~yaml
    apiVersion: extensions/v1beta1
    kind: Deployment
    metadata:
    name: temp-octo-api-deployment
    spec:
       replicas: 4
    selector:
       matchLabels:
         name: temp-api
     template:
    metadata:
      labels:
        name: temp-api
        deploy-date: ${DEPLOY_DATE}
    spec:
      containers:
      - name: temp-api
        image: asia.gcr.io/hilo-1047/octo-api:${VERSION}
        resources:
          requests:
            cpu: 200m
        volumeMounts:
        - mountPath: /octo/etc
          name: config
        ports:
        - containerPort: 8080
          protocol: TCP
      volumes:
        - name: config
          configMap:
            name: octo-config
    revisionHistoryLimit: 100
  ```

#### c. appのdeploy

```~bash
kubectl apply -f deployment.yml
```

#### d. service.yml作成

```~yaml
    apiVersion: v1
    kind: Service
    metadata:
       name: temp-octo-api
       labels:
         name: api
    spec:
     ports:
     - port: 8080
       targetPort: 8080
       protocol: TCP
       nodePort: 30420
     type: NodePort
       selector:
         name: temp-api
```

#### e.サービス追加

```~bash
kubectl create -f service.yml
```

#### f.サービス追加後、EndPoint確認

```~bash
kubectl describe service temp-octo-api
```

#### g. 実行結果の例（Endpointsに何も書いてなければ、service.ymlでselectorのnameを確認する必要がある。

```~yml
    Name:            octo-stg-api
    Namespace:        default
    Labels:            name=api
    Annotations:        <none>
    Selector:        name=stg-api
    Type:            NodePort
    IP:            10.191.246.28
    Port:            <unset>    8080/TCP
    NodePort:        <unset>    30420/TCP
    Endpoints:        10.188.2.35:8080,10.188.2.36:8080,10.188.3.49:8080 + 1 more...
    Session Affinity:    None
    Events:            <none>
  ```

### 6. 疎通確認

- statusのokの確認

```~bash
curl http://externalip:30420/status
```

## OCTO管理ツール設定方法

### 1. 管理ツール用 Instance作成

- Instanceはlinux CentOS 5.7で作成してください。

### 2. 作成したInstanceでocto-adminをbuild

- gitからocto-serverをcloneしてビルド

```~bash
git@github.com:QualiArts/hilo-octo-server.git
cd hilo-octo-server
go get -u github.com/constabulary/gb/...
gb build octo-admin
```

### 3. octo-admin.service作成

- /etc/systemd/system/octo-admin.serviceの作成例

```~service
[Unit]
Description=octo-admin

[Service]
User=octo-admin
Group=octo-admin
WorkingDirectory=/var/www/octo-admin
ExecStart=/var/www/octo-admin/bin/octo-admin -conf config.tml
Restart=always

[Install]
WantedBy=multi-user.target
```

### 4. octo-admin deploy

- gitからocto-devopsをclone、必要なファイルコピーしocto-adminスタート

```~bash
cp hilo-octo-devops/octo/admin/config-prd.tml /var/www/octo-admin/config.tml
cp ../octo-server/bin/octo-admin /var/www/octo-admin/bin/
cp -r ../octo-server/static $host:/var/www/octo-admin/
cp -r ../octo-server/templates $host:/var/www/octo-admin/
sudo systemctl start octo-admin

```

### 5. Admin追加

octo dbにアクセスして下記のDDLを実行してください。
ご自身で使用するGoogleのemailアドレスを設定してください。
```
insert into users(user_id, email, auth_type) values('TODO-Set Googleのemailアドレス', 'TODO-Set Googleのemailアドレス', 3);
insert into user_apps(app_id, email, role_type) values(0, 'TODO-Set Googleのemailアドレス', 2);
```

### 6. octo-adminでApp追加

- [新しいApp追加方法](./octo-new-app.pdf)

## MySQL レプリケーション設定方法

### ①スレーブサーバー作成

マスターサーバーと同じmysql versionで作成します。
GCPの場合masterのsnapshotをとり、そのsnapshotから作成した方がいいと思います。

### ②スレーブサーバーのmy.cnfの編集

my.cnfのserver_idをMasterのserver_idと異なるIDを設定します。

### ③Slaveサーバーの起動

※Slaveを２個作る場合は①~③を繰り返します。server_idはユニークに設定してください。

### ④マスター状にレプリケーション用のユーザーを作成する

```~sql
mysql> GRANT REPLICATION SLAVE ON *.* TO 'ユーザー名'@'スレーブのホスト名' IDENTIFYED BY 'パスワード'；
```

※Slave数分実行します。

### ⑤マスターのデータをダンプする

※この作業ですが、MasterのsnapshotからSlaveを作成してからmasterへのデータ更新がない場合は行う必要はありません。

```~bash
mysqldump -uroot -p --all-databases --master-data=2 --single-transaction --flush-logs > dumpfile.sql
```

⑥ マスターの状態を確認して　FileとPositionを取っておきます。

```~sql
mysql> show master status\G
*************************** 1. row ***************************
             File: mysql-bin.000010
         Position: 100
     Binlog_Do_DB:
 Binlog_Ignore_DB:
Executed_Gtid_Set:
1 row in set (0.00 sec)
```

### ⑦マスターで取得したデータをスレーブへリストアする

※この作業ですが、MasterのsnapshotからSlaveを作成してからmasterへのデータ更新がない場合は行う必要はありません。

```~bash
mysql -uユーザー名 -p < dumpfile.sql
```

### ⑧スレーブ上レプリケーションの設定を行う

```~sql
CHANGE MASTER TO MASTER_HOST='マスターのホスト名またはIPアドレス', MASTER_PORT=3306, MASTER_USER='ユーザー名', MASTER_PASSWORD='パスワード', MASTER_LOG_FILE='⑥で確認したFILE名', MASTER_LOG_POS=⑥で確認したPositionの値;
```

### ⑨レプリケーションを開始する

```~sql
mysql> START SLAVE;
```

### ⑩レプリケーション状態を確認する

```~sql
mysql> SHOW LAVE STATUS\G
```

※⑦〜⑩はSlave数分実行します。

## 現状本番スペック

### octo-asia-app-001

- master zone:asia-east1-a
- node zones:asia-east1-a, asia-east1-c
- num-nodes:4
- machine-type:n1-standard-4
- network:hilo

### octo-asia-app-002

- master zone:asia-east1-a
- node zones:asia-east1-a, asia-east1-c
- num-nodes:4
- machine-type:n1-standard-2
- network:hilo
