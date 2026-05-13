package theme

var DefaultThemes = []*Theme{
	&SlateCalm,
	&ForestFocus,
	&VioletProfessional,
	&AmberPaper,

	&OceanTrust,
	&GraphiteMinimal,
	&MintClinical,
	&IndigoCommand,

	&CatppuccinLatteMocha,
	&CatppuccinFrappe,
	&CatppuccinMacchiato,
	&CatppuccinMocha,

	&DraculaClassic,
	&NordArctic,
	&TokyoNight,
	&GruvboxRetro,
	&SolarizedClassic,
	&RosePine,
}
var SlateCalm = Theme{
	Name:     "Slate Calm",
	IconName: "lucide:layers",
	LightColors: ColorTokens{
		Background: "#F8FAFC",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#F1F5F9",
		Border:     "#CBD5E1",
		Divider:    "#E2E8F0",

		TextPrimary:   "#111827",
		TextSecondary: "#334155",
		TextMuted:     "#64748B",
		TextInverse:   "#F8FAFC",

		Primary:      "#2563EB",
		PrimaryHover: "#1D4ED8",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#475569",
		SecondaryHover: "#334155",
		OnSecondary:    "#FFFFFF",

		Success: "#047857",
		Warning: "#B45309",
		Danger:  "#B91C1C",
		Info:    "#0369A1",

		FocusRing: "#2563EB",
		Selection: "#DBEAFE",
		Disabled:  "#94A3B8",
	},
	DarkColors: ColorTokens{
		Background: "#0B1020",
		Surface:    "#111827",
		SurfaceAlt: "#1E293B",
		Border:     "#334155",
		Divider:    "#1F2937",

		TextPrimary:   "#E5E7EB",
		TextSecondary: "#CBD5E1",
		TextMuted:     "#94A3B8",
		TextInverse:   "#111827",

		Primary:      "#93C5FD",
		PrimaryHover: "#BFDBFE",
		OnPrimary:    "#0B1020",

		Secondary:      "#CBD5E1",
		SecondaryHover: "#E2E8F0",
		OnSecondary:    "#0B1020",

		Success: "#6EE7B7",
		Warning: "#FCD34D",
		Danger:  "#FCA5A5",
		Info:    "#7DD3FC",

		FocusRing: "#93C5FD",
		Selection: "#1E3A8A",
		Disabled:  "#64748B",
	},
}

var ForestFocus = Theme{
	Name:     "Forest Focus",
	IconName: "lucide:leaf",
	LightColors: ColorTokens{
		Background: "#F6FBF8",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#ECFDF5",
		Border:     "#A7F3D0",
		Divider:    "#D1FAE5",

		TextPrimary:   "#102A23",
		TextSecondary: "#28584C",
		TextMuted:     "#5F7F74",
		TextInverse:   "#F6FBF8",

		Primary:      "#047857",
		PrimaryHover: "#065F46",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#0F766E",
		SecondaryHover: "#115E59",
		OnSecondary:    "#FFFFFF",

		Success: "#047857",
		Warning: "#A16207",
		Danger:  "#B91C1C",
		Info:    "#0369A1",

		FocusRing: "#10B981",
		Selection: "#D1FAE5",
		Disabled:  "#8FAFA4",
	},
	DarkColors: ColorTokens{
		Background: "#081C15",
		Surface:    "#0F2A21",
		SurfaceAlt: "#15382D",
		Border:     "#276749",
		Divider:    "#1E3A2F",

		TextPrimary:   "#E6F4EF",
		TextSecondary: "#B7D8CC",
		TextMuted:     "#8FB8AA",
		TextInverse:   "#081C15",

		Primary:      "#6EE7B7",
		PrimaryHover: "#A7F3D0",
		OnPrimary:    "#081C15",

		Secondary:      "#5EEAD4",
		SecondaryHover: "#99F6E4",
		OnSecondary:    "#081C15",

		Success: "#86EFAC",
		Warning: "#FDE68A",
		Danger:  "#FCA5A5",
		Info:    "#7DD3FC",

		FocusRing: "#6EE7B7",
		Selection: "#14532D",
		Disabled:  "#5F7F74",
	},
}

