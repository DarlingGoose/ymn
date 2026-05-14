package theme

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/animations/tween"
	"gopkg.in/yaml.v3"
)

const (
	ModeLight = "light"
	ModeDark  = "dark"
)

var DefaultThemeClient = New("", "wgl")

type Client struct {
	CurrentMode           string `json:"currentMode" yaml:"currentMode"`
	CurrentThemeName      string `json:"currentThemeName" yaml:"currentThemeName"`
	CurrentTypographyName string `json:"currentTypographyName" yaml:"currentTypographyName"`

	dir string
	mu  sync.RWMutex

	loadedThemes []*Theme
	currentTheme Theme

	loadedTypography  []*TypographyTokens
	currentTypography TypographyTokens

	colorTweens       *ColorTokenTweens
	colorTweenRunning bool
}

func New(dir string, name string) *Client {
	if dir == "" {
		configDir, err := os.UserConfigDir()
		if err == nil {
			dir = filepath.Join(configDir, name, "themes")
		} else {
			dir = filepath.Join(".", name, "themes")
		}
	}

	c := &Client{
		dir:                   dir,
		CurrentMode:           ModeDark,
		CurrentThemeName:      SlateCalm.Name,
		CurrentTypographyName: DefaultTypography.Name,

		loadedTypography: []*TypographyTokens{
			&DefaultTypography,
		},

		currentTheme:      SlateCalm,
		currentTypography: DefaultTypography,
	}
	c.loadedThemes = make([]*Theme, len(DefaultThemes))

	for i, theme := range DefaultThemes {
		if theme == nil {
			continue
		}

		c.loadedThemes[i] = new(*theme)
	}
	copy(c.loadedThemes, DefaultThemes)

	_ = c.EnsureDirs()

	c.loadedThemes = appendOrReplaceThemes(c.loadedThemes, c.LoadCustomThemes()...)
	c.loadedTypography = appendOrReplaceTypography(c.loadedTypography, c.LoadCustomTypography()...)

	_ = c.LoadSettings()

	if err := c.SelectTheme(c.CurrentThemeName); err != nil {
		_ = c.SelectTheme(SlateCalm.Name)
	}

	if err := c.SelectTypography(c.CurrentTypographyName); err != nil {
		_ = c.SelectTypography(DefaultTypography.Name)
	}

	return c
}

func (c *Client) GetCurrentMode() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.CurrentMode == ModeLight {
		return ModeLight
	}

	return ModeDark
}
func (c *Client) GetCurrentThemeName() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.CurrentThemeName
}
func (c *Client) GetCustomThemesPath() string {
	return filepath.Join(c.dir, "custom", "themes")
}

func (c *Client) GetCustomTypographyPath() string {
	return filepath.Join(c.dir, "custom", "typography")
}

func (c *Client) GetSettingsPath() string {
	return filepath.Join(c.dir, "settings.yaml")
}

