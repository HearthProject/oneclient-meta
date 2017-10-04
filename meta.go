package main

import (
	"strings"
	"fmt"
)

func main() {
	minecraftMeta()
	//forgeMeta()
	minecraftGenerate()
}

type GradleSpecifier struct {
	Group, Artifact, Version, Classifier string
}

func (g GradleSpecifier) IsNetty() bool {
	return g.Group == "com.mojang" && (g.Artifact == "netty" || g.Artifact == "patchy")
}

func (g GradleSpecifier) IsLWJGL() bool {
	return g.Group == "org.lwjgl.lwjgl" || g.Group == "net.java.jinput" || g.Group == "net.java.jutils"
}

func (g GradleSpecifier) String() string {
	if g.Classifier != "" {
		return fmt.Sprintf("%v:%v:%v:%v", g.Group, g.Artifact, g.Version, g.Classifier)
	}
	return fmt.Sprintf("%v:%v:%v", g.Group, g.Artifact, g.Version)
}

func (g *GradleSpecifier) MarshalJSON() ([]byte, error) {
	return []byte(g.String()), nil
}

func (g *GradleSpecifier) UnmarshalJSON(b []byte) error {
	return CreateSpecifier()
}

func CreateSpecifier(name string) GradleSpecifier {
	split := strings.Split(name, ":")

	if len(split) == 4 {
		return GradleSpecifier{Group: split[0], Artifact: split[1], Version: split[2], Classifier: split[3]}
	}
	return GradleSpecifier{Group: split[0], Artifact: split[1], Version: split[2]}
}
