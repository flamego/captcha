// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captcha

import (
	"bytes"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flamego/flamego"
	"github.com/flamego/session"
)

func TestCaptchaer(t *testing.T) {
	f := flamego.NewWithLogger(&bytes.Buffer{})
	f.Use(session.Sessioner())
	f.Use(Captchaer())

	f.Get("/.captcha/image.jpeg", func(c Captcha) {
		want := template.HTML(`
<a class="captcha" href="javascript:" tabindex="-1">
	<img onclick="this.src=('/.captcha/image.jpeg?refresh=true')" src="/.captcha/image.jpeg">
</a>`)
		assert.Equal(t, want, c.HTML())
	})
	f.Post("/", func(s session.Session, c Captcha) string {
		return strconv.FormatBool(c.ValidText(s.Get(textKey).(string)))
	})

	resp := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/.captcha/image.jpeg", nil)
	assert.Nil(t, err)

	f.ServeHTTP(resp, req)

	cookie := resp.Header().Get("Set-Cookie")

	// Make a request again using the same session ID
	resp = httptest.NewRecorder()
	req, err = http.NewRequest(http.MethodPost, "/", nil)
	assert.Nil(t, err)

	req.Header.Set("Cookie", cookie)
	f.ServeHTTP(resp, req)

	assert.Equal(t, "true", resp.Body.String())
}
