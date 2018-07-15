package main

import (
	"fmt"
	"log"
	"strings"
)

func main() {
	inputDir := "wara"
	startIndex := 13
	files, err := listFiles(inputDir, startIndex)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Join(files, ";"))
}