func (c *Client) EnsureDirs() error {
	for _, dir := range []string{
		c.dir,
		c.GetCustomThemesPath(),
		c.GetCustomTypographyPath(),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) GetThemes() []*Theme {
	c.mu.RLock()
	defer c.mu.RUnlock()

	themes := make([]*Theme, 0, len(c.loadedThemes))
	themes = append(themes, c.loadedThemes...)

	sort.SliceStable(themes, func(i, j int) bool {
		return strings.ToLower(themes[i].Name) < strings.ToLower(themes[j].Name)
	})

	return themes
}

func (c *Client) SetMode(mode string) error {
	mode = strings.ToLower(strings.TrimSpace(mode))

	if mode != ModeLight && mode != ModeDark {
		return fmt.Errorf("invalid mode %q: expected %q or %q", mode, ModeLight, ModeDark)
	}

	c.mu.Lock()

	if c.CurrentMode == mode {
		c.mu.Unlock()
		return nil
	}

	now := time.Now()

	currentTokens := c.currentColorTokensLocked(now)
	nextTokens := themeColorsForMode(c.currentTheme, mode)

	if c.colorTweens == nil {
		c.colorTweens = NewColorTokenTweens(
			220*time.Millisecond,
			tween.EaseOutCubic,
			currentTokens,
		)
	} else {
		c.colorTweens.JumpTo(currentTokens)
	}

	c.colorTweens.AnimateToAt(now, nextTokens)
	c.colorTweenRunning = true

	c.CurrentMode = mode

	c.mu.Unlock()

	return c.SaveSettings()
}

func (c *Client) ToggleMode() error {
	c.mu.RLock()
	mode := c.CurrentMode
	c.mu.RUnlock()

	if mode == ModeDark {
		return c.SetMode(ModeLight)
	}

	return c.SetMode(ModeDark)
}

func (c *Client) SelectTheme(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("theme name is required")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, t := range c.loadedThemes {
		if t == nil {
			continue
		}

		if strings.EqualFold(t.Name, name) {
			now := time.Now()

			currentTokens := c.currentColorTokensLocked(now)
			nextTokens := themeColorsForMode(*t, c.CurrentMode)

			if c.colorTweens == nil {
				c.colorTweens = NewColorTokenTweens(
					220*time.Millisecond,
					tween.EaseOutCubic,
					currentTokens,
				)
			} else {
				c.colorTweens.JumpTo(currentTokens)
			}

			c.colorTweens.AnimateToAt(now, nextTokens)
			c.colorTweenRunning = true

			c.currentTheme = *t
			c.CurrentThemeName = t.Name

			go func() {
				_ = c.SaveSettings()
			}()

			return nil
		}
	}

	return fmt.Errorf("theme %q not found", name)
}

func (c *Client) GetCurrentTheme() Theme {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.currentTheme
}

func (c *Client) GetCurrentColorToken() *ColorTokens {
	c.mu.Lock()
	defer c.mu.Unlock()

	return new(c.currentColorTokensLocked(time.Now()))
}

func (c *Client) ColorTweenRunning() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.colorTweens == nil || !c.colorTweenRunning {
		return false
	}

	_, running := c.colorTweens.Value(time.Now())
	c.colorTweenRunning = running
	return running
}

func (c *Client) LoadCustomThemes() []*Theme {
	customDir := c.GetCustomThemesPath()

	entries, err := os.ReadDir(customDir)
	if err != nil {
		return nil
	}

	themes := make([]*Theme, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			continue
		}

		filePath := filepath.Join(customDir, entry.Name())

		loaded, err := loadThemeFile(filePath)
		if err != nil {
			// You may want to log this.
			continue
		}

		for _, t := range loaded {
			if t == nil || strings.TrimSpace(t.Name) == "" {
				continue
			}
			err := t.DarkColors.MustValidate()
			if err != nil {
				slog.Error("invalid dark theme", "name", t.Name, "err", err)
				continue
			}
			err = t.LightColors.MustValidate()
			if err != nil {
				slog.Error("invalid light theme", "name", t.Name, "err", err)
				continue
			}
			themes = append(themes, t)
		}
	}

	sort.SliceStable(themes, func(i, j int) bool {
		return strings.ToLower(themes[i].Name) < strings.ToLower(themes[j].Name)
	})

	return themes
}

func (c *Client) SaveCustomThemes(themes ...*Theme) error {
	customDir := c.GetCustomThemesPath()
	if err := os.MkdirAll(customDir, 0o755); err != nil {
		return err
	}

	for _, t := range themes {
		if t == nil {
			continue
		}
		err := t.DarkColors.MustValidate()
		if err != nil {
			return fmt.Errorf("invalid dark theme(%s): %w", t.Name, err)
		}
		err = t.LightColors.MustValidate()
		if err != nil {
			return fmt.Errorf("invalid light theme(%s): %w", t.Name, err)

		}
		name := safeThemeFileName(t.Name)
		if name == "" {
			return errors.New("theme name is required")
		}

		filePath := filepath.Join(customDir, name+".yaml")

		data, err := yaml.Marshal(t)
		if err != nil {
			return err
		}

		if err := os.WriteFile(filePath, data, 0o644); err != nil {
			return err
		}
	}

	c.mu.Lock()
	c.loadedThemes = appendOrReplaceThemes(c.loadedThemes, themes...)
	c.mu.Unlock()

	return nil
}

