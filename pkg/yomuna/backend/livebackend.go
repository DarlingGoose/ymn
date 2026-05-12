package backend

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/DarlingGoose/gr"
	"github.com/DarlingGoose/jpndict"
	"github.com/DarlingGoose/jpndict/translate"
	"github.com/DarlingGoose/vntext/pkg/engine"
	"github.com/DarlingGoose/vntext/pkg/engine/auto"
	"github.com/DarlingGoose/vntext/pkg/game"
	"github.com/DarlingGoose/vntext/pkg/gameConfig"
	"github.com/DarlingGoose/wgl/pkg/japanese"
	"github.com/DarlingGoose/wgl/pkg/util"
)

var _ Backend = &LiveBackend{}

type LiveBackend struct {
	gameMu     sync.RWMutex
	games      []*game.Game
	config     TranscriptConfig
	current    *game.Game
	currentRun *gr.Process

	engineSelector *auto.EngineSelectorV2
	translator     translate.Translate

	dictMu sync.Mutex
	dict   jpndict.Dictonary

	currentLines []engine.Line
}

func NewLive() *LiveBackend {
	return &LiveBackend{
		engineSelector: auto.NewEngineSelectorV2(""),
		translator:     translate.NewOllamaClient(translate.OllamaConfig{}),
	}
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

func (b *LiveBackend) RunGame(ctx context.Context, g *game.Game) (*gr.Process, error) {
	if g == nil {
		return nil, fmt.Errorf("game is required")
	}
	e, err := b.GetGameEngine(ctx, g)
	if err != nil {
		return nil, err
	}
	proc, err := e.RunGame(ctx, g)
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

func (b *LiveBackend) Translate(ctx context.Context, r *translate.Request) (*translate.Response, error) {
	if b.translator == nil {
		b.translator = translate.NewOllamaClient(translate.OllamaConfig{})
	}
	return b.translator.Translate(ctx, r)
}

func (b *LiveBackend) Search(ctx context.Context, r *translate.Request) (*translate.Response, error) {
	if b.translator == nil {
		b.translator = translate.NewOllamaClient(translate.OllamaConfig{})
	}
	return b.translator.Search(ctx, r)
}

func (b *LiveBackend) Close() {
	if b.translator != nil {
		b.translator.Close()
	}
}

func (b *LiveBackend) SupportedLanguage() []translate.Language {
	if b.translator == nil {
		b.translator = translate.NewOllamaClient(translate.OllamaConfig{})
	}
	return b.translator.SupportedLanguage()
}

func (b *LiveBackend) IsLanguageSupported(from translate.Language, to translate.Language) bool {
	if b.translator == nil {
		b.translator = translate.NewOllamaClient(translate.OllamaConfig{})
	}
	return b.translator.IsLanguageSupported(from, to)
}

func (b *LiveBackend) SupportedModels() []string {
	if b.translator == nil {
		b.translator = translate.NewOllamaClient(translate.OllamaConfig{})
	}
	return b.translator.SupportedModels()
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
