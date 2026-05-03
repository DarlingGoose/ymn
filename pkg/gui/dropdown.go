package gui

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	bareui "github.com/DarlingGoose/bare/pkg/ui"
	"github.com/DarlingGoose/bare/pkg/ui/icons"
	barethemes "github.com/DarlingGoose/bare/pkg/ui/themes"
)

type OptionsList struct {
	ModeOptions           []DropdownOption
	PaletteOptions        []DropdownOption
	TranscriptSizeOptions []DropdownOption
	RecentLineOptions     []DropdownOption
	NewGameRunnerOptions  []DropdownOption
}

type DropdownOption struct {
	Label           string
	Icon            string
	Mode            barethemes.Mode
	Palette         barethemes.PaletteName
	TextSize        unit.Sp
	RecentLineLimit int
	Value           string
	Clickable       *widget.Clickable
}

func NewDropDownLayout(drop *bareui.Dropdown, icon string) {
	drop.Width = unit.Dp(260)
	drop.MaxHeight = unit.Dp(320)
	drop.OffsetY = unit.Dp(52)
	drop.Prefix = icon
	return
}
func LayoutOptionMenu(gtx layout.Context, options []DropdownOption, selected string, theme barethemes.Theme, iconify *icons.Iconify) layout.Dimensions {
	children := make([]layout.FlexChild, 0, len(options))
	for i := range options {
		opt := options[i]
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			btn := bareui.Button{
				Clickable: opt.Clickable,
				Text:      opt.Label,
				Prefix:    opt.Icon,
				Variant:   dropdownButtonVariant(opt.Label == selected),
			}
			return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return btn.Layout(gtx, theme, iconify)
			})
		}))
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx, children...)
}

func dropdownButtonVariant(active bool) bareui.ButtonVariant {
	if active {
		return bareui.ButtonPrimary
	}
	return bareui.ButtonSecondary
}

func NewModeOptions() []DropdownOption {
	return []DropdownOption{
		{Label: "Dark", Icon: "mdi:weather-night", Mode: barethemes.ModeDark, Clickable: new(widget.Clickable)},
		{Label: "Light", Icon: "mdi:white-balance-sunny", Mode: barethemes.ModeLight, Clickable: new(widget.Clickable)},
	}
}

func NewPaletteOptions() []DropdownOption {
	return []DropdownOption{
		{Label: "Ocean", Icon: "mdi:waves", Palette: barethemes.PaletteOcean, Clickable: new(widget.Clickable)},
		{Label: "Sky", Icon: "mdi:weather-partly-cloudy", Palette: barethemes.PaletteSky, Clickable: new(widget.Clickable)},
		{Label: "Coastal", Icon: "mdi:beach", Palette: barethemes.PaletteCoastal, Clickable: new(widget.Clickable)},
		{Label: "Blush", Icon: "mdi:flower-outline", Palette: barethemes.PaletteBlush, Clickable: new(widget.Clickable)},
		{Label: "Sunset", Icon: "mdi:weather-sunset", Palette: barethemes.PaletteSunset, Clickable: new(widget.Clickable)},
		{Label: "Pastel", Icon: "mdi:palette-swatch-outline", Palette: barethemes.PalettePastel, Clickable: new(widget.Clickable)},
		{Label: "Warm Earth", Icon: "mdi:terrain", Palette: barethemes.PaletteWarmEarth, Clickable: new(widget.Clickable)},
		{Label: "Soft Neutral", Icon: "mdi:circle-outline", Palette: barethemes.PaletteSoftNeutral, Clickable: new(widget.Clickable)},
		{Label: "Lavender", Icon: "mdi:flower", Palette: barethemes.PaletteLavender, Clickable: new(widget.Clickable)},
		{Label: "Harvest", Icon: "mdi:corn", Palette: barethemes.PaletteHarvest, Clickable: new(widget.Clickable)},
		{Label: "Candy", Icon: "mdi:candy", Palette: barethemes.PaletteCandy, Clickable: new(widget.Clickable)},
		{Label: "Creamy Pop", Icon: "mdi:ice-cream", Palette: barethemes.PaletteCreamyPop, Clickable: new(widget.Clickable)},
		{Label: "Violet", Icon: "mdi:palette", Palette: barethemes.PaletteViolet, Clickable: new(widget.Clickable)},
		{Label: "Forest Pop", Icon: "mdi:pine-tree", Palette: barethemes.PaletteForestPop, Clickable: new(widget.Clickable)},
		{Label: "Dark Accent", Icon: "mdi:weather-night", Palette: barethemes.PaletteDarkAccent, Clickable: new(widget.Clickable)},
		{Label: "Retro", Icon: "mdi:record-circle-outline", Palette: barethemes.PaletteRetro, Clickable: new(widget.Clickable)},
	}
}

func NewTranscriptSizeOptions() []DropdownOption {
	return []DropdownOption{
		{Label: "Small", Icon: "mdi:format-font-size-decrease", TextSize: unit.Sp(13), Clickable: new(widget.Clickable)},
		{Label: "Medium", Icon: "mdi:format-size", TextSize: unit.Sp(16), Clickable: new(widget.Clickable)},
		{Label: "Large", Icon: "mdi:format-font-size-increase", TextSize: unit.Sp(20), Clickable: new(widget.Clickable)},
		{Label: "XL", Icon: "mdi:format-letter-case", TextSize: unit.Sp(24), Clickable: new(widget.Clickable)},
		{Label: "XXL", Icon: "mdi:format-letter-case", TextSize: unit.Sp(32), Clickable: new(widget.Clickable)},
	}
}

func NewRecentLineOptions() []DropdownOption {
	return []DropdownOption{
		{Label: "All Lines", Icon: "mdi:unfold-more-horizontal", RecentLineLimit: 0, Clickable: new(widget.Clickable)},
		{Label: "Last 50 Lines", Icon: "mdi:numeric-50-box-outline", RecentLineLimit: 50, Clickable: new(widget.Clickable)},
		{Label: "Last 100 Lines", Icon: "mdi:numeric-100-box-outline", RecentLineLimit: 100, Clickable: new(widget.Clickable)},
		{Label: "Last 200 Lines", Icon: "mdi:numeric-200-box-outline", RecentLineLimit: 200, Clickable: new(widget.Clickable)},
	}
}

func NewFuriganaModeOptions() []DropdownOption {
	return []DropdownOption{
		{Label: "Hide", Icon: "mdi:eye-off-outline", Value: "hidden", Clickable: new(widget.Clickable)},
		{Label: "Above", Icon: "mdi:format-superscript", Value: "above", Clickable: new(widget.Clickable)},
		{Label: "Below", Icon: "mdi:format-subscript", Value: "below", Clickable: new(widget.Clickable)},
	}
}

func NewGameRunnerOptions() []DropdownOption {
	return []DropdownOption{
		{Label: "Auto", Icon: "mdi:magic-staff", Clickable: new(widget.Clickable)},
		{Label: "Wine", Icon: "mdi:glass-wine", Clickable: new(widget.Clickable)},
		{Label: "Proton", Icon: "mdi:flask-outline", Clickable: new(widget.Clickable)},
		{Label: "Steam", Icon: "mdi:steam", Clickable: new(widget.Clickable)},
	}
}
