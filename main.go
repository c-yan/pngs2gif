package main

import (
	"flag"
	"image"
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

	for i := range fileNames {
		in, err := os.Open(filepath.Join(inputDir, fileNames[i]))
		if err != nil {
			log.Fatal(err)
		}
		src, err := png.Decode(in)
		if err != nil {
			log.Fatal(err)
		}
		in.Close()
		collectHistogram(src)
	}

	palette := generatePalette()
	invalidateCache()

	var dst gif.GIF
	prevFrameIndex := 0
	for i := range fileNames {
		in, err := os.Open(filepath.Join(inputDir, fileNames[i]))
		if err != nil {
			log.Fatal(err)
		}
		src, err := png.Decode(in)
		if err != nil {
			log.Fatal(err)
		}
		in.Close()
		if i == 0 {
			dst.Config = image.Config{
				ColorModel: newPalette(palette),
				Width:      src.Bounds().Max.X - src.Bounds().Min.X,
				Height:     src.Bounds().Max.Y - src.Bounds().Min.Y,
			}
		}
		dst.Image = append(dst.Image, generatePalettedImage(src, palette))
		dst.Delay = append(dst.Delay, (i*100/framesPerSec)-(prevFrameIndex*100/framesPerSec))
		prevFrameIndex = i
	}

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
