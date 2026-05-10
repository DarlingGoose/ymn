package input

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func FileExists(message ...string) Rule {
	msg := "file does not exist"
	if len(message) > 0 && strings.TrimSpace(message[0]) != "" {
		msg = message[0]
	}

	return func(text string) error {
		path := strings.TrimSpace(text)
		if path == "" {
			return nil
		}

		info, err := os.Stat(path)
		if err != nil {
			return errors.New(msg)
		}
		if info.IsDir() {
			return errors.New("path is a directory, expected file")
		}
		return nil
	}
}

func DirExists(message ...string) Rule {
	msg := "directory does not exist"
	if len(message) > 0 && strings.TrimSpace(message[0]) != "" {
		msg = message[0]
	}

	return func(text string) error {
		path := strings.TrimSpace(text)
		if path == "" {
			return nil
		}

		info, err := os.Stat(path)
		if err != nil {
			return errors.New(msg)
		}
		if !info.IsDir() {
			return errors.New("path is a file, expected directory")
		}
		return nil
	}
}

func PathExists(message ...string) Rule {
	msg := "path does not exist"
	if len(message) > 0 && strings.TrimSpace(message[0]) != "" {
		msg = message[0]
	}

	return func(text string) error {
		path := strings.TrimSpace(text)
		if path == "" {
			return nil
		}

		if _, err := os.Stat(path); err != nil {
			return errors.New(msg)
		}
		return nil
	}
}

func ExecutableFile(message ...string) Rule {
	msg := "file is not executable"
	if len(message) > 0 && strings.TrimSpace(message[0]) != "" {
		msg = message[0]
	}

	return func(text string) error {
		path := strings.TrimSpace(text)
		if path == "" {
			return nil
		}

		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return errors.New("path is a directory, expected executable file")
		}

		// Windows executables may not have Unix executable bits.
		if runtime.GOOS == "windows" {
			return nil
		}

		if info.Mode()&0o111 == 0 {
			return errors.New(msg)
		}

		return nil
	}
}

func HasExt(ext string, message ...string) Rule {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	msg := fmt.Sprintf("path must end with %s", ext)
	if len(message) > 0 && strings.TrimSpace(message[0]) != "" {
		msg = message[0]
	}

	return func(text string) error {
		path := strings.TrimSpace(text)
		if path == "" || ext == "" {
			return nil
		}

		if strings.ToLower(filepath.Ext(path)) != ext {
			return errors.New(msg)
		}
		return nil
	}
}

func HasAnyExt(exts ...string) Rule {
	normalized := make([]string, 0, len(exts))
	for _, ext := range exts {
		ext = strings.ToLower(strings.TrimSpace(ext))
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		normalized = append(normalized, ext)
	}

	return func(text string) error {
		path := strings.TrimSpace(text)
		if path == "" || len(normalized) == 0 {
			return nil
		}

		got := strings.ToLower(filepath.Ext(path))
		for _, ext := range normalized {
			if got == ext {
				return nil
			}
		}

		return fmt.Errorf("path must end with one of: %s", strings.Join(normalized, ", "))
	}
}
