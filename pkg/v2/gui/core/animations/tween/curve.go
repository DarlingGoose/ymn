package tween

import "math"

// Curve maps x in [0,1] to y in [0,1].
type Curve struct {
	Name string
	Fn   func(x float64) float64
}

func (c Curve) At(x float64) float64 {
	if c.Fn == nil {
		return Clamp01(x)
	}
	return Clamp01(c.Fn(Clamp01(x)))
}

// Common curves.
var (
	Linear = Curve{
		Name: "linear",
		Fn: func(x float64) float64 {
			return x
		},
	}

	EaseInQuad = Curve{
		Name: "ease-in-quad",
		Fn: func(x float64) float64 {
			return x * x
		},
	}

	EaseOutQuad = Curve{
		Name: "ease-out-quad",
		Fn: func(x float64) float64 {
			return 1 - (1-x)*(1-x)
		},
	}

	EaseInOutQuad = Curve{
		Name: "ease-in-out-quad",
		Fn: func(x float64) float64 {
			if x < 0.5 {
				return 2 * x * x
			}
			return 1 - math.Pow(-2*x+2, 2)/2
		},
	}

	EaseInCubic = Curve{
		Name: "ease-in-cubic",
		Fn: func(x float64) float64 {
			return x * x * x
		},
	}

	EaseOutCubic = Curve{
		Name: "ease-out-cubic",
		Fn: func(x float64) float64 {
			return 1 - math.Pow(1-x, 3)
		},
	}

	EaseInOutCubic = Curve{
		Name: "ease-in-out-cubic",
		Fn: func(x float64) float64 {
			if x < 0.5 {
				return 4 * x * x * x
			}
			return 1 - math.Pow(-2*x+2, 3)/2
		},
	}

	EaseInSine = Curve{
		Name: "ease-in-sine",
		Fn: func(x float64) float64 {
			return 1 - math.Cos((x*math.Pi)/2)
		},
	}

	EaseOutSine = Curve{
		Name: "ease-out-sine",
		Fn: func(x float64) float64 {
			return math.Sin((x * math.Pi) / 2)
		},
	}

	EaseInOutSine = Curve{
		Name: "ease-in-out-sine",
		Fn: func(x float64) float64 {
			return -(math.Cos(math.Pi*x) - 1) / 2
		},
	}
)
