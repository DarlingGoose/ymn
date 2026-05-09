package utils

import (
	"image/color"
	"strconv"
	"strings"
)

func HexNRGBA(hex string) color.NRGBA {
	c, _ := ParseHexNRGBA(hex)
	return c
}

func ParseHexNRGBA(hex string) (color.NRGBA, error) {
	s := strings.TrimSpace(hex)
	s = strings.TrimPrefix(s, "#")

	var r, g, b, a uint8 = 0, 0, 0, 255

	switch len(s) {
	case 3: // RGB shorthand: #abc
		rr, err := strconv.ParseUint(strings.Repeat(string(s[0]), 2), 16, 8)
		if err != nil {
			return color.NRGBA{}, err
		}
		gg, err := strconv.ParseUint(strings.Repeat(string(s[1]), 2), 16, 8)
		if err != nil {
			return color.NRGBA{}, err
		}
		bb, err := strconv.ParseUint(strings.Repeat(string(s[2]), 2), 16, 8)
		if err != nil {
			return color.NRGBA{}, err
		}
		r, g, b = uint8(rr), uint8(gg), uint8(bb)

	case 4: // RGBA shorthand: #abcd
		rr, err := strconv.ParseUint(strings.Repeat(string(s[0]), 2), 16, 8)
		if err != nil {
			return color.NRGBA{}, err
		}
		gg, err := strconv.ParseUint(strings.Repeat(string(s[1]), 2), 16, 8)
		if err != nil {
			return color.NRGBA{}, err
		}
		bb, err := strconv.ParseUint(strings.Repeat(string(s[2]), 2), 16, 8)
		if err != nil {
			return color.NRGBA{}, err
		}
		aa, err := strconv.ParseUint(strings.Repeat(string(s[3]), 2), 16, 8)
		if err != nil {
			return color.NRGBA{}, err
		}
		r, g, b, a = uint8(rr), uint8(gg), uint8(bb), uint8(aa)

	case 6: // RRGGBB
		v, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			return color.NRGBA{}, err
		}
		r = uint8(v >> 16)
		g = uint8(v >> 8)
		b = uint8(v)

	case 8: // RRGGBBAA
		v, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			return color.NRGBA{}, err
		}
		r = uint8(v >> 24)
		g = uint8(v >> 16)
		b = uint8(v >> 8)
		a = uint8(v)

	default:
		return color.NRGBA{}, strconv.ErrSyntax
	}

	return color.NRGBA{R: r, G: g, B: b, A: a}, nil
}
