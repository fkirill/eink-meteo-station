package utils

import (
	"github.com/rotisserie/eris"
	"os"
	"path"
)

func GetRootDir() string {
	exec := os.Args[0]
	dir := path.Dir(exec)
	if dir == "." {
		curDir, err := os.Getwd()
		if err != nil {
			panic(eris.ToString(eris.Wrap(err, "Error getting current dir"), true))
		}
		dir = curDir
	}
	return dir
}
