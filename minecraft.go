package main

import (
	"io/ioutil"
	"encoding/json"
	"os"
	"github.com/HearthProject/oneclient-meta/utils"
)

const MAIN = "oneclient/"
const MC = "upstream/minecraft/"
const NETMC = "meta/net.minecraft/"
const VERSION_MANIFEST = "https://launchermeta.mojang.com/mc/game/version_manifest.json"

func minecraftMeta() {
	os.RemoveAll(MAIN+MC)
	os.RemoveAll(MAIN+NETMC)
	if err := os.MkdirAll(MAIN+NETMC, 0700); err != nil {
		panic(err)
	}
	if err := os.MkdirAll(MAIN+MC, 0700); err != nil {
		panic(err)
	}
	manifest := parseVersionManifest()


	for _, version := range manifest["versions"].([]interface{}) {
		v := version.(map[string]interface{})
		createVersion(v)
	}
}

func createVersion(version map[string]interface{}) {
	name := MAIN + NETMC + version["id"].(string) + ".json"
	file, _ := utils.GetString(version["url"].(string))
	err := ioutil.WriteFile(name, []byte(file), 0644)
	if err != nil {
		panic(err)
	}
}

func parseVersionManifest() map[string]interface{} {
	var manifest map[string]interface{}
	err := utils.DownloadFile(MAIN+MC+"version_manifest.json", VERSION_MANIFEST)
	if err != nil {
		panic(err)
	}
	file := utils.ReadStringFromFile(MAIN+MC+"version_manifest.json")
	json.Unmarshal([]byte(file), &manifest)
	return manifest
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
