package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/DarlingGoose/gr"
	"github.com/DarlingGoose/gr/autorunner"
	"github.com/DarlingGoose/gr/gamescope"
	grinstaller "github.com/DarlingGoose/gr/installer"
	"github.com/DarlingGoose/jpndict"
	"github.com/DarlingGoose/jpndict/translate"
	"github.com/DarlingGoose/tr/pkg/textractor"
	"github.com/DarlingGoose/vntext/pkg/engine"
	"github.com/DarlingGoose/vntext/pkg/engine/auto"
	"github.com/DarlingGoose/vntext/pkg/engine/enginerun"
	"github.com/DarlingGoose/vntext/pkg/game"
	"github.com/DarlingGoose/vntext/pkg/gameConfig"
	vnutil "github.com/DarlingGoose/vntext/pkg/util"
	"github.com/DarlingGoose/ymn/pkg/japanese"
	"github.com/DarlingGoose/ymn/pkg/util"
)

var _ Backend = &LiveBackend{}

type LiveBackend struct {
	gameMu     sync.RWMutex
	games      []*game.Game
	config     TranscriptConfig
	current    *game.Game
	currentRun *gr.Process

	engineSelector *auto.EngineSelectorV2
	translatorMu   sync.Mutex
	translatorCfg  TranslatorConfig
	translator     translate.Translate

	dictMu sync.Mutex
	dict   jpndict.Dictonary

	hookHistoryMu sync.Mutex
	currentLines  []engine.Line
}

func NewLive() *LiveBackend {
	cfg := loadTranslatorConfig()
	return &LiveBackend{
		engineSelector: auto.NewEngineSelectorV2(""),
		translatorCfg:  cfg,
		translator:     newOllamaTranslator(cfg),
	}
}

func defaultTranslatorConfig() TranslatorConfig {
	return TranslatorConfig{
		OllamaModel:   "translategemma:4b",
		OllamaBaseURL: "http://localhost:11434",
	}
}

func loadTranslatorConfig() TranslatorConfig {
	cfg := defaultTranslatorConfig()
	data, err := os.ReadFile(translatorConfigPath())
	if err != nil {
		return cfg
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return defaultTranslatorConfig()
	}
	cfg.normalize()
	return cfg
}

func saveTranslatorConfig(cfg TranslatorConfig) error {
	cfg.normalize()
	if err := os.MkdirAll(filepath.Dir(translatorConfigPath()), 0o755); err != nil {
		return fmt.Errorf("create translator config dir: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode translator config: %w", err)
	}
	if err := os.WriteFile(translatorConfigPath(), append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write translator config: %w", err)
	}
	return nil
}

func translatorConfigPath() string {
	return filepath.Join(util.ConfigBaseDir(), "guiv2-translator.json")
}

func (c *TranslatorConfig) normalize() {
	if c == nil {
		return
	}
	defaults := defaultTranslatorConfig()
	c.OllamaModel = strings.TrimSpace(c.OllamaModel)
	c.OllamaBaseURL = strings.TrimRight(strings.TrimSpace(c.OllamaBaseURL), "/")
	if c.OllamaModel == "" {
		c.OllamaModel = defaults.OllamaModel
	}
	if c.OllamaBaseURL == "" {
		c.OllamaBaseURL = defaults.OllamaBaseURL
	}
}

func newOllamaTranslator(cfg TranslatorConfig) translate.Translate {
	cfg.normalize()
	return translate.NewOllamaClient(translate.OllamaConfig{
		BaseURL: cfg.OllamaBaseURL,
		Model:   cfg.OllamaModel,
	})
}

func (b *LiveBackend) IsGameRunning() bool {
	return b.currentRun != nil
}

func (b *LiveBackend) GetGames() []*game.Game {
	b.gameMu.RLock()
	if len(b.games) == 0 {
		b.gameMu.RUnlock()
		_ = b.ReloadGames()
		b.gameMu.RLock()
	}

	defer b.gameMu.RUnlock()
	return append([]*game.Game(nil), b.games...)
}

