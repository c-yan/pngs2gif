package main

import (
	"image"
	"image/color"
	"math"
)

type byteQuad [4]uint8
type byteQuadPalette []byteQuad

func byte2dword(b uint8) uint32 {
	d := uint32(b)
	d |= d << 8
	return d
}

func (bq byteQuad) RGBA() (r, g, b, a uint32) {
	return byte2dword(bq[0]), byte2dword(bq[1]), byte2dword(bq[2]), byte2dword(bq[3])
}

func squareDiff(b1, b2 uint8) int {
	t := int(b1) - int(b2)
	return t * t
}

func (p byteQuadPalette) fastIndex(c byteQuad) int {
	bestDiff := math.MaxInt64
	bestIndex := -1
	for i, t := range p {
		diff := squareDiff(c[0], t[0]) + squareDiff(c[1], t[1]) + squareDiff(c[2], t[2])
		if diff < bestDiff {
			bestDiff = diff
			bestIndex = i
		}
	}
	return bestIndex
}

func (p byteQuadPalette) index(c color.Color) int {
	var bq byteQuad
	r, g, b, _ := c.RGBA()
	bq[0] = uint8(r >> 8)
	bq[1] = uint8(g >> 8)
	bq[2] = uint8(b >> 8)
	return p.fastIndex(bq)
}

type histogramElement struct {
	color    byteQuad
	quantity int64
	weight   [3]int64
}

func newHistogramElement(r, g, b int, quantity int64) (he histogramElement) {
	he.color[0] = uint8(r)
	he.color[1] = uint8(g)
	he.color[2] = uint8(b)
	he.color[3] = 255
	he.quantity = quantity
	for i := 0; i < 3; i++ {
		he.weight[i] = int64(he.color[i]) * he.quantity
	}
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

func calcNewColor(cluster []histogramElement) byteQuad {
	var (
		weightSum   [3]int64
		quantitySum int64
		c           byteQuad
	)
	for _, he := range cluster {
		for i := 0; i < 3; i++ {
			weightSum[i] += he.weight[i]
		}
		quantitySum += he.quantity
	}
	for i := 0; i < 3; i++ {
		c[i] = uint8(weightSum[i] / quantitySum)
	}
	c[3] = 255
	return c
}

func optimizePalette(p byteQuadPalette) []byteQuad {
	clusters := make([][]histogramElement, len(p))
	for _, he := range histogramElements {
		i := p.index(he.color)
		clusters[i] = append(clusters[i], he)
	}
	result := make([]byteQuad, 0, len(p))
	for _, cluster := range clusters {
		if len(cluster) == 0 {
			continue
		}
		result = append(result, calcNewColor(cluster))
	}
	return result
}

func getWebSafePalette() byteQuadPalette {
	var result [216]byteQuad
	for r := 0; r < 6; r++ {
		for g := 0; g < 6; g++ {
			for b := 0; b < 6; b++ {
				index := r*36 + g*6 + b
				result[index][0] = uint8(r * 51)
				result[index][1] = uint8(g * 51)
				result[index][2] = uint8(b * 51)
				result[index][3] = 255
			}
		}
	}
	return result[:]
}

func generatePalette() []byteQuad {
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
	return optimizePalette(optimizePalette(optimizePalette(getWebSafePalette())))
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

func cachedIndex(p byteQuadPalette, c color.Color) int {
	r, g, b, _ := c.RGBA()
	ci := (r&31)<<10 + (g&31)<<5 + b&31
	if (cache[ci].index != -1) && (cache[ci].c == c) {
		return cache[ci].index
	}
	cache[ci].index = p.index(c)
	cache[ci].c = c
	return cache[ci].index
}

func newPalette(p []byteQuad) color.Palette {
	result := make([]color.Color, len(p)+1)
	for i := range p {
		result[i+1] = p[i]
	}
	var c byteQuad
	result[0] = c
	return result
}

func generatePalettedImage(i image.Image, p []byteQuad) *image.Paletted {
	result := image.NewPaletted(image.Rect(0, 0, i.Bounds().Max.X-i.Bounds().Min.X, i.Bounds().Max.Y-i.Bounds().Min.Y), newPalette(p))
	for y := i.Bounds().Min.Y; y < i.Bounds().Max.Y; y++ {
		bi := result.Stride * y
		for x := i.Bounds().Min.X; x < i.Bounds().Max.X; x++ {
			result.Pix[bi+x] = uint8(cachedIndex(p, i.At(x, y)) + 1)
		}
	}
	return result
}
