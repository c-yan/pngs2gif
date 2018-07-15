package main

import (
	"io/ioutil"
	"sort"
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

func getIndex(name string) int {
	i, _ := strconv.Atoi(strings.Split(name, ".")[0])
	return i
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

	sort.Slice(result, func(i, j int) bool { return getIndex(result[i]) < getIndex(result[j]) })

	return result, nil
}
