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

func collectHistograms(c *config, fileNames []string) {
	for i := range fileNames {
		in, err := os.Open(filepath.Join(c.inputDir, fileNames[i]))
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
}

func doTransparentColorOptimization(p *image.Paletted, prevFrameData *[]uint8) {
	if len(*prevFrameData) == 0 {
		*prevFrameData = p.Pix
	} else {
		tmpFrameData := make([]uint8, len(p.Pix))
		for i := range *prevFrameData {
			if (*prevFrameData)[i] == p.Pix[i] {
				tmpFrameData[i] = 0
			} else {
				tmpFrameData[i] = p.Pix[i]
			}
		}
		*prevFrameData = p.Pix
		p.Pix = tmpFrameData
	}
}

func createGifData(c *config, fileNames []string) *gif.GIF {
	var dst gif.GIF

	palette := generatePalette()
	invalidateCache()

	prevFrameIndex := 0
	var prevFrameData []uint8
	for i := range fileNames {
		in, err := os.Open(filepath.Join(c.inputDir, fileNames[i]))
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
		pi := generatePalettedImage(src, palette)
		if c.enableTransparentColorOptimizer {
			doTransparentColorOptimization(pi, &prevFrameData)
		}
		dst.Image = append(dst.Image, pi)
		dst.Delay = append(dst.Delay, (i*100/c.framesPerSec)-(prevFrameIndex*100/c.framesPerSec))
		prevFrameIndex = i
	}
	return &dst
}

func saveGifData(c *config, dst *gif.GIF) {
	out, err := os.Create(c.inputDir + ".gif")
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	err = gif.EncodeAll(out, dst)
	if err != nil {
		log.Fatal(err)
	}
}

type config struct {
	inputDir                        string
	startFileIndex                  int
	framesPerSec                    int
	enableTransparentColorOptimizer bool
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
	}

	fileNames, err := listTargetFileNames(c.inputDir, c.startFileIndex)
	if err != nil {
		log.Fatal(err)
	}

	collectHistograms(&c, fileNames)
	dst := createGifData(&c, fileNames)
	saveGifData(&c, dst)
}
