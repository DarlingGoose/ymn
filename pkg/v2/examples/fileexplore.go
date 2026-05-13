package examples

import (
	"fmt"

	"gioui.org/layout"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/backend/media"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/core/theme"
	"github.com/DarlingGoose/ymn/pkg/v2/gui/pages/fileexplorer"
)

type FileExplore struct {
	Files *fileexplorer.FileExplorer
}

func NewAppPage(tc *theme.Client) *FileExplore {
	if tc == nil {
		tc = theme.DefaultThemeClient
	}
	files := fileexplorer.NewFileExplorer(
		"/home/n9s/Downloads",
		media.DefaultRegistry,
		tc,
	)
	files.OnSelect = func(path string) {
		// Fires when a file row is clicked and previewed.
		fmt.Println("preview:", path)
	}

	files.OnChoose = func(path string) {
		// Fires when the Select button is clicked.
		// Returns selected file, or current directory if no file is selected.
		fmt.Println("chosen:", path)
	}

	files.ShowSelectButton = true
	files.SelectButtonText = "Select"
	return &FileExplore{
		Files: files,
	}
}

func (p *FileExplore) Layout(gtx layout.Context) layout.Dimensions {
	return p.Files.Layout(gtx)
}
