package theme

import "image/color"

func Mix(a, b color.NRGBA, amount float32) color.NRGBA {
	if amount < 0 {
		amount = 0
	}
	if amount > 1 {
		amount = 1
	}

	inv := 1 - amount

	return color.NRGBA{
		R: uint8(float32(a.R)*amount + float32(b.R)*inv),
		G: uint8(float32(a.G)*amount + float32(b.G)*inv),
		B: uint8(float32(a.B)*amount + float32(b.B)*inv),
		A: uint8(float32(a.A)*amount + float32(b.A)*inv),
	}
}
