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
			histogram[r>>16][g>>16][b>>16]++
		}
	}
}

func generatePalette() []color.Color {
	return palette.WebSafe
}
