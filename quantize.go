package main

import (
	"image"
	"image/color"
	"math"
	"sort"
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

var candidaes [32768][]int

func invalidateCandidates() {
	for i := range candidaes {
		candidaes[i] = nil
	}
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func calcMaxDistance(x byte, y [2]byte) int {
	if x < y[0] {
		return int(y[1]) - int(x)
	} else if x > y[1] {
		return int(x) - int(y[0])
	} else {
		return max(int(y[1])-int(x), int(x)-int(y[0]))
	}
}

func calcMinDistance(x byte, y [2]byte) int {
	if x < y[0] {
		return int(y[0]) - int(x)
	} else if x > y[1] {
		return int(x) - int(y[1])
	} else {
		return 0
	}
}

func calculateCandidates(p byteQuadPalette, c byteQuad) []int {
	var result []int
	bounds := [3][2]byte{{c[0] & 0xf8, c[0]&0xf8 + 7}, {c[1] & 0xf8, c[1]&0xf8 + 7}, {c[2] & 0xf8, c[2]&0xf8 + 7}}
	minMaxError := math.MaxInt64
	for _, pe := range p {
		dr := calcMaxDistance(pe[0], bounds[0])
		e := dr * dr
		if e > minMaxError {
			continue
		}
		dg := calcMaxDistance(pe[1], bounds[1])
		db := calcMaxDistance(pe[2], bounds[2])
		e += dg*dg + db*db
		if e < minMaxError {
			minMaxError = e
		}
	}
	for i := range p {
		dr := calcMinDistance(p[i][0], bounds[0])
		e := dr * dr
		if e > minMaxError {
			continue
		}
		dg := calcMinDistance(p[i][1], bounds[1])
		db := calcMinDistance(p[i][2], bounds[2])
		e += dg*dg + db*db
		if e < minMaxError {
			result = append(result, i)
		}
	}
	return result
}

func (p byteQuadPalette) fastIndex(c byteQuad) int {
	block := (int(c[0])&0xf8)<<7 + (int(c[1])&0xf8)<<2 + int(c[2])>>3
	if len(candidaes[block]) == 0 {
		candidaes[block] = calculateCandidates(p, c)
	}
	bestDiff := math.MaxInt64
	bestIndex := -1
	for _, i := range candidaes[block] {
		t := p[i]
		diff := squareDiff(c[0], t[0])
		if diff > bestDiff {
			continue
		}
		diff += squareDiff(c[1], t[1]) + squareDiff(c[2], t[2])
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

func calcCentroid(cluster []histogramElement) byteQuad {
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

func optimizePalette(p byteQuadPalette) ([]byteQuad, [][]histogramElement) {
	clusters := make([][]histogramElement, len(p))
	for _, he := range histogramElements {
		i := p.fastIndex(he.color)
		clusters[i] = append(clusters[i], he)
	}
	newPalette := make([]byteQuad, 0, len(p))
	newCluster := make([][]histogramElement, 0, len(p))
	for _, cluster := range clusters {
		if len(cluster) == 0 {
			continue
		}
		newPalette = append(newPalette, calcCentroid(cluster))
		newCluster = append(newCluster, cluster)
	}
	invalidateCandidates()
	return newPalette, newCluster
}

func divideCluster(cluster []histogramElement, color byteQuad, index int) (byteQuad, byteQuad) {
	var c0, c1 []histogramElement
	for _, he := range cluster {
		if color[index] < he.color[index] {
			c0 = append(c0, he)
		} else {
			c1 = append(c1, he)
		}
	}
	return calcCentroid(c0), calcCentroid(c1)
}

func calcWorstColorIndex(clusterError *[3]int64) int {
	var worstError int64
	worstIndex := -1
	for i := 0; i < 3; i++ {
		if clusterError[i] > worstError {
			worstError = clusterError[i]
			worstIndex = i
		}
	}
	return worstIndex
}

func calcWorstCluster(p byteQuadPalette, clusters [][]histogramElement) (int, int) {
	worstError := int64(-1)
	worstClusterIndex := -1
	worstColorIndex := 0
	for i := range clusters {
		var clusterError [3]int64
		pe := p[i]
		for _, he := range clusters[i] {
			for j := 0; j < 3; j++ {
				clusterError[j] += int64(squareDiff(pe[j], he.color[j])) * he.quantity
			}
		}
		errSum := clusterError[0] + clusterError[1] + clusterError[2]
		if errSum > worstError {
			worstError = errSum
			worstClusterIndex = i
			worstColorIndex = calcWorstColorIndex(&clusterError)
		}
	}
	return worstClusterIndex, worstColorIndex
}

func populatePalette(p byteQuadPalette) []byteQuad {
	p, clusters := optimizePalette(p)
	worstClusterIndex, worstColorIndex := calcWorstCluster(p, clusters)
	c1, c2 := divideCluster(clusters[worstClusterIndex], p[worstClusterIndex], worstColorIndex)
	invalidateCandidates()
	p[worstClusterIndex] = c1
	return append(p, c2)
}

func calcBrightness(c byteQuad) float64 {
	return float64(c[0])*0.299 + float64(c[1])*0.587 + float64(c[2])*0.114
}

func sortPalette(p []byteQuad) {
	sort.Slice(p, func(i, j int) bool { return calcBrightness(p[i]) < calcBrightness(p[j]) })
}

func createHistogramElements() []histogramElement {
	result := make([]histogramElement, 0, colors)
	for r := 0; r < 256; r++ {
		for g := 0; g < 256; g++ {
			for b := 0; b < 256; b++ {
				if histogram[r][g][b] != 0 {
					result = append(result, newHistogramElement(r, g, b, histogram[r][g][b]))
				}
			}
		}
	}
	return result
}

func generatePalette() []byteQuad {
	histogramElements = createHistogramElements()
	p := make([]byteQuad, 1)
	for {
		p = populatePalette(p)
		if len(p) == 255 {
			break
		}
	}
	for i := 0; i < 3; i++ {
		p, _ = optimizePalette(p)
	}
	sortPalette(p)
	return p
}

type lookupCacheElement struct {
	c     color.Color
	index int
}

var cache [32768]lookupCacheElement

func invalidateCache() {
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
