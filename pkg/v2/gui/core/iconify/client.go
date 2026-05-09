package iconify

import (
	"context"
	"errors"
	"fmt"
	"image/color"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
)

const FallbackIcon = "subway:missing"

var errIconNotCached = errors.New("icon not cached locally")
var Icon = NewIconify()

type Iconify struct {
	CacheDir string
	Client   *http.Client

	cache       map[string]*SVGIcon
	failedUntil map[string]time.Time
}

func NewIconify() *Iconify {
	cfg, err := os.UserCacheDir()
	if err != nil {
		cfg = "."
	}

	return &Iconify{
		CacheDir: filepath.Join(cfg, "icons", "iconify"),
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache:       map[string]*SVGIcon{},
		failedUntil: map[string]time.Time{},
	}
}

func (i *Iconify) Layout(
	gtx layout.Context,
	name string, // "mdi:home"
	size unit.Dp,
	col color.NRGBA,
) layout.Dimensions {
	ic, err := i.Icon(context.Background(), name)
	if err != nil {
		return layout.Dimensions{}
	}

	return ic.Layout(gtx, size, col)
}

func (i *Iconify) EnsureFallback(ctx context.Context) {
	_, _ = i.Icon(ctx, FallbackIcon)
}

func (i *Iconify) Icon(ctx context.Context, name string) (*SVGIcon, error) {
	// already cached
	if ic, ok := i.cache[name]; ok {
		return ic, nil
	}
	if until, ok := i.failedUntil[name]; ok && time.Now().Before(until) {
		return nil, errIconNotCached
	}

	data, err := i.LoadSVG(ctx, name)
	if err != nil {
		i.failedUntil[name] = time.Now().Add(5 * time.Second)
		// fallback
		if name != FallbackIcon {
			return i.Icon(ctx, FallbackIcon)
		}
		return nil, err
	}

	ic := &SVGIcon{src: data}

	i.cache[name] = ic
	delete(i.failedUntil, name)
	return ic, nil
}

func (i *Iconify) LoadSVG(ctx context.Context, name string) ([]byte, error) {
	_ = ctx

	path, err := i.cachePath(name)
	if err != nil {
		return nil, err
	}

	// try cache first
	if data, err := os.ReadFile(path); err == nil {
		return data, nil
	}

	data, err := i.download(ctx, name)
	if err != nil {
		return nil, err
	}

	// save (best-effort)
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, data, 0o644)
	return data, nil
}

func (i *Iconify) download(ctx context.Context, name string) ([]byte, error) {
	prefix, icon, err := splitIconName(name)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(
		"https://api.iconify.design/%s/%s.svg?height=none",
		prefix,
		icon,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := i.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download icon %q: %s", name, res.Status)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(string(data), "<svg") {
		return nil, fmt.Errorf("download icon %q: response was not SVG", name)
	}

	return data, nil
}

func (i *Iconify) cachePath(name string) (string, error) {
	prefix, icon, err := splitIconName(name)
	if err != nil {
		return "", err
	}

	return filepath.Join(i.CacheDir, prefix, icon+".svg"), nil
}

func splitIconName(name string) (prefix string, icon string, err error) {
	parts := strings.Split(name, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid icon name %q, expected prefix:name", name)
	}

	prefix = strings.TrimSpace(parts[0])
	icon = strings.TrimSpace(parts[1])

	if prefix == "" || icon == "" {
		return "", "", fmt.Errorf("invalid icon name %q", name)
	}

	return prefix, icon, nil
}

func (i *Iconify) LayoutWithSize(
	gtx layout.Context,
	name string,
	size unit.Dp,
	col color.NRGBA,
) layout.Dimensions {
	ic, err := i.Icon(context.Background(), name)
	if err != nil {
		return layout.Dimensions{}
	}

	return ic.Layout(gtx, size, col)
}
