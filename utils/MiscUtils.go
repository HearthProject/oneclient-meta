package utils

import "github.com/deckarep/golang-set"

func StringSet(strings []string) mapset.Set {
	set :=mapset.NewSet()
	for _,s := range strings {
		set.Add(s)
	}
	return set
}