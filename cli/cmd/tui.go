package cmd

import (
	"fmt"
	"os"

	"digistock-cli/tui"

	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive terminal UI",
	Long: `Open a fullscreen interactive terminal interface for DigiStock.

Navigate all commands with arrow keys — no flags needed.

  🤖 Analyze       AI multi-agent Buy/Sell/Hold recommendation
  📊 Indicators    RSI, MACD, SMA, EMA for any NSE stock
  🔮 Predict       Next-N-day price prediction
  📈 Backtest      SMA crossover strategy vs buy-and-hold
  🔍 Health Check  Verify your environment

Example:
  digistock tui`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := findProjectRoot(projectDir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "✗ "+err.Error())
			return err
		}
		python := resolvePython(root)
		return tui.Start(tui.Config{Root: root, Python: python})
	},
}
