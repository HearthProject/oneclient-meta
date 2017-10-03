package main

import (
	"io/ioutil"
	"encoding/json"
	"os"
	"github.com/HearthProject/oneclient-meta/utils"
	"fmt"
	"github.com/deckarep/golang-set"
)
const MAIN = "oneclient/"
const MC = "upstream/minecraft/"
const VERSIONS = MC + "versions/"
const ASSETS = MC + "assets/"

const NETMC = "meta/net.minecraft/"
const MANIFEST_URL = "https://launchermeta.mojang.com/mc/game/version_manifest.json"
const MANIFEST_PATH = MAIN + MC + "version_manifest.json"

func minecraftMeta() {
	os.RemoveAll(MAIN + NETMC)
	os.RemoveAll(MAIN + VERSIONS)
	os.RemoveAll(MAIN + ASSETS)
	utils.MakeDir(MAIN + NETMC)
	utils.MakeDir(MAIN + MC)
	utils.MakeDir(MAIN + VERSIONS)
	utils.MakeDir(MAIN + ASSETS)
	localManifest, err := parseVersionManifest()
	localVersions := mapset.NewSet()
	if err == nil {
		localVersions = utils.StringSet(localManifest.versionKeys())
	}
	err = utils.DownloadFile(MANIFEST_PATH, MANIFEST_URL)
	if err != nil {
		panic(err)
	}
	remoteManifest, err := parseVersionManifest()
	if err != nil {
		panic(err)
	}
	remoteVersions := utils.StringSet(remoteManifest.versionKeys())
	newIDs := remoteVersions.Difference(localVersions)
	checkIDs := remoteVersions.Difference(newIDs)
	updateIDs := newIDs
	for i, k := range checkIDs.ToSlice() {
		remoteVersion := remoteManifest.Versions[i]
		localVersion := localManifest.Versions[i]
		if remoteVersion.Time > localVersion.Time {
			updateIDs.Add(k)
		}
	}

	assets := []AssetIndex{}
	for i := range updateIDs.ToSlice() {
		version := remoteManifest.Versions[i]
		fmt.Printf("Updating %v to timestamp %v \n", version, version.Release)
		asset := GetVersionFile(version)
		assets = append(assets, asset)
	}

	for _, v := range assets {
		fmt.Println("assets", v.Url, v.Id)
		GetAsset(v)
	}
}

func GetAsset(asset AssetIndex) {
	utils.DownloadFile(MAIN+ASSETS+asset.Id+".json", asset.Url)
}

func GetVersionFile(version Version) AssetIndex {
	name := MAIN + VERSIONS + version.Id + ".json"
	if !utils.FileExists(name) {
		file, _ := utils.GetString(version.URL)

		var version VersionFile
		json.Unmarshal([]byte(file), &version)

		err := ioutil.WriteFile(name, []byte(file), 0644)
		if err != nil {
			panic(err)
		}
		return version.AssetIndex
	}
	return AssetIndex{}
}

func parseVersionManifest() (VersionManifest, error) {
	var manifest VersionManifest
	file, err := utils.ReadStringFromFile(MANIFEST_PATH)
	json.Unmarshal([]byte(file), &manifest)
	return manifest, err
}

func (v VersionManifest) versionKeys() []string {
	keys := make([]string, len(v.Versions))
	for i, k := range v.Versions {
		keys[i] = k.Id
	}
	return keys
}

type VersionManifest struct {
	Versions []Version `json:"versions"`
}

type Version struct {
	Id      string `json:"id"`
	Type    string `json:"type"`
	Time    string `json:"time"`
	Release string `json:"releaseTime"`
	URL     string `json:"url"`
}

type VersionFile struct {
	AssetIndex AssetIndex `json:"assetIndex"`
}

type AssetIndex struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}
