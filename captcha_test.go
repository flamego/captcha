// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captcha

import (
	"bytes"
	"html/template"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flamego/flamego"
	"github.com/flamego/session"
)

func TestCaptchaer(t *testing.T) {
	f := flamego.NewWithLogger(&bytes.Buffer{})
	f.Use(session.Sessioner())
	f.Use(Captchaer())

	f.Get("/.captcha/image.png", func(c Captcha) {
		want := template.HTML(`
<a class="captcha" href="javascript:" tabindex="-1">
	<img onclick="this.src=('/.captcha/image.png?refresh=true')" src="/.captcha/image.png">
</a>`)
		assert.Equal(t, want, c.HTML())
	})
	f.Post("/", func(s session.Session, c Captcha) string {
		return strconv.FormatBool(c.ValidText(s.Get(textKey).(string)))
	})

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/.captcha/image.png", nil)
	assert.Nil(t, err)

	f.ServeHTTP(resp, req)

	assert.Equal(t, "image/png", resp.Header().Get("Content-Type"))
	_, err = png.Decode(resp.Body)
	assert.Nil(t, err)

	cookie := resp.Header().Get("Set-Cookie")

	// Make a request again using the same session ID
	resp = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodPost, "/", nil)
	assert.Nil(t, err)

	req.Header.Set("Cookie", cookie)
	f.ServeHTTP(resp, req)

	assert.Equal(t, "true", resp.Body.String())
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
