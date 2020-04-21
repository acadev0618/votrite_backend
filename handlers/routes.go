package handlers

import (
	"fmt"
	"io/ioutil"
	"os"
)

type MethodInterface struct{}

func ReadRoutes(fileName string) (routes string) {
	file, e := ioutil.ReadFile(fileName)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}
	return string(file)
}
