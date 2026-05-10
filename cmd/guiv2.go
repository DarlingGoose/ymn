package cmd

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/DarlingGoose/wgl/pkg/yomuna"
	"github.com/spf13/cobra"
)

var guiv2Source string

var guiv2Cmd = &cobra.Command{
	Use:   "guiv2 [source-text]",
	Short: "Open the v2 Yomuna GUI",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(os.Getenv("DISPLAY")) == "" && strings.TrimSpace(os.Getenv("WAYLAND_DISPLAY")) == "" {
			return errors.New("guiv2 mode requires a desktop session with DISPLAY or WAYLAND_DISPLAY set")
		}

		source := strings.TrimSpace(guiv2Source)
		if source == "" && len(args) > 0 {
			source = strings.TrimSpace(args[0])
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		return yomuna.New(source).Run(ctx)
	},
}

func init() {
	rootCmd.AddCommand(guiv2Cmd)
	guiv2Cmd.Flags().StringVar(&guiv2Source, "source", "", "initial source text for the translation page")
}