var AmberPaper = Theme{
	Name:     "Amber Paper",
	IconName: "lucide:file-text",
	LightColors: ColorTokens{
		Background: "#FFF7ED",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#FFEDD5",
		Border:     "#FED7AA",
		Divider:    "#FDBA74",

		TextPrimary:   "#1C0F08",
		TextSecondary: "#7C2D12",
		TextMuted:     "#9A6B4F",
		TextInverse:   "#FFF7ED",

		Primary:      "#C2410C",
		PrimaryHover: "#9A3412",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#92400E",
		SecondaryHover: "#78350F",
		OnSecondary:    "#FFFFFF",

		Success: "#047857",
		Warning: "#B45309",
		Danger:  "#B91C1C",
		Info:    "#0369A1",

		FocusRing: "#EA580C",
		Selection: "#FED7AA",
		Disabled:  "#B99A85",
	},
	DarkColors: ColorTokens{
		Background: "#1C0F08",
		Surface:    "#2A160C",
		SurfaceAlt: "#3A2112",
		Border:     "#7C2D12",
		Divider:    "#4A2B18",

		TextPrimary:   "#FDEFE3",
		TextSecondary: "#FED7AA",
		TextMuted:     "#FDBA74",
		TextInverse:   "#1C0F08",

		Primary:      "#FDBA74",
		PrimaryHover: "#FED7AA",
		OnPrimary:    "#1C0F08",

		Secondary:      "#FCD34D",
		SecondaryHover: "#FDE68A",
		OnSecondary:    "#1C0F08",

		Success: "#86EFAC",
		Warning: "#FDE68A",
		Danger:  "#FCA5A5",
		Info:    "#7DD3FC",

		FocusRing: "#FDBA74",
		Selection: "#7C2D12",
		Disabled:  "#9A6B4F",
	},
}

var VioletProfessional = Theme{
	Name: "Violet Professional",
	LightColors: ColorTokens{
		Background: "#FAF5FF",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#F3E8FF",
		Border:     "#DDD6FE",
		Divider:    "#E9D5FF",

		TextPrimary:   "#1B1026",
		TextSecondary: "#4C1D95",
		TextMuted:     "#7E6A9E",
		TextInverse:   "#FAF5FF",

		Primary:      "#7C3AED",
		PrimaryHover: "#6D28D9",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#581C87",
		SecondaryHover: "#4C1D95",
		OnSecondary:    "#FFFFFF",

		Success: "#047857",
		Warning: "#B45309",
		Danger:  "#B91C1C",
		Info:    "#0369A1",

		FocusRing: "#8B5CF6",
		Selection: "#E9D5FF",
		Disabled:  "#A78BFA",
	},
	DarkColors: ColorTokens{
		Background: "#16051F",
		Surface:    "#22102F",
		SurfaceAlt: "#2E1740",
		Border:     "#4C1D95",
		Divider:    "#3B1A52",

		TextPrimary:   "#F4EAFE",
		TextSecondary: "#E9D5FF",
		TextMuted:     "#C4B5FD",
		TextInverse:   "#16051F",

		Primary:      "#D8B4FE",
		PrimaryHover: "#E9D5FF",
		OnPrimary:    "#16051F",

		Secondary:      "#C4B5FD",
		SecondaryHover: "#DDD6FE",
		OnSecondary:    "#16051F",

		Success: "#86EFAC",
		Warning: "#FDE68A",
		Danger:  "#FCA5A5",
		Info:    "#7DD3FC",

		FocusRing: "#D8B4FE",
		Selection: "#581C87",
		Disabled:  "#7E6A9E",
	},
	IconName: "lucide:sparkles",
}

var OceanTrust = Theme{
	Name:     "Ocean Trust",
	IconName: "lucide:waves",
	LightColors: ColorTokens{
		Background: "#F0F9FF",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#E0F2FE",
		Border:     "#BAE6FD",
		Divider:    "#D7EEF9",

		TextPrimary:   "#082F49",
		TextSecondary: "#075985",
		TextMuted:     "#64748B",
		TextInverse:   "#F0F9FF",

		Primary:      "#0284C7",
		PrimaryHover: "#0369A1",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#0E7490",
		SecondaryHover: "#155E75",
		OnSecondary:    "#FFFFFF",

		Success: "#047857",
		Warning: "#B45309",
		Danger:  "#B91C1C",
		Info:    "#0369A1",

		FocusRing: "#0EA5E9",
		Selection: "#BAE6FD",
		Disabled:  "#94A3B8",
	},
	DarkColors: ColorTokens{
		Background: "#061826",
		Surface:    "#0B2538",
		SurfaceAlt: "#12364D",
		Border:     "#075985",
		Divider:    "#164E63",

		TextPrimary:   "#E0F2FE",
		TextSecondary: "#BAE6FD",
		TextMuted:     "#7DD3FC",
		TextInverse:   "#061826",

		Primary:      "#7DD3FC",
		PrimaryHover: "#BAE6FD",
		OnPrimary:    "#061826",

		Secondary:      "#67E8F9",
		SecondaryHover: "#A5F3FC",
		OnSecondary:    "#061826",

		Success: "#86EFAC",
		Warning: "#FDE68A",
		Danger:  "#FCA5A5",
		Info:    "#7DD3FC",

		FocusRing: "#7DD3FC",
		Selection: "#075985",
		Disabled:  "#64748B",
	},
}

