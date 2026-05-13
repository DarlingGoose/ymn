package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/DarlingGoose/gr"
	"github.com/DarlingGoose/gr/gamescope"
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

	currentLines []engine.Line
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

func (b *LiveBackend) RunGame(ctx context.Context, g *game.Game) (*gr.Process, error) {
	if g == nil {
		return nil, fmt.Errorf("game is required")
	}
	g = sanitizeGameForRun(g)
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
	return e.FollowGameText(ctx, g, engine.FollowGameOptions{History: true, MaxLines: 200})
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
	if g == nil || proc == nil {
		return
	}

	e, err := b.GetGameEngine(context.Background(), g)
	if err == nil && e != nil {
		_, _ = e.StopGame(context.Background(), proc)
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
