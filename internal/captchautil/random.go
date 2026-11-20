package captchautil

import (
	crand "crypto/rand"
	"image/color"
	"math"
	"math/big"
	"math/rand/v2"
)

func randomDeepColor() color.RGBA {
	randColor := randomColor()
	increase := float64(120 + rand.IntN(135))

	red := math.Max(float64(randColor.R)-increase, 0)
	green := math.Max(float64(randColor.G)-increase, 0)
	blue := math.Max(float64(randColor.B)-increase, 0)
	return color.RGBA{
		R: uint8(red),
		G: uint8(green),
		B: uint8(blue),
		A: uint8(255),
	}
}

func randomColor() color.RGBA {
	red := rand.IntN(255)
	green := rand.IntN(255)
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
	red := rand.IntN(55) + 200
	green := rand.IntN(55) + 200
	blue := rand.IntN(55) + 200
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
		r, err := crand.Int(crand.Reader, max)
		if err != nil {
			return 0, err
		}
		return int(r.Int64()), nil
	}

	buffer := make([]byte, n)
	max := big.NewInt(int64(len(alphanum)))
	for i := range n {
		index, err := randomInt(max)
		if err != nil {
			return "", err
		}

		buffer[i] = alphanum[index]
	}

	return string(buffer), nil
}