var RoseEditorial = Theme{
	Name:     "Rose Editorial",
	IconName: "lucide:pen-line",
	LightColors: ColorTokens{
		Background: "#FFF1F2",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#FFE4E6",
		Border:     "#FECDD3",
		Divider:    "#FFE4E6",

		TextPrimary:   "#3F0A16",
		TextSecondary: "#9F1239",
		TextMuted:     "#8E5B65",
		TextInverse:   "#FFF1F2",

		Primary:      "#E11D48",
		PrimaryHover: "#BE123C",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#9F1239",
		SecondaryHover: "#881337",
		OnSecondary:    "#FFFFFF",

		Success: "#047857",
		Warning: "#B45309",
		Danger:  "#B91C1C",
		Info:    "#0369A1",

		FocusRing: "#F43F5E",
		Selection: "#FECDD3",
		Disabled:  "#B99AA1",
	},
	DarkColors: ColorTokens{
		Background: "#20070D",
		Surface:    "#2D0D16",
		SurfaceAlt: "#3F111D",
		Border:     "#881337",
		Divider:    "#4C1622",

		TextPrimary:   "#FFE4E6",
		TextSecondary: "#FECDD3",
		TextMuted:     "#FDA4AF",
		TextInverse:   "#20070D",

		Primary:      "#FDA4AF",
		PrimaryHover: "#FECDD3",
		OnPrimary:    "#20070D",

		Secondary:      "#FB7185",
		SecondaryHover: "#FDA4AF",
		OnSecondary:    "#20070D",

		Success: "#86EFAC",
		Warning: "#FDE68A",
		Danger:  "#FCA5A5",
		Info:    "#7DD3FC",

		FocusRing: "#FDA4AF",
		Selection: "#881337",
		Disabled:  "#8E5B65",
	},
}

var GraphiteMinimal = Theme{
	Name:     "Graphite Minimal",
	IconName: "lucide:square",
	LightColors: ColorTokens{
		Background: "#FAFAFA",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#F4F4F5",
		Border:     "#D4D4D8",
		Divider:    "#E4E4E7",

		TextPrimary:   "#18181B",
		TextSecondary: "#3F3F46",
		TextMuted:     "#71717A",
		TextInverse:   "#FAFAFA",

		Primary:      "#27272A",
		PrimaryHover: "#18181B",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#52525B",
		SecondaryHover: "#3F3F46",
		OnSecondary:    "#FFFFFF",

		Success: "#047857",
		Warning: "#B45309",
		Danger:  "#B91C1C",
		Info:    "#0369A1",

		FocusRing: "#52525B",
		Selection: "#E4E4E7",
		Disabled:  "#A1A1AA",
	},
	DarkColors: ColorTokens{
		Background: "#09090B",
		Surface:    "#18181B",
		SurfaceAlt: "#27272A",
		Border:     "#3F3F46",
		Divider:    "#27272A",

		TextPrimary:   "#FAFAFA",
		TextSecondary: "#D4D4D8",
		TextMuted:     "#A1A1AA",
		TextInverse:   "#09090B",

		Primary:      "#E4E4E7",
		PrimaryHover: "#FAFAFA",
		OnPrimary:    "#09090B",

		Secondary:      "#A1A1AA",
		SecondaryHover: "#D4D4D8",
		OnSecondary:    "#09090B",

		Success: "#86EFAC",
		Warning: "#FDE68A",
		Danger:  "#FCA5A5",
		Info:    "#7DD3FC",

		FocusRing: "#E4E4E7",
		Selection: "#3F3F46",
		Disabled:  "#71717A",
	},
}

var MintClinical = Theme{
	Name:     "Mint Clinical",
	IconName: "lucide:cross",
	LightColors: ColorTokens{
		Background: "#F0FDFA",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#CCFBF1",
		Border:     "#99F6E4",
		Divider:    "#CCFBF1",

		TextPrimary:   "#042F2E",
		TextSecondary: "#115E59",
		TextMuted:     "#5F7F7A",
		TextInverse:   "#F0FDFA",

		Primary:      "#0F766E",
		PrimaryHover: "#115E59",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#047857",
		SecondaryHover: "#065F46",
		OnSecondary:    "#FFFFFF",

		Success: "#047857",
		Warning: "#B45309",
		Danger:  "#B91C1C",
		Info:    "#0369A1",

		FocusRing: "#14B8A6",
		Selection: "#99F6E4",
		Disabled:  "#94A3B8",
	},
	DarkColors: ColorTokens{
		Background: "#042F2E",
		Surface:    "#0B3B39",
		SurfaceAlt: "#134E4A",
		Border:     "#0F766E",
		Divider:    "#115E59",

		TextPrimary:   "#F0FDFA",
		TextSecondary: "#CCFBF1",
		TextMuted:     "#99F6E4",
		TextInverse:   "#042F2E",

		Primary:      "#5EEAD4",
		PrimaryHover: "#99F6E4",
		OnPrimary:    "#042F2E",

		Secondary:      "#6EE7B7",
		SecondaryHover: "#A7F3D0",
		OnSecondary:    "#042F2E",

		Success: "#86EFAC",
		Warning: "#FDE68A",
		Danger:  "#FCA5A5",
		Info:    "#7DD3FC",

		FocusRing: "#5EEAD4",
		Selection: "#0F766E",
		Disabled:  "#5F7F7A",
	},
}

