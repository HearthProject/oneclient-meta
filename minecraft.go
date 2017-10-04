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

func minecraftGenerate() {
	dir, _ := ioutil.ReadDir(MAIN + VERSIONS)
	for _, filename := range dir {
		var mojangVersion MojangVersionFile
		file, err := utils.ReadStringFromFile(MAIN + VERSIONS + filename.Name())
		if err != nil {
			panic(err)
		}
		json.Unmarshal([]byte(file), &mojangVersion)
		versionFile := mojangVersion.ToOneClient("Minecraft", "net.minecraft")
		var finalLibs []OneclientLibrary
		for _, lib := range versionFile.Libraries {
			specifier := lib.Name
			if specifier.IsNetty() {
				fmt.Println("Ignoring Mojang netty hack in version " + versionFile.Version)
				continue
			} else if specifier.IsLWJGL() {
				//TODO lwjgl
			} else {
				finalLibs = append(finalLibs, lib)
			}
		}
		versionFile.Libraries = finalLibs
		//TODO legacy
		data, _ := json.MarshalIndent(versionFile, "", "     ")
		err = ioutil.WriteFile(MAIN+NETMC+versionFile.Version+".json", []byte(data), 0644)
		if err != nil {
			panic(err)
		}
	}

}

func minecraftMeta() {
	os.RemoveAll(MAIN + NETMC)
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

		var version MojangVersionFile
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

func (m MojangVersionFile) ToOneClient(name, uid string) OneclientVersionFile {
	oc := OneclientVersionFile{
		Name:               name,
		UID:                uid,
		Version:            m.Id,
		AssetIndex:         m.AssetIndex,
		MainClass:          m.MainClass,
		MinecraftArguments: m.MinecraftArguments,
		ReleaseTime:        m.ReleaseTime,
		Type:               m.Type,
		//TODO lwjgl 3
		Requires: map[string]string{"org.lwjgl": "2.*"},
		Order:    -2,
	}

	if m.Id != "" {
		oc.MainJar = OneclientLibrary{Name: CreateSpecifier(fmt.Sprintf("com.mojang:minecraft:%v:client", m.Id))}
		oc.MainJar.Download = m.Downloads.Client.ToOneClient()
	}

	libs := make([]OneclientLibrary, len(m.Libraries))
	for i, l := range m.Libraries {
		libs[i] = l.ToOneClient()
	}
	oc.Libraries = libs

	return oc
}

func (m MojangLibrary) ToOneClient() OneclientLibrary {
	return OneclientLibrary{Name: CreateSpecifier(m.Name), Download: m.Download.ToOneClient()}
}

func (m MojangArtifact) ToOneClient() OneclientArtifact {
	return OneclientArtifact{m.Sha1, m.Size, m.Url}
}

func (m MojangDownload) ToOneClient() OneclientDownload {
	return OneclientDownload{Artifact: OneclientArtifact{Sha1: m.Artifact.Sha1, Url: m.Artifact.Url, Size: m.Artifact.Size}}
}

func (m JarDownload) ToOneClient() OneclientDownload {
	return OneclientDownload{Artifact: OneclientArtifact{Sha1: m.Sha1, Url: m.Url, Size: m.Size}}
}

type OneclientVersionFile struct {
	Name               string             `json:"name,omitempty"`
	Version            string             `json:"name,omitempty"`
	UID                string             `json:"uid,omitempty"`
	ParentUID          string             `json:"parentUid,omitempty"`
	Requires           map[string]string  `json:"requires,omitempty"`
	AssetIndex         AssetIndex         `json:"assetIndex,omitempty"`
	Libraries          []OneclientLibrary `json:"libraries,omitempty"`
	MainJar            OneclientLibrary   `json:"mainJar,omitempty"`
	JarMods            []OneclientLibrary `json:"jarMods,omitempty"`
	MainClass          string             `json:"mainClass,omitempty"`
	AppletClass        string             `json:"appletClass,omitempty"`
	MinecraftArguments string             `json:"minecraftArguments,omitempty"`
	ReleaseTime        string             `json:"releaseTime,omitempty"`
	Type               string             `json:"type,omitempty"`
	AddTraits          string             `json:"+traits,omitempty"`
	AddTweakers        string             `json:"+tweakers,omitempty"`
	Order              int                `json:"order,omitempty"`
}

type OneclientLibrary struct {
	Name     GradleSpecifier   `json:"name,omitempty"`
	Download OneclientDownload `json:"downloads,omitempty"`
	Url      string            `json:"url,omitempty"`
	Hint     string            `json:"1C-hint,omitempty"`
}

type OneclientDownload struct {
	Artifact  OneclientArtifact  `json:"artifact,omitempty"`
	Classifer OneclientClassifer `json:"classifiers,omitempty"`
	Rules     []MojangRule       `json:"rules,omitempty"`
	Natives                      `json:"natives,omitempty"`
}

type OneclientClassifer struct {
	Artifact OneclientArtifact `json:"artifact,omitempty"`
}

type OneclientArtifact struct {
	Sha1 string `json:"sha1,omitempty"`
	Size int    `json:"size,omitempty"`
	Url  string `json:"url,omitempty"`
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

type MojangVersionFile struct {
	AssetIndex             AssetIndex      `json:"assetIndex"`
	Assets                 string          `json:"assets"`
	Id                     string          `json:"id"`
	Downloads              Jars            `json:"downloads"`
	MainClass              string          `json:"mainClass"`
	MinecraftArguments     string          `json:"minecraftArguments"`
	ReleaseTime            string          `json:"releaseTime"`
	MinimumLauncherVersion int             `json:"minimumLauncherVersion"`
	Time                   string          `json:"time"`
	Type                   string          `json:"type"`
	Libraries              []MojangLibrary `json:"libraries"`
}

type Jars struct {
	Client JarDownload `json:"client"`
	Server JarDownload `json:"server"`
}

type JarDownload struct {
	Sha1 string `json:"sha1"`
	Size int    `json:"size"`
	Url  string `json:"url"`
}

type MojangLibrary struct {
	Name     string         `json:"name,omitempty"`
	Download MojangDownload `json:"downloads,omitempty"`
}

type MojangDownload struct {
	Artifact  MojangArtifact  `json:"artifact"`
	Classifer MojangClassifer `json:"classifiers"`
	Rules     []MojangRule    `json:"rules"`
	Natives                   `json:"natives"`
}

type MojangClassifer struct {
	MojangArtifact `json:"artifact,omitempty"`
}
type MojangRule struct {
	Action string `json:"action,omitempty"`
	OS            `json:"os,omitempty"`
}

type OS struct {
	Name string `json:"name"`
}

type Natives struct {
	Linux   string `json:"linux,omitempty"`
	OSX     string `json:"osx,omitempty"`
	Windows string `json:"windows,omitempty"`
}

type Extract struct {
	Exclude []string `json:"excludes,omitempty"`
}

type MojangArtifact struct {
	Sha1 string `json:"sha1,omitempty"`
	Size int    `json:"size,omitempty"`
	Url  string `json:"url,omitempty"`
	Path string `json:"path,omitempty"`
}

type AssetIndex struct {
	Id        string `json:"id"`
	Url       string `json:"url"`
	Sha1      string `json:"sha1"`
	Size      int    `json:"size"`
	TotalSize int    `json:"totalSize"`
}

type Logging struct {
	Client Logger `json:"client"`
}

type Logger struct {
	File            `json:"file"`
	Argument string `json:"argument"`
	Type     string `json:"type"`
}

type File struct {
	Sha1 string `json:"sha1"`
	Size int    `json:"size"`
	Url  string `json:"url"`
	Id   string `json:"id"`
}
