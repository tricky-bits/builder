package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/tricky-bits/builder/internal/builder"
	"github.com/tricky-bits/builder/internal/serve"
)

var rootCmd = &cobra.Command{
	Use:          "builder",
	Short:        "Static site generator for tricky-bits",
	SilenceUsage: true,
	RunE:         runBuild,
}

var configPath string

func runBuild(cmd *cobra.Command, args []string) error {
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
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the static website from config and markdown files",
	RunE:  runBuild,
}

var bindAddr string
var port int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Build and serve the site locally, rebuilding on file changes",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		return serve.Run(ctx, serve.Options{
			ConfigPath: configPath,
			Bind:       bindAddr,
			Port:       port,
		})
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config.toml", "path to config file")
	serveCmd.Flags().StringVar(&bindAddr, "bind", "127.0.0.1", "address to bind the dev server to")
	serveCmd.Flags().IntVar(&port, "port", 1313, "port to serve on")
	rootCmd.AddCommand(buildCmd, serveCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