var IndigoCommand = Theme{
	Name:     "Indigo Command",
	IconName: "lucide:terminal",
	LightColors: ColorTokens{
		Background: "#EEF2FF",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#E0E7FF",
		Border:     "#C7D2FE",
		Divider:    "#E0E7FF",

		TextPrimary:   "#111827",
		TextSecondary: "#3730A3",
		TextMuted:     "#6B7280",
		TextInverse:   "#EEF2FF",

		Primary:      "#4F46E5",
		PrimaryHover: "#4338CA",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#312E81",
		SecondaryHover: "#1E1B4B",
		OnSecondary:    "#FFFFFF",

		Success: "#047857",
		Warning: "#B45309",
		Danger:  "#B91C1C",
		Info:    "#0369A1",

		FocusRing: "#6366F1",
		Selection: "#C7D2FE",
		Disabled:  "#A5B4FC",
	},
	DarkColors: ColorTokens{
		Background: "#0F1028",
		Surface:    "#181A3A",
		SurfaceAlt: "#242654",
		Border:     "#3730A3",
		Divider:    "#25285A",

		TextPrimary:   "#EEF2FF",
		TextSecondary: "#C7D2FE",
		TextMuted:     "#A5B4FC",
		TextInverse:   "#0F1028",

		Primary:      "#A5B4FC",
		PrimaryHover: "#C7D2FE",
		OnPrimary:    "#0F1028",

		Secondary:      "#818CF8",
		SecondaryHover: "#A5B4FC",
		OnSecondary:    "#0F1028",

		Success: "#86EFAC",
		Warning: "#FDE68A",
		Danger:  "#FCA5A5",
		Info:    "#7DD3FC",

		FocusRing: "#A5B4FC",
		Selection: "#3730A3",
		Disabled:  "#6B7280",
	},
}

var SandstoneWarm = Theme{
	Name:     "Sandstone Warm",
	IconName: "lucide:landmark",
	LightColors: ColorTokens{
		Background: "#FFFBEB",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#FEF3C7",
		Border:     "#FDE68A",
		Divider:    "#FDE68A",

		TextPrimary:   "#1C1917",
		TextSecondary: "#78350F",
		TextMuted:     "#8A6A42",
		TextInverse:   "#FFFBEB",

		Primary:      "#B45309",
		PrimaryHover: "#92400E",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#854D0E",
		SecondaryHover: "#713F12",
		OnSecondary:    "#FFFFFF",

		Success: "#047857",
		Warning: "#B45309",
		Danger:  "#B91C1C",
		Info:    "#0369A1",

		FocusRing: "#D97706",
		Selection: "#FDE68A",
		Disabled:  "#B99A6B",
	},
	DarkColors: ColorTokens{
		Background: "#1C1308",
		Surface:    "#2B1D0E",
		SurfaceAlt: "#3A2A14",
		Border:     "#854D0E",
		Divider:    "#4A3216",

		TextPrimary:   "#FEF3C7",
		TextSecondary: "#FDE68A",
		TextMuted:     "#FCD34D",
		TextInverse:   "#1C1308",

		Primary:      "#FCD34D",
		PrimaryHover: "#FDE68A",
		OnPrimary:    "#1C1308",

		Secondary:      "#FDBA74",
		SecondaryHover: "#FED7AA",
		OnSecondary:    "#1C1308",

		Success: "#86EFAC",
		Warning: "#FDE68A",
		Danger:  "#FCA5A5",
		Info:    "#7DD3FC",

		FocusRing: "#FCD34D",
		Selection: "#854D0E",
		Disabled:  "#8A6A42",
	},
}

var CrimsonDashboard = Theme{
	Name:     "Crimson Dashboard",
	IconName: "lucide:gauge",
	LightColors: ColorTokens{
		Background: "#FEF2F2",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#FEE2E2",
		Border:     "#FECACA",
		Divider:    "#FEE2E2",

		TextPrimary:   "#2A0B0B",
		TextSecondary: "#991B1B",
		TextMuted:     "#8C5A5A",
		TextInverse:   "#FEF2F2",

		Primary:      "#DC2626",
		PrimaryHover: "#B91C1C",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#7F1D1D",
		SecondaryHover: "#641616",
		OnSecondary:    "#FFFFFF",

		Success: "#047857",
		Warning: "#B45309",
		Danger:  "#B91C1C",
		Info:    "#0369A1",

		FocusRing: "#EF4444",
		Selection: "#FECACA",
		Disabled:  "#B99A9A",
	},
	DarkColors: ColorTokens{
		Background: "#1F0808",
		Surface:    "#2A0B0B",
		SurfaceAlt: "#3B1010",
		Border:     "#7F1D1D",
		Divider:    "#4B1616",

		TextPrimary:   "#FEE2E2",
		TextSecondary: "#FECACA",
		TextMuted:     "#FCA5A5",
		TextInverse:   "#1F0808",

		Primary:      "#FCA5A5",
		PrimaryHover: "#FECACA",
		OnPrimary:    "#1F0808",

		Secondary:      "#F87171",
		SecondaryHover: "#FCA5A5",
		OnSecondary:    "#1F0808",

		Success: "#86EFAC",
		Warning: "#FDE68A",
		Danger:  "#FCA5A5",
		Info:    "#7DD3FC",

		FocusRing: "#FCA5A5",
		Selection: "#7F1D1D",
		Disabled:  "#8C5A5A",
	},
}

