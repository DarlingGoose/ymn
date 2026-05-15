package assets

import (
	"bytes"
	_ "embed"
	"image"
	_ "image/png"
	"sync"
)

//go:embed yomuna.png
var yomunaPNG []byte

var (
	yomunaLogoOnce sync.Once
	yomunaLogo     image.Image
)

func YomunaLogo() image.Image {
	yomunaLogoOnce.Do(func() {
		img, _, err := image.Decode(bytes.NewReader(yomunaPNG))
		if err == nil {
			yomunaLogo = img
		}
	})
	return yomunaLogo
}
