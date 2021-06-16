package main

import (
	"io/ioutil"
	"log"
	"strings"

	"gopkg.in/yaml.v2"
	"reflect"
	"strconv"
)

type SingleManifest struct {
	ManifestFileVersion int
	AssetBundleManifest map[string]AssetBundleInfo
}
type AssetBundleInfo struct {
	Dependencies []string
	CRC          uint32
}

type Manifest struct {
	ManifestFileVersion int
	CRC                 uint32
	Assets              []string
	Dependencies        []string
}

func GetDependencyList(manifestFile string, target string) []string {
	target = strings.Split(target, ".")[0]
	buf, err := ioutil.ReadFile(manifestFile)
	if err != nil {
		log.Fatal(manifestFile + " can not open." + err.Error())
		panic(err)
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(buf, &m)
	if err != nil {
		panic(err)
	}

	log.Printf("%s", m["AssetBundleManifest"].(map[interface{}]interface{})["AssetBundleInfos"].(map[interface{}]interface{})["Info_0"].(map[interface{}]interface{})["Name"])

	var dependencyList = []string{}
	for _, asset := range m["AssetBundleManifest"].(map[interface{}]interface{})["AssetBundleInfos"].(map[interface{}]interface{}) {
		name := asset.(map[interface{}]interface{})["Name"]
		if name == target {
			for _, dependency := range asset.(map[interface{}]interface{})["Dependencies"].(map[interface{}]interface{}) {
				log.Println(dependency.(string))
				dependencyList = append(dependencyList, dependency.(string))
			}
		}
	}
	return dependencyList
}

func DecodeBundleManifest(manifestFile string) *SingleManifest {
	manifest1 := DecodeManifest(manifestFile)
	
	//m := make(map[interface{}]interface{})
	//err := yaml.Unmarshal(buf, &m)
	//if err != nil {
	//	panic(err)
	//}

	manifest := new(SingleManifest)
	manifest.ManifestFileVersion = manifest1.ManifestFileVersion
	var assetBundleMap = map[string]AssetBundleInfo{}
	
	assetName := manifestFile[strings.LastIndex(manifestFile,"/")+1:strings.LastIndex(manifestFile,".")]
	//crc := m["CRC"].(uint32)
	//var dependencyList = []string{}
	//for _, dependency := range m["Dependencies"].([]interface{}) {
	//	depStr := typeCheck(dependency)
	//	depStr = depStr[strings.LastIndex(depStr,"/")+1:]
	//	dependencyList = append(dependencyList, depStr)
	//}
	assetBundleMap[assetName] = AssetBundleInfo{Dependencies: manifest1.Dependencies, CRC: manifest1.CRC}
	manifest.AssetBundleManifest = assetBundleMap
	
	if len(manifest1.Dependencies) > 0 {
		return manifest
	} else {
		return nil
	}
}

func DecodeSingleManifest(manifestFile string) SingleManifest {
	buf, err := ioutil.ReadFile(manifestFile)
	if err != nil {
		log.Fatal(manifestFile + " can not open." + err.Error())
		panic(err)
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(buf, &m)
	if err != nil {
		panic(err)
	}

	manifest := SingleManifest{ManifestFileVersion: m["ManifestFileVersion"].(int)}
	var assetBundleMap = map[string]AssetBundleInfo{}
	for _, asset := range m["AssetBundleManifest"].(map[interface{}]interface{})["AssetBundleInfos"].(map[interface{}]interface{}) {
		name := asset.(map[interface{}]interface{})["Name"]
		strName := typeCheck(name)
		var dependencyList = []string{}
		for _, dependency := range asset.(map[interface{}]interface{})["Dependencies"].(map[interface{}]interface{}) {
			depStr := typeCheck(dependency)
			dependencyList = append(dependencyList, depStr)
		}
		assetBundleMap[strName] = AssetBundleInfo{Dependencies: dependencyList}
	}
	manifest.AssetBundleManifest = assetBundleMap
	return manifest
}

func typeCheck(name interface{}) string {
	var strName string
	if reflect.TypeOf(name).Kind() == reflect.Int {
		strName = strconv.Itoa(name.(int))
	} else if reflect.TypeOf(name).Kind() == reflect.Float64 {
		strName = strconv.FormatFloat(name.(float64), 'E', -1, 64)
	} else if reflect.TypeOf(name).Kind() == reflect.String {
		strName = name.(string)
	}
	return strName
}

func DecodeManifest(manifestFile string) Manifest {
	buf, err := ioutil.ReadFile(manifestFile)
	if err != nil {
		log.Fatal(manifestFile + " can not open." + err.Error())
		panic(err)
	}

	m := make(map[interface{}]interface{})
	err = yaml.Unmarshal(buf, &m)
	if err != nil {
		panic(err)
	}

	manifest := Manifest{
		ManifestFileVersion: m["ManifestFileVersion"].(int),
		CRC:                 uint32(m["CRC"].(int)),
	}

	var assets = make([]string, 0)
	ma, ok := m["Assets"]; if ok {
		for _, asset := range ma.([]interface{}) {
			assets = append(assets, asset.(string))
		}
	}else{
		log.Println("Assets_nil", manifestFile)
	}
	manifest.Assets = assets

	var dependencyList = []string{}
	md, ok := m["Dependencies"]
	if ok {
		for _, dependency := range md.([]interface{}) {
			depStr := typeCheck(dependency)
			depStr = depStr[strings.LastIndex(depStr,"/")+1:]
			//	depStr = depStr[strings.LastIndex(depStr,"/")+1:]
			dependencyList = append(dependencyList, depStr)
		}
	}else{
		log.Println("Dependencies_nil", manifestFile)
	}
	manifest.Dependencies = dependencyList

	log.Println("Dependencies_len", len(dependencyList), "Assets_len", len(assets), manifestFile, strings.Join(dependencyList, "|"))
	return manifest
}
