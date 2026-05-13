package input

import (
	"errors"
	"fmt"
	"strings"
)

type Rule func(text string) error

func Validate(text string, rules ...Rule) error {
	var errs []error
	for _, rule := range rules {
		if rule == nil {
			continue
		}
		if err := rule(text); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func Required(message ...string) Rule {
	msg := "required"
	if len(message) > 0 && strings.TrimSpace(message[0]) != "" {
		msg = message[0]
	}

	return func(text string) error {
		if strings.TrimSpace(text) == "" {
			return errors.New(msg)
		}
		return nil
	}
}

func MinLen(n int, message ...string) Rule {
	msg := fmt.Sprintf("must be at least %d characters", n)
	if len(message) > 0 && strings.TrimSpace(message[0]) != "" {
		msg = message[0]
	}

	return func(text string) error {
		if len([]rune(text)) < n {
			return errors.New(msg)
		}
		return nil
	}
}

func MaxLen(n int, message ...string) Rule {
	msg := fmt.Sprintf("must be at most %d characters", n)
	if len(message) > 0 && strings.TrimSpace(message[0]) != "" {
		msg = message[0]
	}

	return func(text string) error {
		if len([]rune(text)) > n {
			return errors.New(msg)
		}
		return nil
	}
}

func NoWhitespace(message ...string) Rule {
	msg := "must not contain whitespace"
	if len(message) > 0 && strings.TrimSpace(message[0]) != "" {
		msg = message[0]
	}

	return func(text string) error {
		if strings.ContainsFunc(text, func(r rune) bool {
			return r == ' ' || r == '\t' || r == '\n' || r == '\r'
		}) {
			return errors.New(msg)
		}
		return nil
	}
}

func Prefix(prefix string, message ...string) Rule {
	msg := fmt.Sprintf("must start with %q", prefix)
	if len(message) > 0 && strings.TrimSpace(message[0]) != "" {
		msg = message[0]
	}

	return func(text string) error {
		if !strings.HasPrefix(text, prefix) {
			return errors.New(msg)
		}
		return nil
	}
}
