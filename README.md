# auth

[![GitHub Workflow Status](https://img.shields.io/github/workflow/status/flamego/captcha/Go?logo=github&style=for-the-badge)](https://github.com/flamego/captcha/actions?query=workflow%3AGo)
[![Codecov](https://img.shields.io/codecov/c/gh/flamego/captcha?logo=codecov&style=for-the-badge)](https://app.codecov.io/gh/flamego/captcha)
[![GoDoc](https://img.shields.io/badge/GoDoc-Reference-blue?style=for-the-badge&logo=go)](https://pkg.go.dev/github.com/flamego/captcha?tab=doc)
[![Sourcegraph](https://img.shields.io/badge/view%20on-Sourcegraph-brightgreen.svg?style=for-the-badge&logo=sourcegraph)](https://sourcegraph.com/github.com/flamego/captcha)

Package captcha is a middleware that provides captcha service for [Flamego](https://github.com/flamego/flamego).

## Installation

The minimum requirement of Go is **1.16**.

	go get github.com/flamego/captcha

## Getting started

```html
<!-- templates/home.tmpl -->
<form>
  {{.CaptchaHTML}}
</form>
```

```go
package main

import (
	"net/http"

	"github.com/flamego/captcha"
	"github.com/flamego/flamego"
	"github.com/flamego/session"
	"github.com/flamego/template"
)

func main() {
	f := flamego.Classic()
	f.Use(session.Sessioner())
	f.Use(captcha.Captchaer())
	f.Use(template.Templater())
	f.Get("/", func(t template.Template, data template.Data, captcha captcha.Captcha) {
		data["CaptchaHTML"] = captcha.HTML()
		t.HTML(http.StatusOK, "home")
	})
	f.Run()
}
```

## License

This project is under the MIT License. See the [LICENSE](LICENSE) file for the full license text.
