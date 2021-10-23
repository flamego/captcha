// Copyright 2016 Minho. All rights reserved.
// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captchautil

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"math"
	"math/rand"
	"time"

	"github.com/golang/freetype"
	"github.com/pkg/errors"
	"golang.org/x/image/font"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

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

// Encode encodes the final JPEG image to the given writer.
func (img *Image) Encode(w io.Writer) error {
	return jpeg.Encode(w, img.nrgba, &jpeg.Options{Quality: 100})
}

// Noise draws background noise on the image.
func Noise(complex Complexity) Option {
	return func(img *Image) error {
		density := 18
		if complex == ComplexityLow {
			density = 28
		} else if complex == ComplexityMedium {
			density = 18
		} else if complex == ComplexityHigh {
			density = 8
		}
		maxSize := (img.height * img.width) / density
		for i := 0; i < maxSize; i++ {
			rw := r.Intn(img.width)
			rh := r.Intn(img.height)
			img.nrgba.Set(rw, rh, randomColor())
			size := r.Intn(maxSize)
			if size%3 == 0 {
				img.nrgba.Set(rw+1, rh+1, randomColor())
			}
		}
		return nil
	}
}

// TextNoise draws background text noise on the image.
func TextNoise(complex Complexity) Option {
	return func(img *Image) error {
		density := 1500
		if complex == ComplexityLow {
			density = 2000
		} else if complex == ComplexityMedium {
			density = 1500
		} else if complex == ComplexityHigh {
			density = 1000
		}
		maxSize := (img.height * img.width) / density

		c := freetype.NewContext()
		c.SetDPI(img.dpi)
		c.SetClip(img.nrgba.Bounds())
		c.SetDst(img.nrgba)
		c.SetHinting(font.HintingFull)
		rawFontSize := float64(img.height) / (1 + float64(r.Intn(7))/float64(10))

		for i := 0; i < maxSize; i++ {
			text, err := RandomText(1)
			if err != nil {
				return errors.Wrap(err, "TextNoise: get random text")
			}
			fontSize := rawFontSize/2 + float64(r.Intn(5))

			c.SetSrc(image.NewUniform(RandomLightColor()))
			c.SetFontSize(fontSize)

			f, err := randFontFamily()
			if err != nil {
				return errors.Wrap(err, "TextNoise: get random font family")
			}
			c.SetFont(f)

			rw := r.Intn(img.width)
			rh := r.Intn(img.height)
			pt := freetype.Pt(rw, rh)

			_, err = c.DrawString(text, pt)
			if err != nil {
				return errors.Wrap(err, "TextNoise: draw string")
			}
		}
		return nil
	}
}

// Text draws text on the image.
func Text(text string) Option {
	return func(img *Image) error {
		c := freetype.NewContext()
		c.SetDPI(img.dpi)
		c.SetClip(img.nrgba.Bounds())
		c.SetDst(img.nrgba)
		c.SetHinting(font.HintingFull)

		fontWidth := img.width / len(text)
		for i, s := range text {
			fontSize := float64(img.height) / (1 + float64(r.Intn(7))/float64(9))
			c.SetSrc(image.NewUniform(randomDeepColor()))
			c.SetFontSize(fontSize)

			f, err := randFontFamily()
			if err != nil {
				return errors.Wrap(err, "Text: get random font family")
			}
			c.SetFont(f)

			x := (fontWidth)*i + (fontWidth)/int(fontSize)
			y := 5 + r.Intn(img.height/2) + int(fontSize/2)
			pt := freetype.Pt(x, y)

			_, err = c.DrawString(string(s), pt)
			if err != nil {
				return errors.Wrap(err, "Text: draw string")
			}
		}
		return nil
	}
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

// Curve draws a curve on the image.
func Curve() Option {
	return func(img *Image) error {
		random := func(min, max int64) float64 {
			decimal := rand.Float64()
			if max <= 0 {
				return (float64(rand.Int63n((min*-1)-(max*-1))+(max*-1)) + decimal) * -1
			}
			if min < 0 && max > 0 {
				if rand.Int()%2 == 0 {
					return float64(rand.Int63n(max)) + decimal
				} else {
					return (float64(rand.Int63n(min*-1)) + decimal) * -1
				}
			}
			return float64(rand.Int63n(max-min)+min) + decimal
		}

		a := r.Intn(img.height / 2)                            // Amplitude
		b := random(int64(-img.height/4), int64(img.height/4)) // Y-axis offset
		f := random(int64(-img.height/4), int64(img.height/4)) // X-axis offset
		// Period
		var t float64 = 0
		if img.height > img.width/2 {
			t = random(int64(img.width/2), int64(img.height))
		} else {
			t = random(int64(img.height), int64(img.width/2))
		}
		w := (2 * math.Pi) / t

		end := int(random(int64(float64(img.width)*0.8), int64(img.width)))
		c := color.RGBA{
			R: uint8(r.Intn(150)),
			G: uint8(r.Intn(150)),
			B: uint8(r.Intn(150)),
			A: uint8(255),
		}

		py := float64(0)
		for px := 0; px < end; px++ {
			if w != 0 {
				py = float64(a)*math.Sin(w*float64(px)+f) + b + (float64(img.width) / float64(5))
				i := img.height / 5
				for i > 0 {
					img.nrgba.Set(px+i, int(py), c)
					i--
				}
			}
		}
		return nil
	}
}
