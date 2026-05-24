package media

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"sync"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/widget"
	_ "golang.org/x/image/bmp"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

type ImageView struct {
	Path string
	Fit  widget.Fit

	mu sync.RWMutex

	img image.Image
	op  paint.ImageOp

	loadedPath  string
	loadingPath string
	err         error
}

func (v *ImageView) Load(path string) error {
	v.mu.RLock()
	if path == v.loadedPath && v.img != nil {
		v.mu.RUnlock()
		return nil
	}
	if path == v.loadingPath {
		v.mu.RUnlock()
		return nil
	}
	v.mu.RUnlock()

	v.mu.Lock()
	v.Path = path
	v.loadingPath = path
	v.err = nil
	v.mu.Unlock()

	go v.decode(path)
	return nil
}

func (v *ImageView) Layout(gtx layout.Context) layout.Dimensions {
	if v.Path != "" && v.Path != v.loadedPath {
		_ = v.Load(v.Path)
	}

	v.mu.RLock()
	img := v.img
	v.mu.RUnlock()

	if img == nil {
		gtx.Execute(op.InvalidateCmd{})
		return layout.Dimensions{}
	}

	size := img.Bounds().Size()

	return layout.Dimensions{
		Size: image.Pt(
			min(size.X, gtx.Constraints.Max.X),
			min(size.Y, gtx.Constraints.Max.Y),
		),
	}
}

func (v *ImageView) Draw(gtx layout.Context) layout.Dimensions {
	if v.Path != "" && v.Path != v.loadedPath {
		_ = v.Load(v.Path)
	}

	v.mu.RLock()
	img := v.img
	imgOp := v.op
	loading := v.loadingPath != "" && v.loadingPath != v.loadedPath
	v.mu.RUnlock()

	if img == nil {
		if loading {
			gtx.Execute(op.InvalidateCmd{})
		}
		return layout.Dimensions{}
	}

	return widget.Image{
		Src:      imgOp,
		Fit:      v.fit(),
		Position: layout.Center,
		Scale:    1.0 / gtx.Metric.PxPerDp,
	}.Layout(gtx)
}

func (v *ImageView) fit() widget.Fit {
	if v == nil || v.Fit == 0 {
		return widget.ScaleDown
	}
	return v.Fit
}

func (v *ImageView) Loading() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.loadingPath != "" && v.loadingPath != v.loadedPath
}

func (v *ImageView) Err() error {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.err
}

func (v *ImageView) decode(path string) {
	f, err := os.Open(path)
	if err != nil {
		v.finishDecode(path, nil, err)
		return
	}
	defer f.Close()

	reader, err := imageDecodeReader(path, f)
	if err != nil {
		v.finishDecode(path, nil, err)
		return
	}

	img, _, err := image.Decode(reader)
	if err != nil {
		v.finishDecode(path, nil, err)
		return
	}

	img = scaleDownImage(img, 2048)
	v.finishDecode(path, img, nil)
}

func imageDecodeReader(path string, r io.Reader) (io.Reader, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if !isRPGMakerMVEncryptedAsset(data) {
		return bytes.NewReader(data), nil
	}

	key, err := findRPGMakerMVEncryptionKey(path)
	if err != nil {
		return nil, err
	}
	decrypted, err := decryptRPGMakerMVAsset(data, key)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(decrypted), nil
}

func isRPGMakerMVEncryptedAsset(data []byte) bool {
	return len(data) >= 16 && bytes.Equal(data[:5], []byte("RPGMV"))
}

func decryptRPGMakerMVAsset(data []byte, key []byte) ([]byte, error) {
	const headerSize = 16

	if len(data) < headerSize {
		return nil, fmt.Errorf("RPG Maker MV asset is missing its header")
	}
	if len(key) == 0 {
		return nil, fmt.Errorf("RPG Maker MV encryption key is empty")
	}

	out := make([]byte, len(data)-headerSize)
	copy(out, data[headerSize:])

	n := min(len(key), len(out))
	for i := 0; i < n; i++ {
		out[i] ^= key[i]
	}

	return out, nil
}

func findRPGMakerMVEncryptionKey(assetPath string) ([]byte, error) {
	dir := filepath.Dir(assetPath)
	for {
		systemPath := filepath.Join(dir, "data", "System.json")
		key, err := readRPGMakerMVEncryptionKey(systemPath)
		if err == nil {
			return key, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}

		parent := filepath.Dir(dir)
		if parent == "" || parent == dir {
			break
		}
		dir = parent
	}

	return nil, fmt.Errorf("RPG Maker MV encryption key not found for %s", assetPath)
}

func readRPGMakerMVEncryptionKey(systemPath string) ([]byte, error) {
	data, err := os.ReadFile(systemPath)
	if err != nil {
		return nil, err
	}

	var system struct {
		EncryptionKey string `json:"encryptionKey"`
	}
	if err := json.Unmarshal(data, &system); err != nil {
		return nil, fmt.Errorf("read RPG Maker MV encryption key: %w", err)
	}
	if system.EncryptionKey == "" {
		return nil, fmt.Errorf("RPG Maker MV encryption key is missing from %s", systemPath)
	}

	key, err := hex.DecodeString(system.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("decode RPG Maker MV encryption key: %w", err)
	}
	return key, nil
}

func (v *ImageView) finishDecode(path string, img image.Image, err error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.loadingPath != path {
		return
	}

	if err != nil {
		v.err = err
		v.loadingPath = ""
		return
	}

	v.img = img
	v.op = paint.NewImageOp(img)
	v.loadedPath = path
	v.loadingPath = ""
	v.err = nil
}

func scaleDownImage(src image.Image, maxDim int) image.Image {
	bounds := src.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()
	if w <= maxDim && h <= maxDim {
		return src
	}

	var nw, nh int
	if w >= h {
		nw = maxDim
		nh = max(1, h*maxDim/w)
	} else {
		nh = maxDim
		nw = max(1, w*maxDim/h)
	}

	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, xdraw.Over, nil)
	return dst
}
