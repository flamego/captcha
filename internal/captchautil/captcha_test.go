// Copyright 2026 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captchautil

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestText_StaysInsideImage(t *testing.T) {
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	tests := []struct {
		text          string
		width, height int
		dpi           int
	}{
		{text: "gjpqy", width: 240, height: 80, dpi: 72},
		{text: "gjpqyQWM", width: 240, height: 80, dpi: 72},
		{text: "gjpqyQ", width: 120, height: 30, dpi: 72},
		{text: "gjAQ", width: 240, height: 80, dpi: 144},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s %dx%d %ddpi", test.text, test.width, test.height, test.dpi), func(t *testing.T) {
			// Glyph sizes, fonts, and positions are random, so render many times.
			for range 200 {
				img, err := CreateImage(test.width, test.height, test.dpi, white, Text(test.text))
				require.NoError(t, err)

				// The outermost pixels, where Border draws, must have no ink.
				edge := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
				b := img.nrgba.Bounds()
				for x := range b.Dx() {
					require.Equal(t, edge, img.nrgba.NRGBAAt(x, 0), "top edge at x=%d", x)
					require.Equal(t, edge, img.nrgba.NRGBAAt(x, b.Dy()-1), "bottom edge at x=%d", x)
				}
				for y := range b.Dy() {
					require.Equal(t, edge, img.nrgba.NRGBAAt(0, y), "left edge at y=%d", y)
					require.Equal(t, edge, img.nrgba.NRGBAAt(b.Dx()-1, y), "right edge at y=%d", y)
				}
			}
		})
	}
}
