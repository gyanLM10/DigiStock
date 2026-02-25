package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var indicatorsCmd = &cobra.Command{
	Use:   "indicators <TICKER>",
	Short: "Fetch technical indicators for an NSE stock",
	Long: `Compute RSI, MACD, SMA-50, SMA-200, and EMA-20 for a given NSE ticker.

Data is sourced from Yahoo Finance (no API key required).

Examples:
  digistock indicators TCS
  digistock indicators RELIANCE
  digistock indicators INFY`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runToolCmd("indicators", args[0], "")
	},
}

var predictDays int

var predictCmd = &cobra.Command{
	Use:   "predict <TICKER>",
	Short: "Predict the next-N-day price for an NSE stock",
	Long: `Use the trained XGBoost model to predict the stock price N days ahead.

Requires models/xgb_model.json and models/scaler.pkl in the project root.

Examples:
  digistock predict TCS
  digistock predict RELIANCE --days 10`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		extra := fmt.Sprintf("%d", predictDays)
		return runToolCmd("predict", args[0], extra)
	},
}

var backtestStrategy string

var backtestCmd = &cobra.Command{
	Use:   "backtest <TICKER>",
	Short: "Backtest an SMA crossover strategy on an NSE stock",
	Long: `Run a 1-year SMA-50/200 crossover backtest and compare it against buy-and-hold.

Data is sourced from Yahoo Finance (no API key required).

Examples:
  digistock backtest TCS
  digistock backtest RELIANCE --strategy sma_cross`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runToolCmd("backtest", args[0], backtestStrategy)
	},
}

// runToolCmd is the shared runner for all tool subcommands.
// It spawns: python runner_tools.py <command> <ticker> [extra]
func runToolCmd(command, ticker, extra string) error {
	cyan := color.New(color.FgCyan, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	white := color.New(color.FgWhite)

	root, err := findProjectRoot(projectDir)
	if err != nil {
		red.Println("✗ " + err.Error())
		return err
	}

	python := resolvePython(root)
	runnerTools := filepath.Join(root, "runner_tools.py")
	if _, err := os.Stat(runnerTools); err != nil {
		red.Printf("✗ runner_tools.py not found at %s\n", runnerTools)
		return fmt.Errorf("runner_tools.py missing")
	}

	pyArgs := []string{runnerTools, command, ticker}
	if extra != "" {
		pyArgs = append(pyArgs, extra)
	}

	cyan.Printf("\n⠿  Running %s for %s...\n\n", command, ticker)

	cmd := exec.Command(python, pyArgs...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to attach stdout: %w", err)
	}

	if err := cmd.Start(); err != nil {
		red.Printf("✗ Failed to start Python: %v\n", err)
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		white.Println(scanner.Text())
	}

	if err := cmd.Wait(); err != nil {
		red.Printf("\n✗ Command exited with error: %v\n", err)
		return err
	}

	fmt.Println()
	green.Println("✔ Done.")
	return nil
}

func init() {
	predictCmd.Flags().IntVar(&predictDays, "days", 5, "Number of days to predict ahead")
	backtestCmd.Flags().StringVar(&backtestStrategy, "strategy", "sma_cross", "Backtest strategy (currently: sma_cross)")
}
