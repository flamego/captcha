// Copyright 2016 Minho. All rights reserved.
// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captchautil

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"math"
	"math/rand/v2"

	"github.com/pkg/errors"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type Complexity int

const (
	ComplexityLow Complexity = iota
	ComplexityMedium
	ComplexityHigh
)

// Image is a captcha image.
type Image struct {
	width  int
	height int
	dpi    float64
	nrgba  *image.NRGBA
}

type Option func(*Image) error

// CreateImage creates and returns a new image with given width, height, DPI,
// background color and additional options.
func CreateImage(width, height, dpi int, bgColor color.RGBA, opts ...Option) (*Image, error) {
	nrgba := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.Draw(nrgba, nrgba.Bounds(), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)

	img := &Image{
		height: height,
		width:  width,
		dpi:    float64(dpi),
		nrgba:  nrgba,
	}

	for _, opt := range opts {
		err := opt(img)
		if err != nil {
			return nil, err
		}
	}
	return img, nil
}

// Encode encodes the final PNG image to the given writer.
func (img *Image) Encode(w io.Writer) error {
	return png.Encode(w, img.nrgba)
}

// Noise draws background noise on the image.
func Noise(complex Complexity) Option {
	return func(img *Image) error {
		density := 18
		switch complex {
		case ComplexityLow:
			density = 28
		case ComplexityMedium:
			density = 18
		case ComplexityHigh:
			density = 8
		}
		maxSize := (img.height * img.width) / density
		for range maxSize {
			rw := rand.IntN(img.width)
			rh := rand.IntN(img.height)
			img.nrgba.Set(rw, rh, randomColor())
			size := rand.IntN(maxSize)
			if size%3 == 0 {
				img.nrgba.Set(rw+1, rh+1, randomColor())
			}
		}
		return nil
	}
}

// Text draws text on the image. All characters share one color, sit close
// enough to touch, have a line drawn through them, and are warped together, so
// they are hard to split into single characters.
func Text(text string) Option {
	return func(img *Image) error {
		if text == "" {
			return nil
		}

		w, h := float64(img.width), float64(img.height)
		layer := image.NewNRGBA(img.nrgba.Bounds())
		ink := randomDeepColor()

		d := &font.Drawer{Dst: layer, Src: image.NewUniform(ink)}
		cell := w * 0.8 / float64(len(text)) // widest a character may be
		step := cell * 0.85                  // 15% overlap
		x := (w - step*float64(len(text))) / 2
		// Keep characters off the border and within reach of the vertical warp.
		margin := min(6, h/4)
		for _, s := range text {
			f, err := randFontFamily()
			if err != nil {
				return errors.Wrap(err, "Text: get random font family")
			}
			newFace := func(size float64) (font.Face, error) {
				face, err := opentype.NewFace(f, &opentype.FaceOptions{
					Size:    size,
					DPI:     img.dpi,
					Hinting: font.HintingFull,
				})
				return face, errors.Wrap(err, "Text: create font face")
			}

			size := h * (0.62 + rand.Float64()*0.18)
			face, err := newFace(size)
			if err != nil {
				return err
			}
			// glyphBounds returns the pixels the character covers around the dot.
			glyphBounds := func() (left, top, right, bottom float64) {
				b, _, _ := face.GlyphBounds(s)
				return float64(b.Min.X.Floor()), float64(b.Min.Y.Floor()), float64(b.Max.X.Ceil()), float64(b.Max.Y.Ceil())
			}

			// Shrink characters that are too wide for their cell or too tall for
			// the image, such as with long text or a high DPI. Hinting rounds the
			// shrunk glyph, so it can take another try to fit.
			left, top, right, bottom := glyphBounds()
			for range 3 {
				scale := min(cell/(right-left), (h-2*margin)/(bottom-top))
				if scale >= 1 {
					break
				}
				_ = face.Close()
				size *= scale
				face, err = newFace(size)
				if err != nil {
					return err
				}
				left, top, right, bottom = glyphBounds()
			}

			// Center the character in its step and vertically, jitter it up or
			// down, and keep its ascender and descender inside the margins.
			y := (h-top-bottom)/2 + (rand.Float64()-0.5)*h*0.15
			y = max(min(y, h-margin-bottom), margin-top)
			d.Face = face
			d.Dot = fixed.P(int(x+(step-right+left)/2-left), int(y))
			d.DrawString(string(s))
			_ = face.Close()
			x += step
		}

		// Strike a wavy line through the characters in the same color.
		amp := h * 0.08
		period := w * (0.6 + rand.Float64()*0.4)
		phase := rand.Float64() * 2 * math.Pi
		mid := h/2 + (rand.Float64()-0.5)*h*0.2
		for px := img.width / 12; px < img.width-img.width/12; px++ {
			py := int(mid + amp*math.Sin(2*math.Pi*float64(px)/period+phase))
			for t := -1; t <= 1; t++ {
				layer.Set(px, py+t, ink)
			}
		}

		draw.Draw(img.nrgba, img.nrgba.Bounds(), warp(layer, 4, 5), image.Point{}, draw.Over)
		return nil
	}
}

// warp displaces every pixel of src along two sine waves, up to ax pixels
// horizontally and ay pixels vertically, sampling bilinearly.
func warp(src *image.NRGBA, ax, ay float64) *image.NRGBA {
	b := src.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewNRGBA(b)
	periodX := float64(w) * (0.4 + rand.Float64()*0.3)
	periodY := float64(h) * (0.8 + rand.Float64()*0.6)
	phaseX, phaseY := rand.Float64()*2*math.Pi, rand.Float64()*2*math.Pi

	at := func(x, y int) color.NRGBA {
		if x < 0 || y < 0 || x >= w || y >= h {
			return color.NRGBA{}
		}
		return src.NRGBAAt(x, y)
	}
	for y := range h {
		for x := range w {
			sx := float64(x) + ax*math.Sin(2*math.Pi*float64(y)/periodY+phaseX)
			sy := float64(y) + ay*math.Sin(2*math.Pi*float64(x)/periodX+phaseY)
			x0, y0 := int(math.Floor(sx)), int(math.Floor(sy))
			fx, fy := sx-float64(x0), sy-float64(y0)

			// Premultiply by alpha so transparent neighbors do not darken edges.
			var r, g, bl, a float64
			for _, s := range [4]struct {
				c      color.NRGBA
				weight float64
			}{
				{at(x0, y0), (1 - fx) * (1 - fy)},
				{at(x0+1, y0), fx * (1 - fy)},
				{at(x0, y0+1), (1 - fx) * fy},
				{at(x0+1, y0+1), fx * fy},
			} {
				sa := float64(s.c.A) * s.weight
				r += float64(s.c.R) * sa
				g += float64(s.c.G) * sa
				bl += float64(s.c.B) * sa
				a += sa
			}
			if a > 0 {
				dst.SetNRGBA(x, y, color.NRGBA{
					R: uint8(math.Round(r / a)),
					G: uint8(math.Round(g / a)),
					B: uint8(math.Round(bl / a)),
					A: uint8(math.Round(a)),
				})
			}
		}
	}
	return dst
}

// Border draws border on the image.
func Border(color color.RGBA) Option {
	return func(img *Image) error {
		for x := 0; x < img.width; x++ {
			img.nrgba.Set(x, 0, color)
			img.nrgba.Set(x, img.height-1, color)
		}
		for y := 0; y < img.height; y++ {
			img.nrgba.Set(0, y, color)
			img.nrgba.Set(img.width-1, y, color)
		}
		return nil
	}
}
