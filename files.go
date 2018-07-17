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

func listFileNames(path string) ([]string, error) {
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

func filterFileNames(fileNames []string, startIndex int) []string {
	result := make([]string, 0, len(fileNames))
	for _, fileName := range fileNames {
		if validateFileName(fileName, startIndex) {
			result = append(result, fileName)
		}
	}
	return result
}

func numericalSortFileNames(fileNames []string) {
	sort.Slice(fileNames, func(i, j int) bool { return getIndex(fileNames[i]) < getIndex(fileNames[j]) })
}

func listTargetFileNames(path string, startIndex int) ([]string, error) {
	fileNames, err := listFileNames(path)
	if err != nil {
		return nil, err
	}

	result := filterFileNames(fileNames, startIndex)
	numericalSortFileNames(result)

	return result, nil
}
