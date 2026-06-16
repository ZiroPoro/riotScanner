package automation

import (
	"image"
	"image/color"
	"testing"
)

func TestScreenMatcher_MatchAll(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(5, 5, color.RGBA{R: 100, G: 150, B: 200, A: 255})

	probes := []PixelProbe{
		{X: 5, Y: 5, R: 100, G: 150, B: 200, Tolerance: 5},
	}

	m := NewScreenMatcher(probes)
	if !m.MatchAll(img) {
		t.Fatal("expected match")
	}

	probes[0].R = 0
	m2 := NewScreenMatcher(probes)
	if m2.MatchAll(img) {
		t.Fatal("expected no match")
	}
}

func TestAverageColorValues(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 100, B: 50, A: 255})
		}
	}

	r, g, b := AverageColor(img, 0, 0, 4, 4)
	if r != 200 || g != 100 || b != 50 {
		t.Fatalf("got (%d,%d,%d), want (200,100,50)", r, g, b)
	}
}
