package main

import (
	"github.com/HearthProject/oneclient-meta/utils"
	"encoding/json"
	"io/ioutil"
	"os"
	"fmt"
	"archive/zip"
)

const FORGE = "upstream/minecraftforge/"
const NETFORGE = "meta/net.minecraftforge/"
const FORGE_LIST = "http://files.minecraftforge.net/maven/net/minecraftforge/forge/json"

func forgeMeta() {
	os.RemoveAll(MAIN + FORGE)
	os.RemoveAll(MAIN + NETFORGE)
	utils.MakeDir(MAIN+NETFORGE)
	utils.MakeDir(MAIN+FORGE)
	data, err := utils.GetString(FORGE_LIST)
	if err != nil {
		panic(err)
	}
	var index ForgeIndex
	json.Unmarshal([]byte(data), &index)

	file, _ := json.MarshalIndent(index, "", "    ")
	err = ioutil.WriteFile(MAIN+FORGE+"index.json", []byte(file), 0644)

	forgeParse(index)
}



func forgeParse(index ForgeIndex) {

	for id, entry := range index.Number {
		fmt.Println(id + "->" + entry.Mcversion)
		if entry.Mcversion == "" {
			fmt.Printf("Skipping %v with missing mc version\n", entry.Build)
			continue
		}
		version := forgeVersion(index.Webpath, index.Artifact, entry)
		if version.url() == "" {
			fmt.Printf("Skipping %v with missing url\n", entry.Build)
			continue
		}
		jar := fmt.Sprintf(MAIN+FORGE+"%v", version.filename())
		if version.useInstaller() {
			profile := fmt.Sprintf(MAIN+FORGE+"%v.json", version.LongVersion)
			if !utils.FileExists(profile) {
				if !utils.FileExists(jar) {
					fmt.Printf("Downloading %v : %v %v\n", version.url(), jar, profile )
					utils.DownloadFile(jar, version.url())
					zip, err := zip.OpenReader(jar)
					if err != nil {
						fmt.Println("zip does not exist")
						continue
					}
					for _, file := range zip.File {
						if file.Name == "install_profile.json" {
							r, _ := file.Open()
							f, _ := ioutil.ReadAll(r)
							ioutil.WriteFile(profile, f, 0644)
							r.Close()
							break
						}
					}
					zip.Close()
				}
			}

		} else { //TODO LEGACY

		}
	}
}

type ForgeIndex struct {
	Adfocus   string                `json:"adfocus"`
	Artifact  string                `json:"artifact"`
	Name      string                `json:"name"`
	Number    map[string]ForgeEntry `json:"number"`
	Branches  map[string][]int      `json:"branches"`
	Homepage  string                `json:"homepage"`
	Mcversion string                `json:"mcversion"`
	Promos    int                   `json:"promo"`
	Webpath   string                `json:"webpath"`
}

type ForgeEntry struct {
	Branch    string     `json:"branch"`
	Build     int        `json:"build"`
	Files     [][]string `json:"files"`
	Mcversion string     `json:"mcversion"`
	Modified  float64    `json:"modified"`
	Version   string     `json:"version"`
}

type ForgeVersion struct {
	Build         int    `json:"build"`
	Version       string `json:"version"`
	Mcversion     string `json:"mcversions"`
	Branch        string `json:"branch"`
	InstallerFile string `json:"installer_file"`
	InstallerURL  string `json:"installer_url"`
	UniversalFile string `json:"universal_file"`
	UniversalURL  string `json:"universal_url"`
	ChangelogURL  string `json:"changelog_url"`
	LongVersion   string `json:"long_version"`
}

func (f ForgeVersion) useInstaller() bool {
	return !(f.InstallerURL == "" || f.Mcversion == "1.5.2")
}

func (f ForgeVersion) name() string {
	return fmt.Sprintf("Forge %v", f.Build)
}

func (f ForgeVersion) filename() string {
	if f.useInstaller() {
		return f.InstallerFile
	}
	return f.UniversalFile
}

func (f ForgeVersion) url() string {
	if f.useInstaller() {
		return f.InstallerURL
	}
	return f.UniversalURL
}

func forgeVersion(webpath, artifact string, entry ForgeEntry) ForgeVersion {

	version := ForgeVersion{Build: entry.Build, Version: entry.Version, Mcversion: entry.Mcversion, Branch: entry.Branch}

	version.LongVersion = entry.Mcversion + "-" + entry.Version
	if version.Branch != "" {
		version.LongVersion += "-" + version.Branch
	}
	for _, file := range entry.Files {
		extension := file[0]
		part := file[1]
		//checksum := file[2]
		filename := fmt.Sprintf("%v-%v-%v.%v", artifact,version.LongVersion, part, extension)
		url := fmt.Sprintf("%v%v/%v", webpath, version.LongVersion, filename)
		switch part {
		case "installer":
			version.InstallerFile = filename
			version.InstallerURL = url
		case "universal":
		case "client":
			version.UniversalFile = filename
			version.UniversalURL = url
		case "changelog":
			version.ChangelogURL = url
		}
	}
	return version

}
