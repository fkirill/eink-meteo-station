package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

func main() {
	entries, err := os.ReadDir(".")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".png") {
			buf, err := os.ReadFile("./" + entry.Name())
			if err != nil {
				panic(err)
			}
			encodedContent := base64.StdEncoding.EncodeToString(buf)
			fmt.Printf("const %s_src = \"data:image/png;base64, %s\"\n", strings.Replace(strings.ToUpper(entry.Name()[:1])+entry.Name()[1:], ".", "_", -1), encodedContent)
		}
	}
}
