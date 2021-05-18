# octo-proto
```
               __                                        __          
  ____   _____/  |_  ____           _____________  _____/  |_  ____  
 /  _ \_/ ___\   __\/  _ \   ______ \____ \_  __ \/  _ \   __\/  _ \ 
(  <_> )  \___|  | (  <_> ) /_____/ |  |_> >  | \(  <_> )  | (  <_> )
 \____/ \___  >__|  \____/          |   __/|__|   \____/|__|  \____/ 
            \/                      |__|                                             
```
octoのProtocolBuffersを管理するプロジェクト
## ProtocolBuffers
ProtocolBuffersはデータのシリアライズツールです。<br>
https://developers.google.com/protocol-buffers/

macにインストールする場合はhomebrew等で入れると簡単です

```
brew install protobuf
```

### golang

protoファイル（定義ファイル）をコンパイルする手順は下記の通りです。

まず、プラグインをインストールします。<br>
この時点でGOPATHの設定などはしておいて下さい。

```
$ go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
```

あとは対象のファイル名を変えて、次のコマンドを実行してください。

```
$ protoc --go_out=. data/data.proto
```

元ファイルと同じ場所にgoのファイルが生成されます。

### CSharp

c#は一度DescriptorSetに変換してからprotobut-netで変換します。<br>
protobuf-netは`tools`ディレクトリに入っています。

対象のファイル名を変えて、次のコマンドを実行してください。

```
$ protoc --descriptor_set_out=data.desc data/data.proto

$ mono tool/ProtoGen/protogen.exe -i:data.desc -o:data.cs
```
指定した位置にdescファイルとcsファイルが生成されます。