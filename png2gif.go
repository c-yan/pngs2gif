package main

import (
	"errors"
	"image"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
)

func collectHistograms(c *config, fileNames []string) error {
	for i := range fileNames {
		in, err := os.Open(filepath.Join(c.inputDir, fileNames[i]))
		if err != nil {
			return err
		}
		src, err := png.Decode(in)
		if err != nil {
			return err
		}
		in.Close()
		collectHistogram(src)
	}
	return nil
}

func doRectangleOptimization(p *image.Paletted, tmpFrameData []uint8, rows, cols []byte) (*image.Paletted, bool) {
	minX := -1
	for x := 0; x < p.Bounds().Max.X; x++ {
		if cols[x] != 0 {
			minX = x
			break
		}
	}
	if minX == -1 {
		return nil, true
	}
	maxX := 0
	for x := p.Bounds().Max.X - 1; x >= 0; x-- {
		if cols[x] != 0 {
			maxX = x + 1
			break
		}
	}
	minY := 0
	for y := 0; y < p.Bounds().Max.Y; y++ {
		if rows[y] != 0 {
			minY = y
			break
		}
	}
	maxY := 0
	for y := p.Bounds().Max.Y - 1; y >= 0; y-- {
		if rows[y] != 0 {
			maxY = y + 1
			break
		}
	}
	newPi := image.NewPaletted(image.Rect(minX, minY, maxX, maxY), p.Palette)
	for y := minY; y < maxY; y++ {
		bi1 := newPi.Stride * (y - minY)
		bi2 := p.Stride * y
		for x := minX; x < maxX; x++ {
			newPi.Pix[bi1+x-minX] = tmpFrameData[bi2+x]
		}
	}
	return newPi, false
}

func doTransparentColorOptimization(c *config, p *image.Paletted, prevFrameData *[]uint8) bool {
	if len(*prevFrameData) == 0 {
		*prevFrameData = p.Pix
		return false
	}
	rows := make([]byte, p.Bounds().Max.Y)
	cols := make([]byte, p.Bounds().Max.X)
	tmpFrameData := make([]uint8, len(p.Pix))
	for y := 0; y < p.Bounds().Max.Y; y++ {
		bi := p.Stride * y
		for x := 0; x < p.Bounds().Max.X; x++ {
			i := bi + x
			if (*prevFrameData)[i] == p.Pix[i] {
				tmpFrameData[i] = 0
			} else {
				tmpFrameData[i] = p.Pix[i]
			}
			rows[y] |= tmpFrameData[i]
			cols[x] |= tmpFrameData[i]
		}
	}
	if c.enableRectangleOptimizer {
		newPi, noDiffs := doRectangleOptimization(p, tmpFrameData, rows, cols)
		if noDiffs {
			return true
		}
		*prevFrameData = p.Pix
		*p = *newPi
	} else {
		*prevFrameData = p.Pix
		p.Pix = tmpFrameData
	}
	return false
}

func loadImage(c *config, path string) (image.Image, error) {
	in, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	src, err := png.Decode(in)
	if err != nil {
		return nil, err
	}
	in.Close()
	return src, nil
}

func validateImage(config image.Config, i image.Image) error {
	if config.Width != i.Bounds().Max.X-i.Bounds().Min.X || config.Height != i.Bounds().Max.Y-i.Bounds().Min.Y {
		return errors.New("Image size error")
	}
	return nil
}

func createGifData(c *config, fileNames []string) (*gif.GIF, error) {
	var dst gif.GIF

	palette := generatePalette()
	invalidateCache()

	prevFrameIndex := 0
	var prevFrameData []uint8
	for i := range fileNames {
		src, err := loadImage(c, filepath.Join(c.inputDir, fileNames[i]))
		if err != nil {
			return nil, err
		}
		if i == 0 {
			dst.Config = image.Config{
				ColorModel: newPalette(palette),
				Width:      src.Bounds().Max.X - src.Bounds().Min.X,
				Height:     src.Bounds().Max.Y - src.Bounds().Min.Y,
			}
		} else {
			err := validateImage(dst.Config, src)
			if err != nil {
				return nil, err
			}
		}
		pi := generatePalettedImage(src, palette)
		if c.enableTransparentColorOptimizer {
			noDiffs := doTransparentColorOptimization(c, pi, &prevFrameData)
			if c.enableRectangleOptimizer && noDiffs {
				continue
			}
		}
		dst.Image = append(dst.Image, pi)
		dst.Delay = append(dst.Delay, (i*100/c.framesPerSec)-(prevFrameIndex*100/c.framesPerSec))
		prevFrameIndex = i
	}
	return &dst, nil
}

func saveGifData(c *config, dst *gif.GIF) error {
	out, err := os.Create(c.inputDir + ".gif")
	if err != nil {
		return err
	}
	defer out.Close()

	err = gif.EncodeAll(out, dst)
	if err != nil {
		return err
	}
	return nil
}
