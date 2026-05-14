package input

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
)

type NumberInput struct {
	*TextInput

	Min *float64
	Max *float64

	Step float64

	AllowDecimal  bool
	AllowNegative bool
	Clamp         bool
}

func (n *NumberInput) Float64() (float64, error) {
	if n == nil || n.TextInput == nil {
		return 0, nil
	}

	text := strings.TrimSpace(n.Text())
	if text == "" {
		return 0, errors.New("empty number")
	}

	v, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return 0, errors.New("invalid number")
	}

	return v, nil
}

func (n *NumberInput) Int() (int, error) {
	v, err := n.Float64()
	if err != nil {
		return 0, err
	}

	if math.Trunc(v) != v {
		return 0, errors.New("expected integer")
	}

	return int(v), nil
}

func (n *NumberInput) SetFloat64(v float64) {
	if n == nil || n.TextInput == nil {
		return
	}

	if n.Clamp {
		v = n.clamp(v)
	}

	if !n.AllowDecimal {
		n.SetText(strconv.FormatInt(int64(v), 10))
		return
	}

	n.SetText(strconv.FormatFloat(v, 'f', -1, 64))
}

func (n *NumberInput) SetInt(v int) {
	if n == nil || n.TextInput == nil {
		return
	}

	n.SetText(strconv.Itoa(v))
}

func (n *NumberInput) WithMin(v float64) *NumberInput {
	if n == nil {
		return n
	}
	n.Min = &v
	return n
}

func (n *NumberInput) WithMax(v float64) *NumberInput {
	if n == nil {
		return n
	}
	n.Max = &v
	return n
}

func (n *NumberInput) WithRange(min, max float64) *NumberInput {
	if n == nil {
		return n
	}
	n.Min = &min
	n.Max = &max
	return n
}

func (n *NumberInput) WithClamp(clamp bool) *NumberInput {
	if n == nil {
		return n
	}
	n.Clamp = clamp
	return n
}

func (n *NumberInput) WithStep(step float64) *NumberInput {
	if n == nil {
		return n
	}
	if step > 0 {
		n.Step = step
	}
	return n
}

func (n *NumberInput) Increment() {
	if n == nil {
		return
	}

	current, err := n.Float64()
	if err != nil {
		current = 0
	}

	step := n.Step
	if step <= 0 {
		step = 1
	}

	n.SetFloat64(current + step)
}

func (n *NumberInput) Decrement() {
	if n == nil {
		return
	}

	current, err := n.Float64()
	if err != nil {
		current = 0
	}

	step := n.Step
	if step <= 0 {
		step = 1
	}

	n.SetFloat64(current - step)
}

func (n *NumberInput) clamp(v float64) float64 {
	if n == nil {
		return v
	}

	if n.Min != nil && v < *n.Min {
		v = *n.Min
	}

	if n.Max != nil && v > *n.Max {
		v = *n.Max
	}

	return v
}

func (n *NumberInput) numberRule() Rule {
	return func(text string) error {
		text = strings.TrimSpace(text)
		if text == "" {
			return nil
		}

		v, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return errors.New("invalid number")
		}

		if !n.AllowNegative && v < 0 {
			return errors.New("must not be negative")
		}

		if !n.AllowDecimal && math.Trunc(v) != v {
			return errors.New("must be a whole number")
		}

		if n.Min != nil && v < *n.Min {
			return fmt.Errorf("must be at least %s", formatNumber(*n.Min))
		}

		if n.Max != nil && v > *n.Max {
			return fmt.Errorf("must be at most %s", formatNumber(*n.Max))
		}

		return nil
	}
}

func Number(message ...string) Rule {
	msg := "invalid number"
	if len(message) > 0 && strings.TrimSpace(message[0]) != "" {
		msg = message[0]
	}

	return func(text string) error {
		text = strings.TrimSpace(text)
		if text == "" {
			return nil
		}

		if _, err := strconv.ParseFloat(text, 64); err != nil {
			return errors.New(msg)
		}

		return nil
	}
}

func Integer(message ...string) Rule {
	msg := "must be a whole number"
	if len(message) > 0 && strings.TrimSpace(message[0]) != "" {
		msg = message[0]
	}

	return func(text string) error {
		text = strings.TrimSpace(text)
		if text == "" {
			return nil
		}

		v, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return errors.New("invalid number")
		}

		if math.Trunc(v) != v {
			return errors.New(msg)
		}

		return nil
	}
}

func NumberRange(min, max float64) Rule {
	return func(text string) error {
		text = strings.TrimSpace(text)
		if text == "" {
			return nil
		}

		v, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return errors.New("invalid number")
		}

		if v < min {
			return fmt.Errorf("must be at least %s", formatNumber(min))
		}

		if v > max {
			return fmt.Errorf("must be at most %s", formatNumber(max))
		}

		return nil
	}
}

func Positive(message ...string) Rule {
	msg := "must be greater than zero"
	if len(message) > 0 && strings.TrimSpace(message[0]) != "" {
		msg = message[0]
	}

	return func(text string) error {
		text = strings.TrimSpace(text)
		if text == "" {
			return nil
		}

		v, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return errors.New("invalid number")
		}

		if v <= 0 {
			return errors.New(msg)
		}

		return nil
	}
}

func NonNegative(message ...string) Rule {
	msg := "must not be negative"
	if len(message) > 0 && strings.TrimSpace(message[0]) != "" {
		msg = message[0]
	}

	return func(text string) error {
		text = strings.TrimSpace(text)
		if text == "" {
			return nil
		}

		v, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return errors.New("invalid number")
		}

		if v < 0 {
			return errors.New(msg)
		}

		return nil
	}
}

func formatNumber(v float64) string {
	if math.Trunc(v) == v {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func (n *NumberInput) WithThemeClient(tc *theme.Client) *NumberInput {
	if n == nil || n.TextInput == nil {
		return n
	}
	n.TextInput.WithThemeClient(tc)
	return n
}