var CyberLime = Theme{
	Name:     "Cyber Lime",
	IconName: "lucide:zap",
	LightColors: ColorTokens{
		Background: "#F7FEE7",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#ECFCCB",
		Border:     "#D9F99D",
		Divider:    "#ECFCCB",

		TextPrimary:   "#1A2E05",
		TextSecondary: "#3F6212",
		TextMuted:     "#657A3A",
		TextInverse:   "#F7FEE7",

		Primary:      "#4D7C0F",
		PrimaryHover: "#3F6212",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#365314",
		SecondaryHover: "#1A2E05",
		OnSecondary:    "#FFFFFF",

		Success: "#047857",
		Warning: "#B45309",
		Danger:  "#B91C1C",
		Info:    "#0369A1",

		FocusRing: "#65A30D",
		Selection: "#D9F99D",
		Disabled:  "#A3B37A",
	},
	DarkColors: ColorTokens{
		Background: "#0F1A05",
		Surface:    "#172408",
		SurfaceAlt: "#22340D",
		Border:     "#3F6212",
		Divider:    "#2D4710",

		TextPrimary:   "#F7FEE7",
		TextSecondary: "#D9F99D",
		TextMuted:     "#BEF264",
		TextInverse:   "#0F1A05",

		Primary:      "#BEF264",
		PrimaryHover: "#D9F99D",
		OnPrimary:    "#0F1A05",

		Secondary:      "#A3E635",
		SecondaryHover: "#BEF264",
		OnSecondary:    "#0F1A05",

		Success: "#86EFAC",
		Warning: "#FDE68A",
		Danger:  "#FCA5A5",
		Info:    "#7DD3FC",

		FocusRing: "#BEF264",
		Selection: "#3F6212",
		Disabled:  "#657A3A",
	},
}

var CatppuccinLatteMocha = Theme{
	Name:     "Catppuccin Latte Mocha",
	IconName: "lucide:coffee",
	LightColors: ColorTokens{
		Background: "#EFF1F5",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#E6E9EF",
		Border:     "#BCC0CC",
		Divider:    "#DCE0E8",

		TextPrimary:   "#4C4F69",
		TextSecondary: "#5C5F77",
		TextMuted:     "#8C8FA1",
		TextInverse:   "#EFF1F5",

		Primary:      "#8839EF",
		PrimaryHover: "#7287FD",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#1E66F5",
		SecondaryHover: "#209FB5",
		OnSecondary:    "#FFFFFF",

		Success: "#40A02B",
		Warning: "#DF8E1D",
		Danger:  "#D20F39",
		Info:    "#04A5E5",

		FocusRing: "#8839EF",
		Selection: "#DCE0E8",
		Disabled:  "#9CA0B0",
	},
	DarkColors: ColorTokens{
		Background: "#1E1E2E",
		Surface:    "#313244",
		SurfaceAlt: "#45475A",
		Border:     "#6C7086",
		Divider:    "#585B70",

		TextPrimary:   "#CDD6F4",
		TextSecondary: "#BAC2DE",
		TextMuted:     "#A6ADC8",
		TextInverse:   "#1E1E2E",

		Primary:      "#CBA6F7",
		PrimaryHover: "#B4BEFE",
		OnPrimary:    "#1E1E2E",

		Secondary:      "#89B4FA",
		SecondaryHover: "#74C7EC",
		OnSecondary:    "#1E1E2E",

		Success: "#A6E3A1",
		Warning: "#F9E2AF",
		Danger:  "#F38BA8",
		Info:    "#89DCEB",

		FocusRing: "#CBA6F7",
		Selection: "#45475A",
		Disabled:  "#6C7086",
	},
}

var CatppuccinFrappe = Theme{
	Name:     "Catppuccin Frappe",
	IconName: "lucide:cup-soda",
	LightColors: ColorTokens{
		Background: "#F2F4F8",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#E6E9EF",
		Border:     "#C6CCDC",
		Divider:    "#DCE0E8",

		TextPrimary:   "#303446",
		TextSecondary: "#51576D",
		TextMuted:     "#737994",
		TextInverse:   "#F2F4F8",

		Primary:      "#8839EF",
		PrimaryHover: "#7287FD",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#1E66F5",
		SecondaryHover: "#209FB5",
		OnSecondary:    "#FFFFFF",

		Success: "#40A02B",
		Warning: "#DF8E1D",
		Danger:  "#D20F39",
		Info:    "#04A5E5",

		FocusRing: "#8839EF",
		Selection: "#DCE0E8",
		Disabled:  "#9CA0B0",
	},
	DarkColors: ColorTokens{
		Background: "#303446",
		Surface:    "#414559",
		SurfaceAlt: "#51576D",
		Border:     "#737994",
		Divider:    "#626880",

		TextPrimary:   "#C6D0F5",
		TextSecondary: "#B5BFE2",
		TextMuted:     "#A5ADCE",
		TextInverse:   "#303446",

		Primary:      "#CA9EE6",
		PrimaryHover: "#BABBF1",
		OnPrimary:    "#303446",

		Secondary:      "#8CAAEE",
		SecondaryHover: "#85C1DC",
		OnSecondary:    "#303446",

		Success: "#A6D189",
		Warning: "#E5C890",
		Danger:  "#E78284",
		Info:    "#99D1DB",

		FocusRing: "#CA9EE6",
		Selection: "#51576D",
		Disabled:  "#737994",
	},
}