func (b *LiveBackend) SelectGameByName(n string) {
	b.gameMu.Lock()
	b.current = findGameByName(b.games, n)
	b.config.SelectGameName = n
	b.gameMu.Unlock()

}

func (b *LiveBackend) InstallGameConfig(ctx context.Context, inputPath string, installHook bool) (*game.Game, error) {
	g, _, err := gameConfig.InstallGame(ctx, inputPath, installHook, "")
	if err != nil {
		return nil, err
	}
	if err := b.ReloadGames(); err != nil {
		return nil, err
	}
	b.SelectGameByName(g.Name)
	return b.CurrentGame(), nil
}

func (b *LiveBackend) InstallGameWithInstaller(ctx context.Context, installerPath, gamePath string, installHook bool) (*game.Game, error) {
	installerPath = strings.TrimSpace(installerPath)
	gamePath = strings.TrimSpace(gamePath)
	if installerPath == "" {
		return nil, fmt.Errorf("installer path is required")
	}
	if gamePath == "" {
		return nil, fmt.Errorf("game path is required")
	}

	preparedInstallerPath, preparedGamePath, err := prepareInstallerPaths(ctx, installerPath, gamePath)
	if err != nil {
		return nil, err
	}
	installerPath = preparedInstallerPath
	gamePath = preparedGamePath

	g, err := b.InstallGameConfig(ctx, gamePath, installHook)
	if err != nil {
		return nil, err
	}
	if g == nil {
		return nil, fmt.Errorf("game config was not created")
	}

	prefix := strings.TrimSpace(g.PrefixPath)
	if prefix == "" {
		return nil, fmt.Errorf("wine prefix is required to run installer")
	}

	r, err := autorunner.NewRunner(prefix)
	if err != nil {
		return nil, err
	}

	useGamescope := false
	if cfg, ok := autorunner.RunnerConfigFor(r); ok {
		useGamescope = cfg.UseGamescope
	}
	config := grinstaller.NewRunConfig(gamePath)
	config.InstallerPath = installerPath
	config.Auto = autorunner.DefaultOptionsConfig{
		WinePrefix:   prefix,
		UseGamescope: useGamescope,
	}
	plan, err := grinstaller.Plan(config)
	if err != nil {
		return nil, err
	}

	if plan.InstallerOptions.ExePath != "" {
		if _, err := r.Run(ctx, plan.InstallerOptions.ExePath, plan.InstallerOptions.Options...); err != nil {
			return nil, fmt.Errorf("run installer: %w", err)
		}
	}

	gameOpts := append([]gr.Option(nil), plan.GameOptions.Options...)
	gameOpts = append(gameOpts, gr.WithBackground(true))
	proc, err := r.Run(ctx, plan.GameOptions.ExePath, gameOpts...)
	if err != nil {
		return nil, fmt.Errorf("run game: %w", err)
	}

	if err := b.ReloadGames(); err != nil {
		return nil, err
	}
	b.gameMu.Lock()
	b.current = findGameByName(b.games, g.Name)
	if b.current == nil {
		b.current = g
	}
	b.currentRun = proc
	b.config.SelectGameName = g.Name
	b.gameMu.Unlock()

	return b.CurrentGame(), nil
}

func prepareInstallerPaths(ctx context.Context, installerPath, gamePath string) (string, string, error) {
	detection, err := grinstaller.DetectArchive(installerPath)
	if err != nil || detection.Kind == grinstaller.ArchiveUnknown {
		return installerPath, gamePath, nil
	}

	extraction, err := grinstaller.ExtractArchive(ctx, installerPath, grinstaller.ExtractConfig{})
	if err != nil {
		return "", "", fmt.Errorf("extract installer archive: %w", err)
	}

	return "", resolveExtractedGamePath(extraction.DestDir, gamePath), nil
}

