# Yomuna

Yomuna (`ymn`) is a visual novel launcher, transcript reader, translation helper, and flashcard tool for Japanese reading workflows on Linux.

It is mainly useful when you want one place to launch a game, capture text from a running visual novel, inspect vocabulary, translate lines, and turn selected words or tokens into review cards.

## Features

- Launch visual novels with Wine, Proton-style runners, or Gamescope-backed Wine configs.
- Install and manage game configs from the GUI.
- Follow live extracted text from supported games and hooks.
- Filter transcript text by hook when multiple text streams are available.
- Analyze selected transcript sentences with token/vocabulary lookup.
- Translate transcript lines with an Ollama-backed local translation setup.
- Create, edit, search, delete, and page through flashcards by game.
- Mark sentence-analysis tokens that already exist in your flashcard deck.
- Sync flashcards to Anki through AnkiConnect.
- View and adjust app, transcript, translation, notification, theme, and storage settings.

## Install

On Arch Linux, install the AUR package:

```sh
yay -S ymn-git
```

Then launch the GUI from your app launcher or run:

```sh
ymn-guiv2
```

You can also launch it through the main binary:

```sh
ymn guiv2
```

## Required Packages

The Arch package declares these runtime dependencies:

- `ffmpeg`
- `glibc`
- `gst-plugins-base-libs`
- `gstreamer`
- `libx11`
- `libxkbcommon`
- `mesa`
- `mpv`
- `wayland`
- `wine`

For manual source builds, you also need:

- `go`
- `git`
- `pkgconf`

## Optional Packages

Install these only for the workflows you use:

- `ollama`: local translation backend.
- `steam`: Steam/Proton launch flows.
- `gamescope`: Gamescope runner configs.
- `winetricks`: automatic or configured Wine dependency installs.
- `tesseract`: legacy OCR workflow.
- `grim`: Wayland screenshot capture for legacy OCR.
- `slurp`: Wayland region selection for legacy OCR.
- Anki with AnkiConnect: flashcard sync from Yomuna to Anki.

## How To Use

1. Open `ymn-guiv2`.
2. Go to **Add Game** and install or create a game config. Enable text hook installation when the game supports it.
3. Use **Game** to review and adjust Wine, Gamescope, runner, and winetricks settings for that game.
4. Use **Transcript** to select the game, launch it, choose the active hook if needed, and read captured text.
5. Select transcript text or sentence tokens to inspect vocabulary and create flashcards.
6. Use **Flashcards** to search, edit, delete, page through, and sync saved cards to Anki.
7. Use **Settings** to configure translation language, Ollama settings, notification level, theme, font sizes, and storage locations.

## Translation

Yomuna currently defaults to an Ollama translation backend. The default endpoint is:

```text
http://localhost:11434
```

The default model is:

```text
translategemma:4b
```

You can change the model and endpoint from **Settings**. Make sure Ollama is running and the selected model is installed before using auto-translate.

## Anki Sync

Flashcards can be synced to Anki through AnkiConnect at:

```text
http://127.0.0.1:8765
```

Start Anki, enable the AnkiConnect add-on, select a game in Yomuna, then press **Sync Anki** from the Flashcards page.

## Storage

Yomuna stores app settings, transcript settings, translator settings, flashcards, exports, translations, dictionary data, voices, and game configs under your user config/data directories. The **Settings** page shows the exact paths and current size of each storage item.

## Development

Development, source-build, packaging, and AUR maintenance notes live in [DEVELOPMENT.md](./DEVELOPMENT.md).
