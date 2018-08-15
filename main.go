package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")

type config struct {
	inputDir                        string
	startFileIndex                  int
	framesPerSec                    int
	enableTransparentColorOptimizer bool
	enableRectangleOptimizer        bool
}

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

	c := config{
		inputDir:                        "wara",
		startFileIndex:                  13,
		framesPerSec:                    24,
		enableTransparentColorOptimizer: true,
		enableRectangleOptimizer:        true,
	}

	fileNames, err := listTargetFileNames(c.inputDir, c.startFileIndex)
	if err != nil {
		log.Fatal(err)
	}
	err = collectHistograms(&c, fileNames)
	if err != nil {
		log.Fatal(err)
	}
	dst, err := createGifData(&c, fileNames)
	if err != nil {
		log.Fatal(err)
	}
	err = saveGifData(&c, dst)
	if err != nil {
		log.Fatal(err)
	}
}