func resolveExtractedGamePath(destDir, gamePath string) string {
	if gamePath == "" || filepath.IsAbs(gamePath) {
		return gamePath
	}
	if _, err := os.Stat(gamePath); err == nil {
		return gamePath
	}
	return filepath.Join(destDir, gamePath)
}

func (b *LiveBackend) RunGame(ctx context.Context, g *game.Game) (*gr.Process, error) {
	if g == nil {
		return nil, fmt.Errorf("game is required")
	}
	g = b.latestGameForRun(g)
	g = sanitizeGameForRun(g)
	if err := installGamescopeRunDeps(ctx, g); err != nil {
		return nil, err
	}
	b.resetGameEngineState(ctx, g)
	proc, err := b.runGame(ctx, g)
	if err != nil {
		return nil, err
	}

	b.gameMu.Lock()
	b.current = g
	b.currentRun = proc
	b.config.SelectGameName = g.Name
	b.gameMu.Unlock()
	return proc, nil
}

func (b *LiveBackend) MissingWinetrickDependencies(g *game.Game) ([]string, error) {
	if g == nil {
		return nil, fmt.Errorf("game is required")
	}
	g = b.latestGameForRun(g)
	g = sanitizeGameForRun(g)
	return missingGamescopeRunDeps(g)
}

func (b *LiveBackend) latestGameForRun(g *game.Game) *game.Game {
	if b == nil || g == nil {
		return g
	}
	name := strings.TrimSpace(g.Name)
	if name == "" {
		return g
	}

	b.gameMu.RLock()
	defer b.gameMu.RUnlock()
	if b.current != nil && strings.TrimSpace(b.current.Name) == name {
		return b.current
	}
	for _, saved := range b.games {
		if saved != nil && strings.TrimSpace(saved.Name) == name {
			return saved
		}
	}
	return g
}

func (b *LiveBackend) runGame(ctx context.Context, g *game.Game) (*gr.Process, error) {
	if shouldLaunchGamescopeDirectly(g) {
		return runGamescopeGame(ctx, g)
	}

	e, err := b.GetGameEngine(ctx, g)
	if err != nil {
		return nil, err
	}
	return e.RunGame(ctx, g)
}

func (b *LiveBackend) resetGameEngineState(ctx context.Context, g *game.Game) {
	e, err := b.GetGameEngine(ctx, g)
	if err != nil || e == nil {
		return
	}
	_ = e.Shutdown()
}

func shouldLaunchGamescopeDirectly(g *game.Game) bool {
	return g != nil &&
		g.Runner == game.RunnerGamescope &&
		g.GamescopeConfig != nil &&
		hasGamescopeRuntimeConfig(g.GamescopeConfig) &&
		!g.GamescopeConfig.Fullscreen
}

func runGamescopeGame(ctx context.Context, g *game.Game) (*gr.Process, error) {
	if err := enginerun.ValidateGame(g); err != nil {
		return nil, err
	}

	target, args := enginerun.WineTarget(g)
	opts, err := enginerun.WineOptions(g, args...)
	if err != nil {
		return nil, err
	}
	opts = append(opts, gr.WithBackground(true))

	return gamescope.NewFromOptions(gamescopeOptionsForGame(g)).Run(ctx, target, opts...)
}

func installGamescopeRunDeps(ctx context.Context, g *game.Game) error {
	deps, opts, err := missingGamescopeRunDepsWithOptions(g)
	if err != nil {
		return err
	}
	if len(deps) == 0 {
		return nil
	}
	return installGamescopeWinetricksDeps(ctx, g, opts, deps)
}

func missingGamescopeRunDeps(g *game.Game) ([]string, error) {
	deps, _, err := missingGamescopeRunDepsWithOptions(g)
	return deps, err
}

