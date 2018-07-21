package main

import (
	"flag"
	"image/gif"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	inputDir := "wara"
	startFileIndex := 13
	framesPerSec := 24
	fileNames, err := listTargetFileNames(inputDir, startFileIndex)
	if err != nil {
		log.Fatal(err)
	}

	prevFrameIndex := 0
	currentFrameIndex := 0

	in, err := os.Open(filepath.Join(inputDir, fileNames[0]))
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()

	src, err := png.Decode(in)
	if err != nil {
		log.Fatal(err)
	}

	collectHistogram(src)
	palette := generatePalette()

	initializeCache()

	var dst gif.GIF
	dst.Image = make([]*image.Paletted, 1)
	dst.Image[0] = p
	dst.Delay = make([]int, 1)
	dst.Delay[0] = (currentFrameIndex * 100 / framesPerSec) - (prevFrameIndex * 100 / framesPerSec)
	prevFrameIndex = currentFrameIndex
	currentFrameIndex++

	out, err := os.Create(inputDir + ".gif")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	err = gif.EncodeAll(out, &dst)
	if err != nil {
		log.Fatal(err)
	}
}
