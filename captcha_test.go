// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captcha

import (
	"bytes"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flamego/flamego"
	"github.com/flamego/session"
)

// newServer returns a server with the captcha middleware, a route that renders
// HTML(), a route that peeks at the secret text without consuming it, and a
// route that validates the "text" form value.
func newServer(opts ...Options) *flamego.Flame {
	f := flamego.NewWithLogger(&bytes.Buffer{})
	f.Use(session.Sessioner())
	f.Use(Captchaer(opts...))
	f.Get("/", func(c Captcha) string { return string(c.HTML()) })
	f.Get("/secret", func(s session.Session) string {
		text, _ := s.Get(textKey).(string)
		return text
	})
	f.Post("/", func(r *http.Request, c Captcha) string {
		return strconv.FormatBool(c.ValidText(r.FormValue("text")))
	})
	return f
}

// client sends requests to f and keeps the session cookie between them.
type client struct {
	t      *testing.T
	f      *flamego.Flame
	cookie string
}

func (c *client) do(method, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	if c.cookie != "" {
		req.Header.Set("Cookie", c.cookie)
	}
	resp := httptest.NewRecorder()
	c.f.ServeHTTP(resp, req)
	if cookie := resp.Header().Get("Set-Cookie"); cookie != "" {
		c.cookie = cookie
	}
	assert.Equal(c.t, http.StatusOK, resp.Code, "%s %s", method, target)
	return resp
}

func (c *client) secret() string {
	return c.do(http.MethodGet, "/secret").Body.String()
}

func (c *client) valid(text string) string {
	return c.do(http.MethodPost, "/?text="+url.QueryEscape(text)).Body.String()
}

func swapCase(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsUpper(r) {
			return unicode.ToLower(r)
		}
		return unicode.ToUpper(r)
	}, s)
}

func TestCaptchaer(t *testing.T) {
	c := &client{t: t, f: newServer()}

	want := `
<a class="captcha" href="javascript:" tabindex="-1">
	<img onclick="this.src=('/.captcha/image.png?refresh=true&amp;t='+Date.now())" src="/.captcha/image.png">
</a>`
	assert.Equal(t, want, c.do(http.MethodGet, "/").Body.String())

	resp := c.do(http.MethodGet, "/.captcha/image.png")
	assert.Equal(t, "image/png", resp.Header().Get("Content-Type"))
	assert.Equal(t, "no-store", resp.Header().Get("Cache-Control"))
	img, err := png.Decode(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, image.Rect(0, 0, 240, 80), img.Bounds())

	text := c.secret()
	assert.Len(t, text, 4)
}

func TestCaptcha_ValidText(t *testing.T) {
	t.Run("no image requested", func(t *testing.T) {
		c := &client{t: t, f: newServer()}
		assert.Equal(t, "false", c.valid(""))
	})

	t.Run("correct text, ignoring case", func(t *testing.T) {
		c := &client{t: t, f: newServer()}
		c.do(http.MethodGet, "/.captcha/image.png")
		assert.Equal(t, "true", c.valid(swapCase(c.secret())))
	})

	t.Run("text is single use", func(t *testing.T) {
		c := &client{t: t, f: newServer()}
		c.do(http.MethodGet, "/.captcha/image.png")
		text := c.secret()
		assert.Equal(t, "true", c.valid(text))
		assert.Equal(t, "false", c.valid(text))
	})

	t.Run("wrong text uses up the attempt", func(t *testing.T) {
		c := &client{t: t, f: newServer()}
		c.do(http.MethodGet, "/.captcha/image.png")
		text := c.secret()
		assert.Equal(t, "false", c.valid(text+"x"))
		assert.Equal(t, "false", c.valid(text))
	})
}

func TestCaptchaer_Refresh(t *testing.T) {
	c := &client{t: t, f: newServer()}
	c.do(http.MethodGet, "/.captcha/image.png")
	first := c.secret()

	// Reloading the image without refresh keeps the text the user is reading.
	c.do(http.MethodGet, "/.captcha/image.png")
	assert.Equal(t, first, c.secret())

	// Refresh replaces it, so the old text no longer validates. Retry a few
	// times in case the random text repeats.
	var second string
	for range 5 {
		c.do(http.MethodGet, "/.captcha/image.png?refresh=true")
		if second = c.secret(); second != first {
			break
		}
	}
	require.NotEqual(t, first, second)
	assert.Equal(t, "false", c.valid(first))
}

func TestCaptchaer_Options(t *testing.T) {
	c := &client{t: t, f: newServer(Options{
		URLPrefix:  "/captcha/",
		TextLength: 6,
		Width:      120,
		Height:     30,
	})}

	assert.Contains(t, c.do(http.MethodGet, "/").Body.String(), `src="/captcha/image.png"`)

	img, err := png.Decode(c.do(http.MethodGet, "/captcha/image.png").Body)
	require.NoError(t, err)
	assert.Equal(t, image.Rect(0, 0, 120, 30), img.Bounds())
	assert.Len(t, c.secret(), 6)

	// The default prefix is no longer served.
	resp := httptest.NewRecorder()
	c.f.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/.captcha/image.png", nil))
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestCaptchaer_TinyImage(t *testing.T) {
	for _, size := range [][2]int{{1, 1}, {3, 2}, {10, 3}} {
		c := &client{t: t, f: newServer(Options{Width: size[0], Height: size[1]})}
		img, err := png.Decode(c.do(http.MethodGet, "/.captcha/image.png").Body)
		require.NoError(t, err)
		assert.Equal(t, image.Rect(0, 0, size[0], size[1]), img.Bounds())
	}
}

func TestCaptchaer_Concurrent(t *testing.T) {
	f := flamego.NewWithLogger(&bytes.Buffer{})
	f.Use(session.Sessioner())
	f.Use(Captchaer())

	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			resp := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/.captcha/image.png", nil)
			f.ServeHTTP(resp, req)
			assert.Equal(t, http.StatusOK, resp.Code)
		})
	}
	wg.Wait()
}
