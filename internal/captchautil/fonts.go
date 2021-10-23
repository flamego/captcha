// Copyright 2016 Minho. All rights reserved.
// Copyright 2021 Flamego. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captchautil

import (
	"embed"
	"path"
	"sync"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/pkg/errors"
)

//go:embed fonts/*.ttf
var fonts embed.FS

var (
	fontFamilyNames []string
	fontFamilies    map[string]*truetype.Font
	fontFamilyOnce  sync.Once
	fontFamilyErr   error
)

func randFontFamily() (*truetype.Font, error) {
	fontFamilyOnce.Do(func() {
		files, err := fonts.ReadDir("fonts")
		if err != nil {
			fontFamilyErr = errors.Wrap(err, "read directory")
			return
		}

		fontFamilyNames = make([]string, 0, len(files))
		fontFamilies = make(map[string]*truetype.Font, len(files))
		for _, fi := range files {
			data, err := fonts.ReadFile(path.Join("fonts", fi.Name()))
			if err != nil {
				fontFamilyErr = errors.Wrapf(err, "read file %q", fi.Name())
				return
			}

			f, err := freetype.ParseFont(data)
			if err != nil {
				fontFamilyErr = errors.Wrapf(err, "parse font %q", fi.Name())
				return
			}

			fontFamilyNames = append(fontFamilyNames, fi.Name())
			fontFamilies[fi.Name()] = f
		}
	})

	if fontFamilyErr != nil {
		return nil, fontFamilyErr
	}

	name := fontFamilyNames[r.Intn(len(fontFamilyNames))]
	return fontFamilies[name], nil
}
