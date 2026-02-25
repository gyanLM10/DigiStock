package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// projectDir is a persistent flag shared by all subcommands
var projectDir string

var rootCmd = &cobra.Command{
	Use:   "digistock",
	Short: "DigiStock CLI — AI-powered NSE stock analysis",
	Long: color.New(color.FgCyan, color.Bold).Sprint(`
╔═══════════════════════════════════════════════════╗
║          DigiStock — AI NSE Trading Bot           ║
║   Multi-agent stock analysis at your fingertips   ║
╚═══════════════════════════════════════════════════╝`) + `

Runs the multi-agent pipeline directly — no server required.
The CLI locates your DigiStock project automatically and
invokes the Python agents as a subprocess.`,
}

// Execute is the entry point called from main.go
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&projectDir,
		"dir", "d",
		"",
		"Path to DigiStock project root (auto-detected if not set)",
	)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(healthCmd)
}
