package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

func validateFileName(name string, startIndex int) bool {
	t := strings.Split(name, ".")
	if len(t) != 2 {
		return false
	}
	i, err := strconv.Atoi(t[0])
	if err != nil {
		return false
	}
	if i < startIndex {
		return false
	}
	if t[1] != "png" {
		return false
	}
	return true
}

func listFiles(path string, startIndex int) ([]string, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(files))
	for _, file := range files {
		name := file.Name()
		if validateFileName(name, startIndex) {
			result = append(result, name)
		}
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
