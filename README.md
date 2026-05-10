# Yomuna (YMN)

Yomuna is a visual novel game launcher and transcript/flashcard tool.

## Required Packages

Install these before running the GUI:

- `go` and `git` to build from source
- `wine` to install and launch Windows visual novels
- GStreamer runtime/plugins for dictionary audio playback through `jpndict`
- `mpv` for media preview/playback in the Bare file manager
- `ffmpeg` for thumbnails and media preview support

On Arch Linux, a practical baseline is:

```sh
sudo pacman -S go git wine gst-plugins-base gst-plugins-good gst-plugins-bad gst-plugins-ugly mpv ffmpeg
```

## Optional Packages

- `steam` if you want Steam/Proton launch flows
- Proton installed through Steam if using Proton runner configs
- `tesseract`, `grim`, `slurp`, and `hyprctl` for the older OCR workflow code paths
- `ollama` for translating

## Build

```sh
go build ./...
```

For package validation:

```sh
makepkg -si
namcap PKGBUILD
namcap *.pkg.tar.zst
```


## TODO
* add more to v2 gui pkg
  * modal - center,left,right - size
  * toast notifcations, center,left right,bottom,top, 
  * topbar with animations, build on tabs
  * tabs with animations
  * sidebar with tabs/animations
  * 
* use theme in v2 gui pkg
* icon button hover not working
* dropdown component, wierd ness
* add theme dropdown selector 
* implement key press + keybinding
* add more to panel
* create a backroundapi interface that all gui pkgs hit so i can easly mock it
* add a scrollbar
* add exit button on sidebar
* add a copy/highlight text component


## future features
- multi language