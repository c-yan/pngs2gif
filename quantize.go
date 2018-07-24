package main

import (
	"image"
	"image/color"
	"image/color/palette"
)

type histogramElement struct {
	color    color.RGBA
	quantity int64
	weightR  int64
	weightG  int64
	weightB  int64
}

func newHistogramElement(r, g, b int, quantity int64) (he histogramElement) {
	he.color = color.RGBA{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: 255,
	}
	he.quantity = quantity
	he.weightR = int64(r) * quantity
	he.weightG = int64(g) * quantity
	he.weightB = int64(b) * quantity
	return he
}

var (
	colors            int64
	histogram         [256][256][256]int64
	histogramElements []histogramElement
)

func collectHistogram(i image.Image) {
	for y := i.Bounds().Min.Y; y < i.Bounds().Max.Y; y++ {
		for x := i.Bounds().Min.X; x < i.Bounds().Max.X; x++ {
			r, g, b, _ := i.At(x, y).RGBA()
			r >>= 8
			g >>= 8
			b >>= 8
			if histogram[r][g][b] == 0 {
				colors++
			}
			histogram[r][g][b]++
		}
	}
}

func optimizePalette(p color.Palette) []color.Color {
	clusters := make([][]histogramElement, len(p))
	for _, he := range histogramElements {
		i := p.Index(he.color)
		clusters[i] = append(clusters[i], he)
	}
	result := make([]color.Color, 0, len(p))
	for _, cluster := range clusters {
		if len(cluster) == 0 {
			continue
		}
		var weightRSum, weightGSum, weightBSum, quantitySum int64
		for _, he := range cluster {
			weightRSum += he.weightR
			weightGSum += he.weightG
			weightBSum += he.weightB
			quantitySum += he.quantity
		}
		result = append(result, color.RGBA{
			R: uint8(weightRSum / quantitySum),
			G: uint8(weightGSum / quantitySum),
			B: uint8(weightBSum / quantitySum),
		})
	}
	return result
}

func generatePalette() []color.Color {
	histogramElements = make([]histogramElement, 0, colors)
	for r := 0; r < 256; r++ {
		for g := 0; g < 256; g++ {
			for b := 0; b < 256; b++ {
				if histogram[r][g][b] != 0 {
					histogramElements = append(histogramElements, newHistogramElement(r, g, b, histogram[r][g][b]))
				}
			}
		}
	}
	return optimizePalette(optimizePalette(optimizePalette(palette.WebSafe)))
}

type lookupCacheElement struct {
	c     color.Color
	index int
}

var cache [32768]lookupCacheElement

func initializeCache() {
	for i := range cache {
		cache[i].index = -1
	}
}

func cachedIndex(p color.Palette, c color.Color) int {
	r, g, b, _ := c.RGBA()
	ci := (r&31)<<10 + (g&31)<<5 + b&31
	if (cache[ci].index != -1) && (cache[ci].c == c) {
		return cache[ci].index
	}
	cache[ci].index = p.Index(c)
	cache[ci].c = c
	return cache[ci].index
}

func generatePalettedImage(i image.Image, p []color.Color) *image.Paletted {
	result := image.NewPaletted(image.Rect(0, 0, i.Bounds().Max.X-i.Bounds().Min.X, i.Bounds().Max.Y-i.Bounds().Min.Y), p)
	for y := i.Bounds().Min.Y; y < i.Bounds().Max.Y; y++ {
		bi := result.Stride * y
		for x := i.Bounds().Min.X; x < i.Bounds().Max.X; x++ {
			result.Pix[bi+x] = uint8(cachedIndex(result.Palette, i.At(x, y)))
		}
	}
	return result
}