func missingGamescopeRunDepsWithOptions(g *game.Game) ([]string, []gr.Option, error) {
	if g == nil || g.Runner != game.RunnerGamescope {
		return nil, nil, nil
	}
	if err := enginerun.ValidateGame(g); err != nil {
		return nil, nil, err
	}
	_, args := enginerun.WineTarget(g)
	opts, err := enginerun.WineOptions(g, args...)
	if err != nil {
		return nil, nil, err
	}
	o := gr.ApplyOptions(opts...)
	return missingWinetrickDeps(gamescopeWinePrefix(g, o), o.Dependencies()), opts, nil
}

func installGamescopeWinetricksDeps(ctx context.Context, g *game.Game, opts []gr.Option, deps []string) error {
	o := gr.ApplyOptions(opts...)
	deps = normalizeWinetrickDeps(deps)
	if len(deps) == 0 {
		return nil
	}

	prefix := gamescopeWinePrefix(g, o)
	if strings.TrimSpace(prefix) == "" {
		return fmt.Errorf("wine prefix is required to install winetricks dependencies")
	}
	if err := os.MkdirAll(prefix, 0o755); err != nil {
		return fmt.Errorf("create wine prefix: %w", err)
	}

	args := append([]string{"-q"}, deps...)
	cmd := exec.CommandContext(ctx, winetricksBinForGame(g), args...)
	cmd.Env = gamescopeWinetricksEnv(prefix, o)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install winetricks deps %v: %w", deps, err)
	}
	return nil
}

func gamescopeWinePrefix(g *game.Game, o gr.Options) string {
	if prefix := strings.TrimSpace(o.WinePrefix()); prefix != "" {
		return prefix
	}
	if g == nil {
		return ""
	}
	if g.GamescopeConfig != nil {
		if prefix := strings.TrimSpace(g.GamescopeConfig.DefaultWinePrefix); prefix != "" {
			return prefix
		}
	}
	return strings.TrimSpace(g.PrefixPath)
}

func winetricksBinForGame(g *game.Game) string {
	if g != nil && g.WineConfig != nil {
		if bin := strings.TrimSpace(g.WineConfig.WineTricksBin); bin != "" {
			return bin
		}
	}
	return "winetricks"
}

func gamescopeWinetricksEnv(prefix string, o gr.Options) []string {
	env := os.Environ()
	env = upsertEnv(env, "WINEPREFIX", prefix)

	switch strings.ToLower(strings.TrimSpace(o.SystemArch())) {
	case "win32", "32":
		env = upsertEnv(env, "WINEARCH", "win32")
	case "win64", "64":
		env = upsertEnv(env, "WINEARCH", "win64")
	}

	for _, item := range o.Envs() {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		key, _, ok := strings.Cut(item, "=")
		if !ok || strings.TrimSpace(key) == "" {
			env = append(env, item)
			continue
		}
		env = upsertEnv(env, key, strings.TrimPrefix(item, key+"="))
	}
	return env
}

func upsertEnv(env []string, key, value string) []string {
	prefix := key + "="
	entry := prefix + value
	for i, item := range env {
		if strings.HasPrefix(item, prefix) {
			env[i] = entry
			return env
		}
	}
	return append(env, entry)
}

func missingWinetrickDeps(prefix string, deps []string) []string {
	deps = normalizeWinetrickDeps(deps)
	if len(deps) == 0 {
		return nil
	}

	installed := installedWinetrickDeps(prefix)
	if len(installed) == 0 {
		return deps
	}

	missing := make([]string, 0, len(deps))
	for _, dep := range deps {
		if _, ok := installed[dep]; !ok {
			missing = append(missing, dep)
		}
	}
	return missing
}

