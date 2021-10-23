// Copyright 2016 Minho. All rights reserved.
// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captchautil

import (
	"crypto/rand"
	"image/color"
	"math"
	"math/big"
)

func randomDeepColor() color.RGBA {
	randColor := randomColor()
	increase := float64(30 + r.Intn(255))

	red := math.Abs(math.Min(float64(randColor.R)-increase, 255))
	green := math.Abs(math.Min(float64(randColor.G)-increase, 255))
	blue := math.Abs(math.Min(float64(randColor.B)-increase, 255))
	return color.RGBA{
		R: uint8(red),
		G: uint8(green),
		B: uint8(blue),
		A: uint8(255),
	}
}

func randomColor() color.RGBA {
	red := r.Intn(255)
	green := r.Intn(255)
	blue := 0
	if (red + green) <= 400 {
		blue = 400 - green - red
	}
	if blue > 255 {
		blue = 255
	}
	return color.RGBA{
		R: uint8(red),
		G: uint8(green),
		B: uint8(blue),
		A: uint8(255),
	}
}

// RandomLightColor returns a random light color.
func RandomLightColor() color.RGBA {
	red := r.Intn(55) + 200
	green := r.Intn(55) + 200
	blue := r.Intn(55) + 200
	return color.RGBA{
		R: uint8(red),
		G: uint8(green),
		B: uint8(blue),
		A: uint8(255),
	}
}

// RandomText returns a generated string in given number of random characters.
func RandomText(n int) (string, error) {
	const alphanum = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	randomInt := func(max *big.Int) (int, error) {
		r, err := rand.Int(rand.Reader, max)
		if err != nil {
			return 0, err
		}
		return int(r.Int64()), nil
	}

	buffer := make([]byte, n)
	max := big.NewInt(int64(len(alphanum)))
	for i := 0; i < n; i++ {
		index, err := randomInt(max)
		if err != nil {
			return "", err
		}

		buffer[i] = alphanum[index]
	}

	return string(buffer), nil
}