func (c *Client) ReloadCustomThemes() {
	custom := c.LoadCustomThemes()

	c.mu.Lock()
	defer c.mu.Unlock()
	base := make([]*Theme, len(DefaultThemes))

	for i, theme := range DefaultThemes {
		if theme == nil {
			continue
		}

		base[i] = new(*theme)
	}

	c.loadedThemes = appendOrReplaceThemes(base, custom...)

	for _, t := range c.loadedThemes {
		if t != nil && strings.EqualFold(t.Name, c.CurrentThemeName) {
			c.currentTheme = *t
			return
		}
	}

	c.currentTheme = SlateCalm
	c.CurrentThemeName = SlateCalm.Name
}

func (c *Client) LoadSettings() error {
	data, err := os.ReadFile(c.GetSettingsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var settings struct {
		CurrentMode           string `json:"currentMode" yaml:"currentMode"`
		CurrentThemeName      string `json:"currentThemeName" yaml:"currentThemeName"`
		CurrentTypographyName string `json:"currentTypographyName" yaml:"currentTypographyName"`
	}

	if err := yaml.Unmarshal(data, &settings); err != nil {
		return err
	}

	mode := strings.ToLower(strings.TrimSpace(settings.CurrentMode))
	if mode != ModeLight && mode != ModeDark {
		mode = ModeDark
	}

	c.mu.Lock()
	c.CurrentMode = mode

	if strings.TrimSpace(settings.CurrentThemeName) != "" {
		c.CurrentThemeName = settings.CurrentThemeName
	}

	if strings.TrimSpace(settings.CurrentTypographyName) != "" {
		c.CurrentTypographyName = settings.CurrentTypographyName
	}

	c.mu.Unlock()

	return nil
}

func (c *Client) SaveSettings() error {
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}

	c.mu.RLock()
	settings := struct {
		CurrentMode           string `json:"currentMode" yaml:"currentMode"`
		CurrentThemeName      string `json:"currentThemeName" yaml:"currentThemeName"`
		CurrentTypographyName string `json:"currentTypographyName" yaml:"currentTypographyName"`
	}{
		CurrentMode:           c.CurrentMode,
		CurrentThemeName:      c.CurrentThemeName,
		CurrentTypographyName: c.CurrentTypographyName,
	}
	c.mu.RUnlock()

	data, err := yaml.Marshal(settings)
	if err != nil {
		return err
	}

	return os.WriteFile(c.GetSettingsPath(), data, 0o644)
}

func loadTypographyFile(filePath string) ([]*TypographyTokens, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(filePath))

	var single TypographyTokens
	var many []*TypographyTokens

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &many); err == nil && len(many) > 0 {
			return many, nil
		}

		if err := json.Unmarshal(data, &single); err != nil {
			return nil, err
		}

	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &many); err == nil && len(many) > 0 {
			return many, nil
		}

		if err := yaml.Unmarshal(data, &single); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported typography file extension: %s", ext)
	}

	if strings.TrimSpace(single.Name) == "" {
		return nil, errors.New("typography is missing name")
	}

	single.FillDefaults()

	return []*TypographyTokens{&single}, nil
}

func loadThemeFile(filePath string) ([]*Theme, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(filePath))

	var single Theme
	var many []*Theme

	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &many); err == nil && len(many) > 0 {
			return many, nil
		}

		if err := json.Unmarshal(data, &single); err != nil {
			return nil, err
		}

	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &many); err == nil && len(many) > 0 {
			return many, nil
		}

		if err := yaml.Unmarshal(data, &single); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported theme file extension: %s", ext)
	}

	if strings.TrimSpace(single.Name) == "" {
		return nil, errors.New("theme is missing name")
	}

	return []*Theme{&single}, nil
}

func appendOrReplaceThemes(existing []*Theme, next ...*Theme) []*Theme {
	byName := make(map[string]int, len(existing))

	out := make([]*Theme, 0, len(existing)+len(next))

	for _, t := range existing {
		if t == nil || strings.TrimSpace(t.Name) == "" {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(t.Name))
		byName[key] = len(out)
		out = append(out, t)
	}

	for _, t := range next {
		if t == nil || strings.TrimSpace(t.Name) == "" {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(t.Name))
		if idx, ok := byName[key]; ok {
			out[idx] = t
			continue
		}

		byName[key] = len(out)
		out = append(out, t)
	}

	return out
}