func installedWinetrickDeps(prefix string) map[string]struct{} {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return nil
	}
	data, err := os.ReadFile(filepath.Join(prefix, "winetricks.log"))
	if err != nil {
		return nil
	}
	items := strings.FieldsFunc(string(data), func(r rune) bool {
		return r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	installed := make(map[string]struct{})
	for _, item := range normalizeWinetrickDeps(items) {
		installed[item] = struct{}{}
	}
	return installed
}

func normalizeWinetrickDeps(items []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(strings.ToLower(item))
		item = strings.Trim(item, "\"'")
		if item == "" || strings.HasPrefix(item, "#") {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func gamescopeOptionsForGame(g *game.Game) gamescope.Options {
	cfg := gamescope.ApplyOptions()
	if g != nil && hasGamescopeRuntimeConfig(g.GamescopeConfig) {
		cfg = *g.GamescopeConfig
	}
	cfg.UseWine = true
	if g == nil {
		return cfg
	}
	if strings.TrimSpace(g.RunnerPath) != "" {
		cfg.GamescopeBin = strings.TrimSpace(g.RunnerPath)
	}
	if strings.TrimSpace(cfg.DefaultWinePrefix) == "" {
		cfg.DefaultWinePrefix = g.PrefixPath
	}
	if strings.TrimSpace(cfg.GamescopeBin) == "" {
		cfg.GamescopeBin = "gamescope"
	}
	if strings.TrimSpace(cfg.WineBin) == "" {
		cfg.WineBin = "wine"
	}
	if strings.TrimSpace(cfg.WineServerBin) == "" {
		cfg.WineServerBin = "wineserver"
	}
	if strings.TrimSpace(cfg.Name) == "" {
		cfg.Name = "gamescope"
	}
	return cfg
}

func hasGamescopeRuntimeConfig(c *gamescope.Options) bool {
	return c != nil && (c.Name != "" ||
		c.GamescopeBin != "" ||
		c.WineBin != "" ||
		c.WineServerBin != "" ||
		c.DefaultWinePrefix != "" ||
		c.UseWine ||
		c.WineStartWait ||
		c.KillWineOnExit ||
		c.Width != 0 ||
		c.Height != 0 ||
		c.RefreshRate != 0 ||
		c.OutputWidth != 0 ||
		c.OutputHeight != 0 ||
		c.Fullscreen ||
		c.Borderless ||
		c.ForceGrab ||
		c.SteamDeckMode ||
		c.ExposeWayland ||
		c.Scaler != "" ||
		c.Filter != "" ||
		len(c.ExtraArgs) > 0)
}

func sanitizeGameForRun(g *game.Game) *game.Game {
	if g == nil || strings.TrimSpace(g.RunnerPath) == "" {
		return g
	}
	path := strings.TrimSpace(g.RunnerPath)
	name := strings.ToLower(filepath.Base(path))
	switch g.Runner {
	case game.RunnerGamescope:
		if strings.Contains(name, "wine") || strings.Contains(name, "proton") {
			cp := *g
			cp.RunnerPath = ""
			return &cp
		}
	case game.RunnerWine:
		if strings.Contains(name, "gamescope") {
			cp := *g
			cp.RunnerPath = ""
			return &cp
		}
	}
	return g
}

func (b *LiveBackend) FollowGameText(ctx context.Context, g *game.Game) (chan engine.Line, error) {
	if g == nil {
		return nil, fmt.Errorf("game is required")
	}
	e, err := b.GetGameEngine(ctx, g)
	if err != nil {
		return nil, err
	}
	ch, err := e.FollowGameText(ctx, g, engine.FollowGameOptions{History: true, MaxLines: 100})
	if err != nil {
		return nil, err
	}
	return b.recordGameTextHookHistory(ctx, g, ch), nil
}

func (b *LiveBackend) GetGameTextHooks(ctx context.Context, g *game.Game) ([]string, bool, error) {
	if g == nil {
		return nil, false, fmt.Errorf("game is required")
	}
	e, err := b.GetGameEngine(ctx, g)
	if err != nil {
		return nil, false, err
	}
	tr := e.GetTextractor(g)
	if tr == nil {
		return nil, false, nil
	}

	seen := map[string]struct{}{}
	for hook := range tr.Hooks() {
		group := strings.TrimSpace(textractor.HookGroup(hook))
		if group != "" {
			seen[group] = struct{}{}
		}
	}
	for _, hook := range tr.HookGroups() {
		group := strings.TrimSpace(textractor.HookGroup(hook))
		if group != "" {
			seen[group] = struct{}{}
		}
	}

	hooks := make([]string, 0, len(seen))
	for hook := range seen {
		hooks = append(hooks, hook)
	}
	sort.Strings(hooks)
	return hooks, true, nil
}

func (b *LiveBackend) SetGameTextHookFilter(g *game.Game, filters []string) error {
	if g == nil {
		return fmt.Errorf("game is required")
	}

	normalized := normalizeTextHookFilters(filters)

	b.gameMu.Lock()
	for _, saved := range b.games {
		if saved == nil || saved.Name != g.Name {
			continue
		}
		saved.TextHookFilter = append([]string(nil), normalized...)
		g = saved
		break
	}
	if b.current != nil && b.current.Name == g.Name {
		b.current.TextHookFilter = append([]string(nil), normalized...)
	}
	g.TextHookFilter = append([]string(nil), normalized...)
	b.gameMu.Unlock()

	return gameConfig.WriteGameConfig(gameConfig.DefaultGameConfigPath(g), g)
}

func normalizeTextHookFilters(filters []string) []string {
	out := make([]string, 0, len(filters))
	seen := map[string]struct{}{}
	for _, filter := range filters {
		filter = strings.TrimSpace(textractor.HookGroup(filter))
		if filter == "" {
			continue
		}
		if _, ok := seen[filter]; ok {
			continue
		}
		out = append(out, filter)
		seen[filter] = struct{}{}
	}
	return out
}

func (b *LiveBackend) GetGameEngine(ctx context.Context, g *game.Game) (engine.EngineV2, error) {
	if g == nil {
		return nil, fmt.Errorf("game is required")
	}
	if b.engineSelector == nil {
		b.engineSelector = auto.NewEngineSelectorV2("")
	}

	if g.EngineName != "" {
		if e := b.engineSelector.ByName(g.EngineName); e != nil {
			return e, nil
		}
	}
	if g.Executable != "" {
		if e := b.engineSelector.ByName(g.Executable); e != nil {
			return e, nil
		}
	}
	if g.GamePath != "" {
		return b.engineSelector.Select(g.GamePath)
	}
	if g.WorkingDir != "" {
		return b.engineSelector.Select(g.WorkingDir)
	}
	return nil, fmt.Errorf("could not select engine for %q", g.Name)
}

func (b *LiveBackend) GetGamePlugins(ctx context.Context, g *game.Game) ([]*engine.Plugin, error) {
	e, err := b.GetGameEngine(ctx, g)
	if err != nil {
		return nil, err
	}
	plugins := e.GetPlugins()
	out := make([]*engine.Plugin, 0, len(plugins))
	for _, plugin := range plugins {
		if plugin == nil {
			continue
		}
		cp := *plugin
		cp.Game = g
		out = append(out, &cp)
	}
	return out, nil
}

func (b *LiveBackend) InstallGamePlugin(ctx context.Context, g *game.Game, name string) error {
	e, plugin, err := b.gamePlugin(ctx, g, name)
	if err != nil {
		return err
	}
	return e.InstallPlugin(plugin)
}

func (b *LiveBackend) UninstallGamePlugin(ctx context.Context, g *game.Game, name string) error {
	e, plugin, err := b.gamePlugin(ctx, g, name)
	if err != nil {
		return err
	}
	return e.UnInstallPlugin(plugin)
}

func (b *LiveBackend) gamePlugin(ctx context.Context, g *game.Game, name string) (engine.EngineV2, *engine.Plugin, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, nil, fmt.Errorf("plugin name is required")
	}
	e, err := b.GetGameEngine(ctx, g)
	if err != nil {
		return nil, nil, err
	}
	for _, plugin := range e.GetPlugins() {
		if plugin == nil || !samePluginName(plugin.Name, name) {
			continue
		}
		cp := *plugin
		cp.Game = g
		return e, &cp, nil
	}
	return nil, nil, fmt.Errorf("%w: %s", engine.ErrUnsupportedPlugin, name)
}

func samePluginName(a, b string) bool {
	normalize := func(v string) string {
		v = strings.ToLower(strings.TrimSpace(v))
		v = strings.ReplaceAll(v, "_", "-")
		v = strings.ReplaceAll(v, " ", "-")
		return v
	}
	return normalize(a) == normalize(b)
}

func (b *LiveBackend) ReloadGames() error {
	games, err := gameConfig.LoadInstalledGames(
		filepath.Join(gameConfig.ConfigBaseDir(), "games"),
		filepath.Join(util.ConfigBaseDir(), "games"),
	)
	if err != nil {
		//todo show in ui
		return err
	}
	b.gameMu.Lock()
	defer b.gameMu.Unlock()
	b.games = games
	if b.current == nil && b.config.SelectGameName != "" {
		b.current = findGameByName(games, b.config.SelectGameName)
	}

	return nil
}

func (b *LiveBackend) SelectGame(g *game.Game) {
	b.gameMu.Lock()
	defer b.gameMu.Unlock()
	b.current = g
	if g == nil {
		b.config.SelectGameName = ""
		return
	}
	b.config.SelectGameName = g.Name
}

func (b *LiveBackend) CurrentGame() *game.Game {
	b.gameMu.RLock()
	defer b.gameMu.RUnlock()
	return b.current
}

func (b *LiveBackend) SaveGameConfig(g *game.Game, previousName string) error {
	if g == nil {
		return fmt.Errorf("game is required")
	}

	if err := gameConfig.WriteGameConfig(gameConfig.DefaultGameConfigPath(g), g); err != nil {
		return err
	}

	oldName := strings.TrimSpace(previousName)
	if oldName != "" && oldName != strings.TrimSpace(g.Name) {
		oldFile := vnutil.SanitizeName(oldName) + ".json"
		oldPath := filepath.Join(gameConfig.ConfigBaseDir(), "games", oldFile)
		oldLegacyPath := filepath.Join(util.ConfigBaseDir(), "games", oldFile)
		newPath := gameConfig.DefaultGameConfigPath(g)
		if oldPath != newPath {
			_ = os.Remove(oldPath)
		}
		if oldLegacyPath != newPath {
			_ = os.Remove(oldLegacyPath)
		}
	}

	b.gameMu.Lock()
	b.current = g
	b.config.SelectGameName = g.Name
	b.gameMu.Unlock()

	return b.ReloadGames()
}

func (b *LiveBackend) DeleteGameConfig(g *game.Game) error {
	if g == nil {
		return fmt.Errorf("game is required")
	}
	name := strings.TrimSpace(g.Name)
	if name == "" {
		return fmt.Errorf("game name is required")
	}

	file := vnutil.SanitizeName(name) + ".json"
	paths := []string{
		filepath.Join(gameConfig.ConfigBaseDir(), "games", file),
		filepath.Join(util.ConfigBaseDir(), "games", file),
	}
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	b.gameMu.Lock()
	if b.current != nil && strings.TrimSpace(b.current.Name) == name {
		b.current = nil
		b.currentRun = nil
		b.config.SelectGameName = ""
	}
	b.gameMu.Unlock()

	return b.ReloadGames()
}

func (b *LiveBackend) DeleteCustomRunnerConfig(g *game.Game) error {
	if g == nil {
		return fmt.Errorf("game is required")
	}
	g.RunnerPath = ""
	g.RunnerConfig = gr.Config{}
	g.WineConfig = nil
	g.GamescopeConfig = nil
	if err := gameConfig.WriteGameConfig(gameConfig.DefaultGameConfigPath(g), g); err != nil {
		return err
	}

	b.gameMu.Lock()
	if b.current != nil && strings.TrimSpace(b.current.Name) == strings.TrimSpace(g.Name) {
		b.current = g
	}
	b.gameMu.Unlock()

	return b.ReloadGames()
}

func (b *LiveBackend) StopCurrentGame() {
	b.gameMu.RLock()
	g := b.current
	proc := b.currentRun
	b.gameMu.RUnlock()
	if g == nil {
		return
	}

	e, err := b.GetGameEngine(context.Background(), g)
	if err == nil && e != nil {
		if proc != nil {
			_, _ = e.StopGame(context.Background(), proc)
		}
		_ = e.Shutdown()
	}

	b.gameMu.Lock()
	if b.currentRun == proc {
		b.currentRun = nil
	}
	b.gameMu.Unlock()
}

func (b *LiveBackend) SearchTerm(search jpndict.Search) (*jpndict.Response, error) {
	dict, err := b.dictionary()
	if err != nil {
		return nil, err
	}
	return dict.Search(search)
}

func (b *LiveBackend) SearchAllTerm(search jpndict.Search) ([]*jpndict.Response, error) {
	dict, err := b.dictionary()
	if err != nil {
		return nil, err
	}
	return dict.SearchAll(search)
}

func (b *LiveBackend) AnalyzeSentence(sentence string) (japanese.Analysis, error) {
	return japanese.AnalyzeSentence(sentence)
}

func (b *LiveBackend) TranslatorConfig() TranslatorConfig {
	if b == nil {
		return defaultTranslatorConfig()
	}
	b.translatorMu.Lock()
	defer b.translatorMu.Unlock()
	cfg := b.translatorCfg
	cfg.normalize()
	return cfg
}

func (b *LiveBackend) SaveTranslatorConfig(cfg TranslatorConfig) error {
	if b == nil {
		return fmt.Errorf("backend is required")
	}
	cfg.normalize()
	if err := saveTranslatorConfig(cfg); err != nil {
		return err
	}

	b.translatorMu.Lock()
	old := b.translator
	b.translatorCfg = cfg
	b.translator = newOllamaTranslator(cfg)
	b.translatorMu.Unlock()

	if old != nil {
		old.Close()
	}
	return nil
}

func (b *LiveBackend) Translate(ctx context.Context, r *translate.Request) (*translate.Response, error) {
	translator := b.currentTranslator()
	return translator.Translate(ctx, r)
}

func (b *LiveBackend) Search(ctx context.Context, r *translate.Request) (*translate.Response, error) {
	translator := b.currentTranslator()
	return translator.Search(ctx, r)
}

func (b *LiveBackend) Close() {
	if b == nil {
		return
	}
	b.translatorMu.Lock()
	translator := b.translator
	b.translator = nil
	b.translatorMu.Unlock()
	if translator != nil {
		translator.Close()
	}
}

func (b *LiveBackend) SupportedLanguage() []translate.Language {
	translator := b.currentTranslator()
	return translator.SupportedLanguage()
}

func (b *LiveBackend) IsLanguageSupported(from translate.Language, to translate.Language) bool {
	translator := b.currentTranslator()
	return translator.IsLanguageSupported(from, to)
}

func (b *LiveBackend) SupportedModels() []string {
	translator := b.currentTranslator()
	return translator.SupportedModels()
}

func (b *LiveBackend) currentTranslator() translate.Translate {
	b.translatorMu.Lock()
	defer b.translatorMu.Unlock()
	if b.translator == nil {
		b.translator = newOllamaTranslator(b.translatorCfg)
	}
	return b.translator
}

func (b *LiveBackend) dictionary() (jpndict.Dictonary, error) {
	b.dictMu.Lock()
	defer b.dictMu.Unlock()
	if b.dict != nil {
		return b.dict, nil
	}

	dict, err := jpndict.NewJiTenDex(filepath.Join(util.ConfigBaseDir(), "jpndict"), "", true)
	if err != nil {
		return nil, err
	}
	if err := dict.Download(); err != nil {
		return nil, err
	}
	b.dict = dict
	return b.dict, nil
}

func findGameByName(games []*game.Game, name string) *game.Game {
	for _, g := range games {
		if g != nil && g.Name == name {
			return g
		}
	}
	return nil
}