var CatppuccinMacchiato = Theme{
	Name:     "Catppuccin Macchiato",
	IconName: "lucide:milk",
	LightColors: ColorTokens{
		Background: "#F4F6FB",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#E6E9EF",
		Border:     "#CAD0E0",
		Divider:    "#DCE0E8",

		TextPrimary:   "#24273A",
		TextSecondary: "#494D64",
		TextMuted:     "#6E738D",
		TextInverse:   "#F4F6FB",

		Primary:      "#8839EF",
		PrimaryHover: "#7287FD",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#1E66F5",
		SecondaryHover: "#209FB5",
		OnSecondary:    "#FFFFFF",

		Success: "#40A02B",
		Warning: "#DF8E1D",
		Danger:  "#D20F39",
		Info:    "#04A5E5",

		FocusRing: "#8839EF",
		Selection: "#DCE0E8",
		Disabled:  "#9CA0B0",
	},
	DarkColors: ColorTokens{
		Background: "#24273A",
		Surface:    "#363A4F",
		SurfaceAlt: "#494D64",
		Border:     "#6E738D",
		Divider:    "#5B6078",

		TextPrimary:   "#CAD3F5",
		TextSecondary: "#B8C0E0",
		TextMuted:     "#A5ADCB",
		TextInverse:   "#24273A",

		Primary:      "#C6A0F6",
		PrimaryHover: "#B7BDF8",
		OnPrimary:    "#24273A",

		Secondary:      "#8AADF4",
		SecondaryHover: "#7DC4E4",
		OnSecondary:    "#24273A",

		Success: "#A6DA95",
		Warning: "#EED49F",
		Danger:  "#ED8796",
		Info:    "#91D7E3",

		FocusRing: "#C6A0F6",
		Selection: "#494D64",
		Disabled:  "#6E738D",
	},
}

var CatppuccinMocha = Theme{
	Name:     "Catppuccin Mocha",
	IconName: "lucide:moon",
	LightColors: ColorTokens{
		Background: "#EFF1F5",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#E6E9EF",
		Border:     "#BCC0CC",
		Divider:    "#DCE0E8",

		TextPrimary:   "#4C4F69",
		TextSecondary: "#5C5F77",
		TextMuted:     "#8C8FA1",
		TextInverse:   "#EFF1F5",

		Primary:      "#8839EF",
		PrimaryHover: "#7287FD",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#1E66F5",
		SecondaryHover: "#209FB5",
		OnSecondary:    "#FFFFFF",

		Success: "#40A02B",
		Warning: "#DF8E1D",
		Danger:  "#D20F39",
		Info:    "#04A5E5",

		FocusRing: "#8839EF",
		Selection: "#DCE0E8",
		Disabled:  "#9CA0B0",
	},
	DarkColors: ColorTokens{
		Background: "#11111B",
		Surface:    "#1E1E2E",
		SurfaceAlt: "#313244",
		Border:     "#585B70",
		Divider:    "#45475A",

		TextPrimary:   "#CDD6F4",
		TextSecondary: "#BAC2DE",
		TextMuted:     "#A6ADC8",
		TextInverse:   "#11111B",

		Primary:      "#CBA6F7",
		PrimaryHover: "#B4BEFE",
		OnPrimary:    "#11111B",

		Secondary:      "#89B4FA",
		SecondaryHover: "#74C7EC",
		OnSecondary:    "#11111B",

		Success: "#A6E3A1",
		Warning: "#F9E2AF",
		Danger:  "#F38BA8",
		Info:    "#89DCEB",

		FocusRing: "#CBA6F7",
		Selection: "#313244",
		Disabled:  "#6C7086",
	},
}

