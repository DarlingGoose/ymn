package backend

import (
	"context"

	"gioui.org/unit"
	"github.com/DarlingGoose/gr"
	"github.com/DarlingGoose/jpndict"
	"github.com/DarlingGoose/jpndict/translate"
	"github.com/DarlingGoose/vntext/pkg/engine"
	"github.com/DarlingGoose/vntext/pkg/game"
	"github.com/DarlingGoose/ymn/pkg/japanese"
)

type Backend interface {
	GameLogic
	translate.Translate
	//make sure to run download for jpndict
	SearchTerm(search jpndict.Search) (*jpndict.Response, error)
	SearchAllTerm(search jpndict.Search) ([]*jpndict.Response, error)

	// jpn analuze
	AnalyzeSentence(sentence string) (japanese.Analysis, error)
}

type GameLogic interface {
	ReloadGames() error
	GetGames() []*game.Game
	InstallGameConfig(ctx context.Context, inputPath string, installHook bool) (*game.Game, error)
	InstallGameWithInstaller(ctx context.Context, installerPath, gamePath string, installHook bool) (*game.Game, error)
	RunGame(ctx context.Context, g *game.Game) (*gr.Process, error)
	MissingWinetrickDependencies(g *game.Game) ([]string, error)
	FollowGameText(ctx context.Context, g *game.Game) (chan engine.Line, error)
	ReadGameTextHookHistory(g *game.Game, filters []string, maxLines int) ([]engine.Line, error)
	//todo GetTesseract()
	GetGameEngine(ctx context.Context, g *game.Game) (engine.EngineV2, error)
	GetGamePlugins(ctx context.Context, g *game.Game) ([]*engine.Plugin, error)
	InstallGamePlugin(ctx context.Context, g *game.Game, name string) error
	UninstallGamePlugin(ctx context.Context, g *game.Game, name string) error
	GetGameTextHooks(ctx context.Context, g *game.Game) ([]string, bool, error)
	SetGameTextHookFilter(g *game.Game, filters []string) error
	SelectGame(g *game.Game)
	SelectGameByName(n string)
	CurrentGame() *game.Game
	SaveGameConfig(g *game.Game, previousName string) error
	DeleteGameConfig(g *game.Game) error
	DeleteCustomRunnerConfig(g *game.Game) error
	StopCurrentGame()
	IsGameRunning() bool
}

type TranslatorConfig struct {
	OllamaModel   string `json:"ollama_model,omitempty"`
	OllamaBaseURL string `json:"ollama_base_url,omitempty"`
}

// load/save
type TranscriptConfig struct {
	SelectGameName      string  `yaml:"selectGameName"`
	TranscriptTextSize  unit.Dp `yaml:"transcriptTextSize"`
	TranslationTextSize unit.Dp `yaml:"translationTextSize"`
}
