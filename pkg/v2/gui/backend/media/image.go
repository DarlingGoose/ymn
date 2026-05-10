package media

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sync"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/widget"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

type ImageView struct {
	Path string

	mu sync.RWMutex

	img image.Image
	op  paint.ImageOp

	loadedPath string
	loadingPath string
	err        error
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
		Fit:      widget.ScaleDown,
		Position: layout.Center,
		Scale:    1.0 / gtx.Metric.PxPerDp,
	}.Layout(gtx)
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

	img, _, err := image.Decode(f)
	if err != nil {
		v.finishDecode(path, nil, err)
		return
	}

	img = scaleDownImage(img, 2048)
	v.finishDecode(path, img, nil)
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
