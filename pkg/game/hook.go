package game

type Hook interface {
	InstallHook(inputPath string) (TextHookInstallResult, error)
	IsInstalled() bool
}
