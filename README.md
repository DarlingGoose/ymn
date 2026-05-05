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
* add clear cache button - should be in game settings
* add function to show how large the cache is in bytes - should be in game settings

* whole ui stops when audio is playing
* add translation logic to the row as well, and have it so when the button is pressed it goes 
* back and forth between the language

* add settings page for your default to language, use system settings if not default to english
* add support for tts in vntext/pkg/tts/TTS interface by using NewF5 tts
* all lines/vocab words that don't have audio should be able to use this to play audio, i should be able to reference a previous transript audio as a reference for the tts
* please add warning or something before using tts from ingame voice acting for the first time to say that this is intended to help with listening practice and emersion not to infiringe on the voice actors talents, you will not use this for anything that the voice actor would not agree too, and have it so they have to agree to use it.
* allow users to only show lines on transcript with just speakers
* add an option to minimize the timestamps in the live transcript

* would love option to delete all flashcards
* need to improve ux for game adding setup, specifliy the config and changing files and browsing
* 



* would also love a in game timer, so it logs how much time you spent in game     