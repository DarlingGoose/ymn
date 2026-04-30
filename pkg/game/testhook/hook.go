package testhook

type Engine string

const (
	EngineRPGMakerMV Engine = "rpgmaker_mv"
	EngineRPGMakerMZ Engine = "rpgmaker_mz"
	EngineKirikiri   Engine = "kirikiri"
	EngineUnknown    Engine = "unknown"
)

type Method string

const (
	MethodMod          Method = "mod"
	MethodScriptPatch  Method = "script_patch"
	MethodExternalHook Method = "external_hook"
)

type Hook interface {
	Detect(inputPath string) (TextHookStatus, error)
	InstallHook(inputPath string) (TextHookInstallResult, error)
	IsInstalled(inputPath string) (bool, error)
}

func NewAutoHook(inputPath string) Hook {
	if _, _, err := resolveRPGMakerProjectRoot(inputPath); err == nil {
		return &RPGMakerHook{}
	}

	if _, err := resolveKirikiriProjectRoot(inputPath); err == nil {
		return &KirikiriHook{}
	}

	return &ExternalTextHook{}
}
