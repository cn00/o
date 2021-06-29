package main

import (
	"github.com/codegangsta/cli"
	"octo-cli/utils"
	"os"

	"octo-cli/commands/copy"
	"octo-cli/commands/list"
	"octo-cli/config"

	"github.com/BurntSushi/toml"
)

var Conf config.Config

func main() {

	app := cli.NewApp()
	app.Version = "v2.9" // "v" + float is required
	app.Name = "octo-cli"
	app.Usage = "octo uploader"
	app.EnableBashCompletion = true
	utils.App = app

	globalFlags := []cli.Flag{
		cli.StringFlag{
			Name:   "secret, s",
			Value:  "",
			Usage:  "octo application secret",
			EnvVar: "OCTO_APP_SECRET",
		},
		cli.StringFlag{
			Name:  "config, c",
			Value: "config.tml",
			Usage: "Specify the location of the config file.",
		},
		cli.BoolFlag{
			Name:  "cors, cr",
			Usage: "set cors for bucket ex) -cors or -cr",
		},
		cli.StringFlag{
			Name:  "corsStr, crs",
			Value: "",
			Usage: "set custom cors for bucket ex) -corsStr=\"{\"maxAge\":60, \"methods\": [\"GET\", \"POST\", \"PUT\", \"DELETE\", \"OPTIONS\"], \"origins\": [\"*\"], \"responseHeaders\":[\"X-Octo-Key\"]}\"",
		},
	}

	//MaxAge          time.Duration `json:"maxAge"`
	//Methods         []string      `json:"methods"`
	//Origins         []string      `json:"origins"`
	//ResponseHeaders []string      `json:"responseHeaders"`

	copyFlags := []cli.Flag{
		cli.IntFlag{
			Name:  "sourceVersionId, sv",
			Usage: "source version id",
		},
		cli.IntFlag{
			Name:  "destinationVersionId, dv",
			Usage: "destination version id",
		},
		cli.StringSliceFlag{
			Name:  "filenames, f",
			Usage: "target filenames",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug logging",
		},
	}

	checkFlags := []cli.Flag{
		cli.IntFlag{
			Name:  "versionId, v",
			Usage: "version id",
		},
		cli.StringFlag{
			Name:  "files, f",
			Usage: "files, comma delimited.",
		},
	}

	listFlags := []cli.Flag{
		cli.IntFlag{
			Name:  "versionId, v",
			Usage: "target version id",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "uploadOneAssetBundle",
			Aliases: []string{"uab"},
			Usage:   "upload one assetbundle which has diffrent crc.",
			Before: before,
			Action: func(c *cli.Context) {
				UploadAssetBundle(c.Int("version"), c.String("manifest"), c.StringSlice("tags"),
					c.Int("priority"), c.Bool("useOldTag"), c.String("buildNumber"), 
					c.Bool("cors"), c.String("corsStr"), c.String("specificManifest"), c.String("filter"), c)
			},
			Flags: append([]cli.Flag{
				cli.BoolFlag{
					Name:  "list, l",
					Usage: "manifest list module",
				},
				cli.StringFlag{
					Name:  "filter, f",
					Usage: "filter manifest",
				},
				cli.IntFlag{
					Name:  "version, v",
					Usage: "target asset versionId",
				},
				cli.StringSliceFlag{
					Name:  "tags, t",
					Usage: "add tags to assetbundle",
				},
				cli.IntFlag{
					Name:  "priority, p",
					Usage: "set priority to assetbundle",
				},
				cli.BoolFlag{
					Name:  "useOldTag, u",
					Usage: "add this flag if you don't want to update tags",
				},
				cli.StringFlag{
					Name:  "manifest, m",
					Usage: "Unity SingleManifestFile",
				},
				cli.StringFlag{
					Name:  "buildNumber, bn",
					Usage: "add build number to assetbundle",
				},
				cli.StringFlag{
					Name:  "specificManifest, sm",
					Usage: "Unity Specific ManifestFaile",
				},
			}, globalFlags...),
		},

		{
			Name:    "uploadAllAssetBundles",
			Aliases: []string{"ua"},
			Usage:   "upload all assetbundle which has diffrent crc.",
			Before: before,
			Action: func(c *cli.Context) {
				MultiUploadAssetBundle(c.Int("version"), c.String("manifest"), c.StringSlice("tags"), c.Int("priority"), c.Bool("useOldTag"), c.String("buildNumber"), c.Bool("cors"), c.String("corsStr"), c.String("specificManifest"))
			},

			Flags: append([]cli.Flag{
				cli.IntFlag{
					Name:  "version, v",
					Usage: "target asset versionId",
				},
				cli.StringSliceFlag{
					Name:  "tags, t",
					Usage: "add tags to assetbundle",
				},
				cli.IntFlag{
					Name:  "priority, p",
					Usage: "set priority to assetbundle",
				},
				cli.BoolFlag{
					Name:  "useOldTag, u",
					Usage: "add this flag if you don't want to update tags",
				},
				cli.StringFlag{
					Name:  "manifest, m",
					Usage: "Unity SingleManifestFile",
				},
				cli.StringFlag{
					Name:  "buildNumber, bn",
					Usage: "add build number to assetbundle",
				},
				cli.StringFlag{
					Name:  "specificManifest, sm",
					Usage: "Unity Specific ManifestFaile",
				},
			}, globalFlags...),
		},
		{
			Name:    "uploadAllResources",
			Aliases: []string{"uar"},
			Usage:   "upload all resources which has diffrent md5 in your specific directory.",
			Before: before,
			Action: func(c *cli.Context) {
				MultiUploadResources(c.Int("version"), c.String("basedir"), c.StringSlice("tags"),
					c.Int("priority"), c.Bool("useOldTag"), c.String("buildNumber"),
					c.String("corsStr"), c.Bool("cors"), c.Bool("recursion"), c.String("specificFilePath"))
			},

			Flags: append([]cli.Flag{
				cli.IntFlag{
					Name:  "version, v",
					Usage: "target asset versionId",
				},
				cli.StringSliceFlag{
					Name:  "tags, t",
					Usage: "add tags to assetbundle",
				},
				cli.IntFlag{
					Name:  "priority, p",
					Usage: "set priority to assetbundle",
				},
				cli.BoolFlag{
					Name:  "useOldTag, u",
					Usage: "add this flag if you don't want to update tags",
				},
				cli.StringFlag{
					Name:  "basedir, b",
					Usage: "base directory",
				},
				cli.BoolFlag{
					Name:  "recursion, r",
					Usage: "recursion subdir on dir ex) -r or --recursion",
				},
				cli.StringFlag{
					Name:  "buildNumber, bn",
					Usage: "add build number to assetbundle",
				},
				cli.StringFlag{
					Name:  "specificFilePath, fp",
					Usage: "set filepath for specific file",
				},
			}, globalFlags...),
		},
		{
			Name:    "addTagToAssetBundles",
			Aliases: []string{"ta"},
			Usage:   "add tag to assetBundles.",

			Flags: append([]cli.Flag{
				cli.IntFlag{
					Name:  "version, v",
					Usage: "target asset versionId",
				},
				cli.StringSliceFlag{
					Name:  "files, f",
					Usage: "target AssetBundle name",
				},
				cli.StringSliceFlag{
					Name:  "tags, t",
					Usage: "add tags to assetbundle",
				},
			}, globalFlags...),
			Before: before,
			Action: func(c *cli.Context) {
				addTagToAssetBundle(c.Int("version"), c.StringSlice("files"), c.StringSlice("tags"))
			},
		},
		{
			Name:    "addTagToResources",
			Aliases: []string{"tr"},
			Usage:   "add tag to resources.",

			Flags: append([]cli.Flag{
				cli.IntFlag{
					Name:  "version, v",
					Usage: "target asset versionId",
				},
				cli.StringSliceFlag{
					Name:  "files, f",
					Usage: "target Resource name",
				},
				cli.StringSliceFlag{
					Name:  "tags, t",
					Usage: "add tags to resource",
				},
			}, globalFlags...),
			Before: before,
			Action: func(c *cli.Context) {
				addTagToResource(c.Int("version"), c.StringSlice("files"), c.StringSlice("tags"))
			},
		},
		{
			Name:    "removeTagToAssetBundles",
			Aliases: []string{"rta"},
			Usage:   "remove tag to assetBundles.",

			Flags: append([]cli.Flag{
				cli.IntFlag{
					Name:  "version, v",
					Usage: "target asset versionId",
				},
				cli.StringSliceFlag{
					Name:  "files, f",
					Usage: "target AssetBundle name",
				},
				cli.StringSliceFlag{
					Name:  "tags, t",
					Usage: "remove tags to assetbundle",
				},
			}, globalFlags...),
			Before: before,
			Action: func(c *cli.Context) {
				removeTagToAssetBundle(c.Int("version"), c.StringSlice("files"), c.StringSlice("tags"))
			},
		},
		{
			Name:    "removeTagToResources",
			Aliases: []string{"rtr"},
			Usage:   "remove tag to resources.",

			Flags: append([]cli.Flag{
				cli.IntFlag{
					Name:  "version, v",
					Usage: "target asset versionId",
				},
				cli.StringSliceFlag{
					Name:  "files, f",
					Usage: "target Resource name",
				},
				cli.StringSliceFlag{
					Name:  "tags, t",
					Usage: "remove tags to resource",
				},
			}, globalFlags...),
			Before: before,
			Action: func(c *cli.Context) {
				removeTagToResource(c.Int("version"), c.StringSlice("files"), c.StringSlice("tags"))
			},
		},
		{
			Name:    "deleteAssetBundles",
			Aliases: []string{"da"},
			Usage:   "delete assetBundles.",

			Flags: append([]cli.Flag{
				cli.IntFlag{
					Name:  "version, v",
					Usage: "target asset versionId",
				},
				cli.StringSliceFlag{
					Name:  "files, f",
					Usage: "target AssetBundle name",
				},
			}, globalFlags...),
			Before: before,
			Action: func(c *cli.Context) {
				deleteAssetBundle(c.Int("version"), c.StringSlice("files"))
			},
		},
		{
			Name:    "deleteResources",
			Aliases: []string{"dr"},
			Usage:   "delete resources.",

			Flags: append([]cli.Flag{
				cli.IntFlag{
					Name:  "version, v",
					Usage: "target asset versionId",
				},
				cli.StringSliceFlag{
					Name:  "files, f",
					Usage: "target Resource name",
				},
			}, globalFlags...),
			Before: before,
			Action: func(c *cli.Context) {
				deleteResource(c.Int("version"), c.StringSlice("files"))
			},
		},
		{
			Name:    "copyAssetBundles",
			Aliases: []string{"ca"},
			Usage:   "copy assetBundles.",

			Flags:  append(copyFlags, globalFlags...),
			Before: before,
			Action: func(c *cli.Context) {
				copy.CopyAssetBundle(copy.CopyOptions{
					Config:               Conf,
					SourceVersionId:      c.Int("sourceVersionId"),
					DestinationVersionId: c.Int("destinationVersionId"),
					Filenames:            c.StringSlice("filenames"),
					Debug:                c.Bool("debug"),
				})
			},
		},
		{
			Name:    "copyResources",
			Aliases: []string{"cr"},
			Usage:   "copy resources.",

			Flags:  append(copyFlags, globalFlags...),
			Before: before,
			Action: func(c *cli.Context) {
				copy.CopyResource(copy.CopyOptions{
					Config:               Conf,
					SourceVersionId:      c.Int("sourceVersionId"),
					DestinationVersionId: c.Int("destinationVersionId"),
					Filenames:            c.StringSlice("filenames"),
					Debug:                c.Bool("debug"),
				})
			},
		},
		{
			Name:    "diffSync",
			Aliases: []string{"ds"},
			Usage:   "diff sync assetBundles and resources.",

			Flags: append([]cli.Flag{
				cli.IntFlag{
					Name:  "version, v",
					Usage: "target asset versionId",
				},
				cli.IntFlag{
					Name:  "sourceApp, sa",
					Usage: "source appId",
				},
				cli.IntFlag{
					Name:  "sourceVersion, sv",
					Usage: "source versionId",
				},
				cli.IntFlag{
					Name:  "revision, r",
					Usage: "target revision",
				},
			}, globalFlags...),
			Before: before,
			Action: func(c *cli.Context) {
				diffSync(c.Int("version"), c.Int("sourceApp"), c.Int("sourceVersion"), c.Int("revision"))
			},
		},
		{
			Name:    "diffSyncLatest",
			Aliases: []string{"dsl"},
			Usage:   "diff sync latest revision assetBundles and resources.",

			Flags: append([]cli.Flag{
				cli.IntFlag{
					Name:  "version, v",
					Usage: "target asset versionId",
				},
				cli.IntFlag{
					Name:  "sourceApp, sa",
					Usage: "source appId",
				},
				cli.IntFlag{
					Name:  "sourceVersion, sv",
					Usage: "source versionId",
				},
			}, globalFlags...),
			Before: before,
			Action: func(c *cli.Context) {
				diffSyncLatest(c.Int("version"), c.Int("sourceApp"), c.Int("sourceVersion"))
			},
		},
		{
			Name:    "checkAssetBundlesExistence",
			Aliases: []string{"cae"},
			Usage:   "check assetBundles existence.",

			Flags:  append(checkFlags, globalFlags...),
			Before: before,
			Action: func(c *cli.Context) {
				checkAssetBundleExistence(c.Int("versionId"), c.String("files"))
			},
		},
		{
			Name:    "checkResouceExistence",
			Aliases: []string{"cre"},
			Usage:   "check resources existence.",

			Flags:  append(checkFlags, globalFlags...),
			Before: before,
			Action: func(c *cli.Context) {
				checkResourceExistence(c.Int("versionId"), c.String("files"))
			},
		},
		{
			Name:    "listAssetBundles",
			Aliases: []string{"la"},
			Usage:   "list assetbundles.",

			Flags:  append(listFlags, globalFlags...),
			Before: before,
			Action: func(c *cli.Context) {
				list.ListAssetbundles(list.ListOptions{
					Config:    Conf,
					VersionId: c.Int("versionId"),
				})
			},
		},
		{
			Name:    "listResources",
			Aliases: []string{"lr"},
			Usage:   "list resources.",

			Flags:  append(listFlags, globalFlags...),
			Before: before,
			Action: func(c *cli.Context) {
				list.ListResources(list.ListOptions{
					Config:    Conf,
					VersionId: c.Int("versionId"),
				})
			},
		},
	}

	app.Run(os.Args)
}

type StdOutTest struct {
	Config string
	Secret string
}

func before(c *cli.Context) error {
	decodeTomle(c.String("config"))
	utils.AppSecret = Conf.App.Secret // c.String("secret")
	return nil
}

func decodeTomle(filePath string) {
	_, err := toml.DecodeFile(filePath, &Conf)
	if err != nil {
		panic(err)
	}
}
