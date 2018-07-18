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
	startIndex := 13
	fileNames, err := listTargetFileNames(inputDir, startIndex)
	if err != nil {
		log.Fatal(err)
	}

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

	p := image.NewPaletted(image.Rect(0, 0, src.Bounds().Max.X-src.Bounds().Min.X, src.Bounds().Max.Y-src.Bounds().Min.Y), generatePalette())

	initializeCache()
	for y := src.Bounds().Min.Y; y < src.Bounds().Max.Y; y++ {
		bi := p.Stride * y
		for x := src.Bounds().Min.X; x < src.Bounds().Max.X; x++ {
			p.Pix[bi+x] = uint8(cachedIndex(p.Palette, src.At(x, y)))
		}
	}

	var dst gif.GIF
	dst.Image = make([]*image.Paletted, 1)
	dst.Image[0] = p
	dst.Delay = make([]int, 1)
	dst.Delay[0] = 0

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
