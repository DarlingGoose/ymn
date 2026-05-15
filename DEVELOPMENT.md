# Yomuna Development

This file collects source-build and maintainer notes. User-facing install and usage docs are in [README.md](./README.md).

## Build From Source

Install build dependencies:

```sh
sudo pacman -S go git pkgconf
```

Build the project:

```sh
go build ./...
```

Or build the main binary through the Makefile:

```sh
make build
```

## Local Install

Install the `ymn` binary, GUI launcher wrapper, desktop entry, and license file:

```sh![create-a-clean--modern-svg-app-icon-for--yomuna---.png](create-a-clean--modern-svg-app-icon-for--yomuna---.png)
sudo make install PREFIX=/usr
```

After installing, launch with:

```sh
ymn guiv2
ymn-guiv2
```

## Current Development Notes

- Main GUI code lives under `pkg/v2/gui` and `pkg/yomuna`.
- Reusable Gio components live under `pkg/v2/gui/core/components`.
- Theme, typography, and animation helpers live under `pkg/v2/gui/core/theme` and `pkg/v2/gui/core/animations`.
- Yomuna app pages live under `pkg/yomuna/yomunapages` and `pkg/v2/gui/pages`.

## Ideas And TODOs

- Add more reusable v2 GUI components.
- Continue improving modal, notification, tab, sidebar, scrollbar, and copy/highlight text components.
- Implement key press and keybinding helpers.
- Keep GUI-facing backend APIs easy to mock.
- Continue multi-language improvements.
