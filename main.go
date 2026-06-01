package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tricky-bits/builder/internal/builder"
)

var rootCmd = &cobra.Command{
	Use:          "builder",
	Short:        "Static site generator for tricky-bits",
	SilenceUsage: true,
}

var configPath string

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the static website from config and markdown files",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := builder.ReadConfigFile(configPath)
		if err != nil {
			return err
		}
		b, err := builder.New(cfg)
		if err != nil {
			return err
		}
		if err := b.Load(); err != nil {
			return err
		}
		return b.Build()
	},
}

func init() {
	buildCmd.Flags().StringVarP(&configPath, "config", "c", "tbb.toml", "path to config file")
	rootCmd.AddCommand(buildCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