var DraculaClassic = Theme{
	Name:     "Dracula Classic",
	IconName: "lucide:badge",
	LightColors: ColorTokens{
		Background: "#F8F8F2",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#EDEDF4",
		Border:     "#C9CBDD",
		Divider:    "#E1E2EC",

		TextPrimary:   "#282A36",
		TextSecondary: "#44475A",
		TextMuted:     "#6272A4",
		TextInverse:   "#F8F8F2",

		Primary:      "#7C3AED",
		PrimaryHover: "#6D28D9",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#DB2777",
		SecondaryHover: "#BE185D",
		OnSecondary:    "#FFFFFF",

		Success: "#047857",
		Warning: "#B45309",
		Danger:  "#B91C1C",
		Info:    "#0369A1",

		FocusRing: "#BD93F9",
		Selection: "#E9D5FF",
		Disabled:  "#9CA3AF",
	},
	DarkColors: ColorTokens{
		Background: "#282A36",
		Surface:    "#343746",
		SurfaceAlt: "#44475A",
		Border:     "#6272A4",
		Divider:    "#3B3E4E",

		TextPrimary:   "#F8F8F2",
		TextSecondary: "#E9E9E3",
		TextMuted:     "#C7C9D1",
		TextInverse:   "#282A36",

		Primary:      "#BD93F9",
		PrimaryHover: "#D6B8FF",
		OnPrimary:    "#282A36",

		Secondary:      "#FF79C6",
		SecondaryHover: "#FF92D0",
		OnSecondary:    "#282A36",

		Success: "#50FA7B",
		Warning: "#F1FA8C",
		Danger:  "#FF5555",
		Info:    "#8BE9FD",

		FocusRing: "#BD93F9",
		Selection: "#44475A",
		Disabled:  "#6272A4",
	},
}

var NordArctic = Theme{
	Name:     "Nord Arctic",
	IconName: "lucide:snowflake",
	LightColors: ColorTokens{
		Background: "#ECEFF4",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#E5E9F0",
		Border:     "#D8DEE9",
		Divider:    "#E5E9F0",

		TextPrimary:   "#2E3440",
		TextSecondary: "#3B4252",
		TextMuted:     "#4C566A",
		TextInverse:   "#ECEFF4",

		Primary:      "#5E81AC",
		PrimaryHover: "#4C719D",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#2E6F7E",
		SecondaryHover: "#255E6B",
		OnSecondary:    "#FFFFFF",

		Success: "#047857",
		Warning: "#B45309",
		Danger:  "#B91C1C",
		Info:    "#0369A1",

		FocusRing: "#5E81AC",
		Selection: "#D8DEE9",
		Disabled:  "#8D96A8",
	},
	DarkColors: ColorTokens{
		Background: "#2E3440",
		Surface:    "#3B4252",
		SurfaceAlt: "#434C5E",
		Border:     "#4C566A",
		Divider:    "#434C5E",

		TextPrimary:   "#ECEFF4",
		TextSecondary: "#E5E9F0",
		TextMuted:     "#D8DEE9",
		TextInverse:   "#2E3440",

		Primary:      "#88C0D0",
		PrimaryHover: "#8FBCBB",
		OnPrimary:    "#2E3440",

		Secondary:      "#81A1C1",
		SecondaryHover: "#88C0D0",
		OnSecondary:    "#2E3440",

		Success: "#A3BE8C",
		Warning: "#EBCB8B",
		Danger:  "#BF616A",
		Info:    "#88C0D0",

		FocusRing: "#88C0D0",
		Selection: "#4C566A",
		Disabled:  "#6B7280",
	},
}

var TokyoNight = Theme{
	Name:     "Tokyo Night",
	IconName: "lucide:building-2",
	LightColors: ColorTokens{
		Background: "#F7F8FC",
		Surface:    "#FFFFFF",
		SurfaceAlt: "#EDEFF7",
		Border:     "#C8D3F5",
		Divider:    "#E1E5F2",

		TextPrimary:   "#1A1B26",
		TextSecondary: "#414868",
		TextMuted:     "#6B7280",
		TextInverse:   "#F7F8FC",

		Primary:      "#3D59A1",
		PrimaryHover: "#2F467F",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#7C3AED",
		SecondaryHover: "#6D28D9",
		OnSecondary:    "#FFFFFF",

		Success: "#047857",
		Warning: "#B45309",
		Danger:  "#B91C1C",
		Info:    "#0369A1",

		FocusRing: "#7AA2F7",
		Selection: "#C8D3F5",
		Disabled:  "#9AA5CE",
	},
	DarkColors: ColorTokens{
		Background: "#1A1B26",
		Surface:    "#24283B",
		SurfaceAlt: "#292E42",
		Border:     "#3B4261",
		Divider:    "#32384F",

		TextPrimary:   "#C0CAF5",
		TextSecondary: "#A9B1D6",
		TextMuted:     "#9AA5CE",
		TextInverse:   "#1A1B26",

		Primary:      "#7AA2F7",
		PrimaryHover: "#9ABDFC",
		OnPrimary:    "#1A1B26",

		Secondary:      "#BB9AF7",
		SecondaryHover: "#D2B7FF",
		OnSecondary:    "#1A1B26",

		Success: "#9ECE6A",
		Warning: "#E0AF68",
		Danger:  "#F7768E",
		Info:    "#7DCFFF",

		FocusRing: "#7AA2F7",
		Selection: "#3B4261",
		Disabled:  "#565F89",
	},
}

