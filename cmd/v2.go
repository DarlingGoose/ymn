/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/widget/material"
	"github.com/DarlingGoose/wgl/pkg/v2/examples"
	"github.com/spf13/cobra"
)

// v2Cmd represents the v2 command
var v2Cmd = &cobra.Command{
	Use:   "v2",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		appName, err := cmd.Flags().GetString("app")
		if err != nil {
			return
		}
		go func() {
			window := new(app.Window)

			err := run(window, appName)
			if err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		}()
		app.Main()
	},
}

func init() {
	v2Cmd.Flags().String("app", "slider", "")
	rootCmd.AddCommand(v2Cmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// v2Cmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// v2Cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func run(window *app.Window, appName string) error {
	theme := material.NewTheme()
	var ops op.Ops
	sliderApp := examples.NewSliderAppUI(theme)
	sidebarApp := examples.NewSidebarAppUI(theme)
	tabApp := examples.NewTabApp()
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			// This graphics context is used for managing the rendering state.
			gtx := app.NewContext(&ops, e)
			switch appName {
			case "tab":
				tabApp.Layout(gtx)
			case "slider":
				sliderApp.Layout(gtx)
			case "sidebar":
				sidebarApp.Layout(gtx)
			default:
				sliderApp.Layout(gtx)
			}

			e.Frame(gtx.Ops)
		}
	}
}