func safeThemeFileName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(name))

	lastDash := false

	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastDash = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == '-' || r == '_' || r == ' ':
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}

	return strings.Trim(b.String(), "-")
}

func (c *Client) GetTypographyOptions() []*TypographyTokens {
	c.mu.RLock()
	defer c.mu.RUnlock()

	options := make([]*TypographyTokens, 0, len(c.loadedTypography))
	options = append(options, c.loadedTypography...)

	sort.SliceStable(options, func(i, j int) bool {
		return strings.ToLower(options[i].Name) < strings.ToLower(options[j].Name)
	})

	return options
}

func (c *Client) GetTypography(name string) (*TypographyTokens, bool) {
	name = strings.TrimSpace(name)

	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, t := range c.loadedTypography {
		if t == nil {
			continue
		}

		if strings.EqualFold(t.Name, name) {
			cp := *t
			return &cp, true
		}
	}

	return nil, false
}

func (c *Client) GetCurrentTypography() TypographyTokens {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.currentTypography
}

func (c *Client) SelectTypography(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("typography name is required")
	}

	c.mu.Lock()

	for _, t := range c.loadedTypography {
		if t == nil {
			continue
		}

		if strings.EqualFold(t.Name, name) {
			c.currentTypography = *t
			c.CurrentTypographyName = t.Name
			c.mu.Unlock()

			return c.SaveSettings()
		}
	}

	c.mu.Unlock()
	return fmt.Errorf("typography %q not found", name)
}

func (c *Client) LoadCustomTypography() []*TypographyTokens {
	customDir := c.GetCustomTypographyPath()

	entries, err := os.ReadDir(customDir)
	if err != nil {
		return nil
	}

	items := make([]*TypographyTokens, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			continue
		}

		filePath := filepath.Join(customDir, entry.Name())

		loaded, err := loadTypographyFile(filePath)
		if err != nil {
			continue
		}

		for _, t := range loaded {
			if t == nil || strings.TrimSpace(t.Name) == "" {
				continue
			}

			t.FillDefaults()
			items = append(items, t)
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	return items
}

func (c *Client) SaveCustomTypography(items ...*TypographyTokens) error {
	customDir := c.GetCustomTypographyPath()
	if err := os.MkdirAll(customDir, 0o755); err != nil {
		return err
	}

	for _, t := range items {
		if t == nil {
			continue
		}

		name := safeFileName(t.Name)
		if name == "" {
			return errors.New("typography name is required")
		}

		cp := *t
		cp.FillDefaults()

		data, err := yaml.Marshal(&cp)
		if err != nil {
			return err
		}

		filePath := filepath.Join(customDir, name+".yaml")
		if err := os.WriteFile(filePath, data, 0o644); err != nil {
			return err
		}
	}

	c.mu.Lock()
	c.loadedTypography = appendOrReplaceTypography(c.loadedTypography, items...)
	c.mu.Unlock()

	return nil
}

func (c *Client) ReloadCustomTypography() {
	custom := c.LoadCustomTypography()

	c.mu.Lock()
	defer c.mu.Unlock()

	base := []*TypographyTokens{
		&DefaultTypography,
	}

	c.loadedTypography = appendOrReplaceTypography(base, custom...)

	for _, t := range c.loadedTypography {
		if t != nil && strings.EqualFold(t.Name, c.CurrentTypographyName) {
			c.currentTypography = *t
			c.currentTypography.FillDefaults()
			return
		}
	}

	c.currentTypography = DefaultTypography
	c.CurrentTypographyName = DefaultTypography.Name
}

func (c *Client) currentColorTokensLocked(now time.Time) ColorTokens {
	if c.colorTweens != nil && c.colorTweenRunning {
		tokens, running := c.colorTweens.Value(now)
		c.colorTweenRunning = running
		return tokens
	}

	return themeColorsForMode(c.currentTheme, c.CurrentMode)
}

func themeColorsForMode(t Theme, mode string) ColorTokens {
	switch mode {
	case ModeLight:
		return t.LightColors
	case ModeDark:
		return t.DarkColors
	default:
		return t.DarkColors
	}
}
