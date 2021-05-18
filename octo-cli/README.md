# octo-cli
<img height="200" src="https://storage.googleapis.com/octo-static/OctoLogo.png">

Cli tool for octo.

## Precondition

OCTO use Google Cloud Storage. So you need to install Google Cloud SDK.

install gcloud sdk from
https://cloud.google.com/sdk/

you need your ACCOUNT and json file and projectId from administrator

```
$ export GOOGLE_APPLICATION_CREDENTIALS=[your.json]
$ gcloud auth activate-service-account [ACCOUNT] --key-file [your.json] --project [projectId]
```

you need to change id of App and baseUrl of API on config.tml.
* <b>Please contact to account manager of octo that about AppId and BaseUrl.</b>   
```
[Api]
base_url = "set octo api url"

[App]
id = you AppId

```

## Usage

`octo-cli` use `global options`, `command`, `command options` and `arguments`

```
$ bin/octo-cli [global options] command [command options] [arguments...]
```

there have 7 command type

```
uploadAllAssetBundles, ua	upload all assetbundle which has diffrent crc.
uploadAllResources, uar	upload all resources which has diffrent md5 in your specific directory.
addTagToAssetBundles, ta	add tag to assetBundles.
addTagToResources, tr	add tag to resources.
removeTagFromAssetBundles, rta    remove tag from assetBundles.
removeTagFromResources, rtr   remove tag from assetBundles.
deleteAssetBundles, da	delete assetBundles.
deleteResources, dr		delete resources.
diffSync, ds			diff sync assetBundles and resources.
diffSyncLatest, dsl		diff sync latest revision assetBundles and resources.
checkAssetBundlesExistence, cae	check assetBundles existence.
checkResouceExistence, cre		check resources existence.
listAssetBundles, la  get assetbundle list.
listResources, lr get resources list.
help, h			Shows a list of commands or help for one command
```

and 5 global options

```
--secret, -s "xxxxx"		octo application secret [$OCTO_APP_SECRET]
--config, -c "config.tml"	Specify the location of the config file.
--cros, -cr		set cors for bucket
--corsStr, -crs "{\"maxAge\":60, \"methods\": [\"GET\", \"POST\", \"PUT\", \"DELETE\", \"OPTIONS\"], \"origins\": [\"*\"], \"responseHeaders\":[\"X-Octo-Key\"]}"
" set custom cors. need to set cors  
--help, -h			show help
--generate-bash-completion
--version, -v		print the version
--recursion, -r     recursion subdir on dir
```

secret can set by enciroment variables like

```
$ export OCTO_APP_SECRET=[secret]
```

you can move config.tml to anywhere when use `--config` option.

any commands have `--help` command option.

### Upload

You can upload AssetBundles and Resources.

Here is AssetBundles example.

```
$ bin/octo-cli ua -v [version id] -m [manifest file]
```

and you can add tags.

```
$ bin/octo-cli ua -v [version id] -m [manifest file] -t [tagname1] -t [tagname2]
```

and you can add build number
```
$ bin/octo-cli ua -v [version id] -m [manifest file] -bn [build number]
```

and you can specific assetbundle 
```
$ bin/octo-cli ua -v [version id] -m [manifest file] -sm [specific assetbundle manifest that exclude ext name]
ex) bin/octo-cli ua -v 8500 -m /test/v1.manifest -sm test1
```

Here is Resources example
```
$ bin/octo-cli uar -v [version id] -b [resource file directory]
```

and you can add tags.
```
$ bin/octo-cli uar -v [version id] -b [resource file directory] -t [tagname1] -t [tagname2]
```

and you can add build number
```
$ bin/octo-cli uar -v [version id] -b [resource file directory] -bn [build number]
```

and you can specific file 
```
$ bin/octo-cli uar -v [version id] -fp [specific filepath]
ex) bin/octo-cli uar -v 9320 -fp /test/test.txt
```

### Add Tag

You can add tags after upload.

```
$ bin/octo-cli ta -v [version id] -f [target assetbundle name1] -f [target assetbundle name2] -t [tagname1] -t [tagname2]
```

### Remove Tag

You can remove tags after upload

```
$ bin/octo-cli rta -v [version id] -f [target assetbundle name1] -f [target assetbundle name2] -t [tagname1] -t [tagname2]
```

Remove all tags if there is no target tag 

```
$ bin/octo-cli rta -v [version id] -f [target assetbundle name1] -f [target assetbundle name2]
```


### Delete

Delete is just a logical delete.

```
$ bin/octo-cli da -v [version id] -f [target assetbundle name1] -f [target assetbundle name2]
```

### Diff Sync

You can syncronize your AssetBundles and resources from other project or version.

```
bin/octo-cli dsl -v [version id] -sa [source application id] -sv [source version id]
```

if you need to choose specific revision. use `ds` command

```
bin/octo-cli ds -v [version id] -r [revision id] -sa [source application id] -sv [source version id]
```

You can not use normal upload in combination .

### Check Existence

You can check your AssetBundles and resources existence.

```
bin/octo-cli cae -v [version id] -f [target assetbundle name, comma delimited]
```

### cors set example

#### AssetBundle 
```
$ ./octo-cli ua -v [version]  -m [manifest file] -cr
```

#### Resources
```
$ ./octo-cli uar -v [version]  -b [base-path] --cors
```

### custom cors set example

#### AssetBundle 
```
$ ./octo-cli ua -v [version] -m [manifest file] --cors --corsStr="{\"maxAge\":60, \"methods\": [\"GET\", \"PUT\", \"DELETE\", \"OPTIONS\"], \"origins\": [\"*\"], \"responseHeaders\":[\"X-Octo-Key\"]}"
```

#### Resources
```
$ ./octo-cli uar -v [version]  -b [base-path] --cors
 --corsStr="{\"maxAge\":60, \"methods\": [\"GET\", \"POST\", \"PUT\", \"DELETE\", \"OPTIONS\"], \"origins\": [\"*\"], \"responseHeaders\":[\"X-Octo-Key\"]}"
```


### recursion subdir resources file updload example(Only Use upload resources)
```
bin/octo-cli uar -v [version] -b [base-path] -r

or

bin/octo-cli uar -v [version] -b [base-path] -recursion
```

### assetbundle list sample
```
$ ./octo-cli la -v [version]
```

### resources list sample
```
$ ./octo-cli lr -v [verison]
```
