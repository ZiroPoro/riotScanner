package automation

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"

	"github.com/disintegration/imaging"
)

type Region struct {
	X int `yaml:"x"`
	Y int `yaml:"y"`
	W int `yaml:"w"`
	H int `yaml:"h"`
}

type PixelProbe struct {
	X         int `yaml:"x"`
	Y         int `yaml:"y"`
	R         int `yaml:"r"`
	G         int `yaml:"g"`
	B         int `yaml:"b"`
	Tolerance int `yaml:"tolerance"`
}

type TapAction struct {
	Name   string       `yaml:"name"`
	Probes []PixelProbe `yaml:"probes"`
	TapX   int          `yaml:"tap_x"`
	TapY   int          `yaml:"tap_y"`
}

type ScreenMatcher struct {
	probes []PixelProbe
}

func NewScreenMatcher(probes []PixelProbe) *ScreenMatcher {
	return &ScreenMatcher{probes: probes}
}

func LoadImage(data []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	return img, nil
}

func (m *ScreenMatcher) MatchAll(img image.Image) bool {
	for _, p := range m.probes {
		if !matchPixel(img, p) {
			return false
		}
	}
	return true
}

func matchPixel(img image.Image, p PixelProbe) bool {
	bounds := img.Bounds()
	if p.X < bounds.Min.X || p.Y < bounds.Min.Y || p.X >= bounds.Max.X || p.Y >= bounds.Max.Y {
		return false
	}

	c := img.At(p.X, p.Y)
	r, g, b, _ := c.RGBA()
	r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)

	tol := p.Tolerance
	if tol == 0 {
		tol = 10
	}

	return abs(int(r8)-p.R) <= tol &&
		abs(int(g8)-p.G) <= tol &&
		abs(int(b8)-p.B) <= tol
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func FindColor(img image.Image, r, g, b, tolerance uint8) (int, int, bool) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			cr, cg, cb, _ := img.At(x, y).RGBA()
			r8, g8, b8 := uint8(cr>>8), uint8(cg>>8), uint8(cb>>8)
			if colorClose(r8, g8, b8, r, g, b, tolerance) {
				return x, y, true
			}
		}
	}
	return 0, 0, false
}

func colorClose(r1, g1, b1, r2, g2, b2, tol uint8) bool {
	return abs(int(r1)-int(r2)) <= int(tol) &&
		abs(int(g1)-int(g2)) <= int(tol) &&
		abs(int(b1)-int(b2)) <= int(tol)
}

func CropRegion(img image.Image, x, y, w, h int) image.Image {
	return imaging.Crop(img, image.Rect(x, y, x+w, y+h))
}

func AverageColor(img image.Image, x, y, w, h int) (r, g, b uint8) {
	region := CropRegion(img, x, y, w, h)
	bounds := region.Bounds()
	var sr, sg, sb int64
	n := int64(bounds.Dx() * bounds.Dy())
	if n == 0 {
		return 0, 0, 0
	}
	for py := bounds.Min.Y; py < bounds.Max.Y; py++ {
		for px := bounds.Min.X; px < bounds.Max.X; px++ {
			cr, cg, cb, _ := region.At(px, py).RGBA()
			sr += int64(cr >> 8)
			sg += int64(cg >> 8)
			sb += int64(cb >> 8)
		}
	}
	return uint8(sr / n), uint8(sg / n), uint8(sb / n)
}
