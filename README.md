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

## Launcher install

Install the `ymn` binary, the `ymn-guiv2` launcher wrapper, and the desktop entry used by Wofi and other app launchers:

```sh
sudo make install PREFIX=/usr
```

After installing, launch the v2 GUI with either:

```sh
ymn guiv2
ymn-guiv2
```

For package validation:

```sh
makepkg -si
makepkg --printsrcinfo > .SRCINFO
namcap PKGBUILD
namcap *.pkg.tar.zst
```

## AUR publish

The AUR package is a separate Git repository. The Makefile keeps that checkout under `.aur/ymn-git`, copies `PKGBUILD` and `.SRCINFO` into it, commits, and pushes:

```sh
make aur-push
```

To build the package first and then push the AUR metadata:

```sh
make aur-release
```

Override the AUR remote if needed:

```sh
make aur-push AUR_REMOTE=ssh://aur@aur.archlinux.org/ymn-git.git
```


## TODO
* add more to v2 gui pkg
  * modal - center,left,right - size
  * toast notifcations, center,left right,bottom,top, 
  * topbar with animations, build on tabs
  * tabs with animations
  * sidebar with tabs/animations
  * 

* implement key press + keybinding
* add more to panel
* create a backroundapi interface that all gui pkgs hit so i can easly mock it
* add a scrollbar
* add exit button on sidebar
* add a copy/highlight text component


## future features
- multi language
