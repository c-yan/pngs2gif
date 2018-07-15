package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

func listFiles(path string, startIndex int) ([]string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(files))
	for _, file := range files {
		result = append(result, file.Name())
	}
	return result, nil
}

func main() {
	inputDir := "wara"
	startIndex := 13
	files, err := listFiles(inputDir, startIndex)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Join(files, ";"))
}
