// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captcha

import (
	"fmt"
	"html/template"
	"image/color"
	"net/http"
	"reflect"
	"strings"

	"github.com/flamego/flamego"
	"github.com/flamego/flamego/inject"
	"github.com/flamego/session"

	"github.com/flamego/captcha/internal/captchautil"
)

// Captcha represents a captcha service and is used to validate text in captcha
// images.
type Captcha interface {
	// HTML returns the HTML content to display and refresh captcha images.
	HTML() template.HTML
	// ValidText validates the passed text against the secret text.
	ValidText(t string) bool
}

var _ Captcha = (*captcha)(nil)

type captcha struct {
	session   session.Session
	urlPrefix string
}

func (c *captcha) HTML() template.HTML {
	return template.HTML(
		fmt.Sprintf(`
<a class="captcha" href="javascript:" tabindex="-1">
	<img onclick="this.src=('%[1]simage.jpeg?refresh=true')" src="%[1]simage.jpeg">
</a>`,
			c.urlPrefix),
	)
}

const textKey = "flamego::captcha::text"

func (c *captcha) ValidText(t string) bool {
	defer c.session.Delete(textKey)

	want, ok := c.session.Get(textKey).(string)
	if !ok {
		return false
	}
	return strings.EqualFold(want, t)
}

// Options contains options for the captcha.Captchaer middleware.
type Options struct {
	// URLPrefix is the URL path prefix for serving captcha images. Default is
	// "/.captcha/".
	URLPrefix string
	// TextLength is the length of text to be generated in captcha images. Default
	// is 4.
	TextLength int
	// Width is the image width of captcha. Default is 240.
	Width int
	// Height is the image height of captcha. Default is 80.
	Height int
	// DPI is the image DPI of captcha. Default is 72.
	DPI int
}

var _ inject.FastInvoker = (*captchaInvoker)(nil)

// captchaInvoker is an inject.FastInvoker implementation of
// `func(flamego.Context, session.Session)`.
type captchaInvoker func(flamego.Context, session.Session)

func (invoke captchaInvoker) Invoke(args []interface{}) ([]reflect.Value, error) {
	invoke(args[0].(flamego.Context), args[1].(session.Session))
	return nil, nil
}

// Captchaer returns a middleware handler that injects captcha.Captcha into the
// request context, which is used for generating and validating text in captcha
// images.
func Captchaer(opts ...Options) flamego.Handler {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}

	parseOptions := func(opts Options) Options {
		if opts.URLPrefix == "" {
			opts.URLPrefix = "/.captcha/"
		}

		if opts.TextLength <= 0 {
			opts.TextLength = 4
		}

		if opts.Width <= 0 {
			opts.Width = 240
		}

		if opts.Height <= 0 {
			opts.Height = 80
		}

		if opts.DPI <= 0 {
			opts.DPI = 72
		}
		return opts
	}

	opt = parseOptions(opt)
	return captchaInvoker(func(c flamego.Context, s session.Session) {
		cpt := &captcha{
			session:   s,
			urlPrefix: opt.URLPrefix,
		}
		c.MapTo(cpt, (*Captcha)(nil))

		if !strings.HasPrefix(c.Request().URL.Path, opt.URLPrefix) {
			return
		}

		text, ok := s.Get(textKey).(string)
		if !ok || c.QueryBool("refresh") {
			var err error
			text, err = captchautil.RandomText(opt.TextLength)
			if err != nil {
				panic("captcha: generate random text: " + err.Error())
			}
			s.Set(textKey, text)
		}

		img, err := captchautil.CreateImage(
			opt.Width, opt.Height, opt.DPI, captchautil.RandomLightColor(),
			captchautil.Noise(captchautil.ComplexityLow),
			captchautil.TextNoise(captchautil.ComplexityLow),
			captchautil.Text(text),
			captchautil.Curve(),
			captchautil.Border(
				color.RGBA{R: 170, G: 170, B: 170, A: 255},
			),
		)
		if err != nil {
			panic("captcha: create image: " + err.Error())
		}

		c.ResponseWriter().Header().Set("Cache-Control", "no-store")
		c.ResponseWriter().WriteHeader(http.StatusOK)
		err = img.Encode(c.ResponseWriter())
		if err != nil {
			panic("captcha: encode image: " + err.Error())
		}
	})
}