var GruvboxRetro = Theme{
	Name:     "Gruvbox Retro",
	IconName: "lucide:radio",
	LightColors: ColorTokens{
		Background: "#FBF1C7",
		Surface:    "#FFF8D8",
		SurfaceAlt: "#EBDBB2",
		Border:     "#D5C4A1",
		Divider:    "#E5D6A8",

		TextPrimary:   "#3C3836",
		TextSecondary: "#504945",
		TextMuted:     "#7C6F64",
		TextInverse:   "#FBF1C7",

		Primary:      "#458588",
		PrimaryHover: "#076678",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#B16286",
		SecondaryHover: "#8F3F71",
		OnSecondary:    "#FFFFFF",

		Success: "#79740E",
		Warning: "#B57614",
		Danger:  "#AF3A03",
		Info:    "#076678",

		FocusRing: "#458588",
		Selection: "#D5C4A1",
		Disabled:  "#A89984",
	},
	DarkColors: ColorTokens{
		Background: "#282828",
		Surface:    "#3C3836",
		SurfaceAlt: "#504945",
		Border:     "#665C54",
		Divider:    "#3C3836",

		TextPrimary:   "#FBF1C7",
		TextSecondary: "#EBDBB2",
		TextMuted:     "#D5C4A1",
		TextInverse:   "#282828",

		Primary:      "#83A598",
		PrimaryHover: "#8EC07C",
		OnPrimary:    "#282828",

		Secondary:      "#D3869B",
		SecondaryHover: "#FB4934",
		OnSecondary:    "#282828",

		Success: "#B8BB26",
		Warning: "#FABD2F",
		Danger:  "#FB4934",
		Info:    "#83A598",

		FocusRing: "#83A598",
		Selection: "#504945",
		Disabled:  "#7C6F64",
	},
}

var SolarizedClassic = Theme{
	Name:     "Solarized Classic",
	IconName: "lucide:sun-medium",
	LightColors: ColorTokens{
		Background: "#FDF6E3",
		Surface:    "#FFFBED",
		SurfaceAlt: "#EEE8D5",
		Border:     "#D8D0B8",
		Divider:    "#EEE8D5",

		TextPrimary:   "#073642",
		TextSecondary: "#586E75",
		TextMuted:     "#657B83",
		TextInverse:   "#FDF6E3",

		Primary:      "#268BD2",
		PrimaryHover: "#2076B2",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#6C71C4",
		SecondaryHover: "#565BB0",
		OnSecondary:    "#FFFFFF",

		Success: "#859900",
		Warning: "#B58900",
		Danger:  "#DC322F",
		Info:    "#2AA198",

		FocusRing: "#268BD2",
		Selection: "#EEE8D5",
		Disabled:  "#93A1A1",
	},
	DarkColors: ColorTokens{
		Background: "#002B36",
		Surface:    "#073642",
		SurfaceAlt: "#0B4452",
		Border:     "#586E75",
		Divider:    "#16434D",

		TextPrimary:   "#FDF6E3",
		TextSecondary: "#EEE8D5",
		TextMuted:     "#93A1A1",
		TextInverse:   "#002B36",

		Primary:      "#268BD2",
		PrimaryHover: "#2AA198",
		OnPrimary:    "#002B36",

		Secondary:      "#6C71C4",
		SecondaryHover: "#B58900",
		OnSecondary:    "#002B36",

		Success: "#859900",
		Warning: "#B58900",
		Danger:  "#DC322F",
		Info:    "#2AA198",

		FocusRing: "#268BD2",
		Selection: "#073642",
		Disabled:  "#586E75",
	},
}

var RosePine = Theme{
	Name:     "Rosé Pine",
	IconName: "lucide:flower-2",
	LightColors: ColorTokens{
		Background: "#FAF4ED",
		Surface:    "#FFFDFB",
		SurfaceAlt: "#F2E9E1",
		Border:     "#DFD6CE",
		Divider:    "#EEE6DE",

		TextPrimary:   "#575279",
		TextSecondary: "#797593",
		TextMuted:     "#9893A5",
		TextInverse:   "#FAF4ED",

		Primary:      "#907AA9",
		PrimaryHover: "#7B6495",
		OnPrimary:    "#FFFFFF",

		Secondary:      "#286983",
		SecondaryHover: "#1F596F",
		OnSecondary:    "#FFFFFF",

		Success: "#56949F",
		Warning: "#B4637A",
		Danger:  "#B91C1C",
		Info:    "#286983",

		FocusRing: "#907AA9",
		Selection: "#DFD6CE",
		Disabled:  "#B8AFA8",
	},
	DarkColors: ColorTokens{
		Background: "#191724",
		Surface:    "#1F1D2E",
		SurfaceAlt: "#26233A",
		Border:     "#403D52",
		Divider:    "#2A273F",

		TextPrimary:   "#E0DEF4",
		TextSecondary: "#D6D4E8",
		TextMuted:     "#908CAA",
		TextInverse:   "#191724",

		Primary:      "#C4A7E7",
		PrimaryHover: "#D7C0F4",
		OnPrimary:    "#191724",

		Secondary:      "#9CCFD8",
		SecondaryHover: "#B8E0E6",
		OnSecondary:    "#191724",

		Success: "#9CCFD8",
		Warning: "#F6C177",
		Danger:  "#EB6F92",
		Info:    "#31748F",

		FocusRing: "#C4A7E7",
		Selection: "#403D52",
		Disabled:  "#6E6A86",
	},
}
