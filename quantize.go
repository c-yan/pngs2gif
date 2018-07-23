package main

import (
	"image"
	"image/color"
	"image/color/palette"
)

var histogram [256][256][256]int64

func collectHistogram(i image.Image) {
	for y := i.Bounds().Min.Y; y < i.Bounds().Max.Y; y++ {
		for x := i.Bounds().Min.X; x < i.Bounds().Max.X; x++ {
			r, g, b, _ := i.At(x, y).RGBA()
			histogram[r>>8][g>>8][b>>8]++
		}
	}
}

func generatePalette() []color.Color {
	return palette.WebSafe
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
