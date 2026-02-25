package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [query]",
	Short: "Run a multi-agent NSE stock analysis",
	Long: `Submit a query to the DigiStock multi-agent pipeline.

The agents will identify NSE stocks, gather market data, analyze news,
and produce a structured trading recommendation — all streamed live to your terminal.

The command runs the Python agent pipeline directly — no server required.

Examples:
  digistock analyze
  digistock analyze "What are the top NSE stocks to buy this week?"
  digistock analyze "Should I buy RELIANCE or INFY today?"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		query := "Analyze current NSE market and recommend 2 stocks to trade today."
		if len(args) > 0 {
			query = strings.Join(args, " ")
		}
		return runAnalyze(query)
	},
}

// findProjectRoot walks up from the CLI binary location looking for pyproject.toml
func findProjectRoot(hint string) (string, error) {
	if hint != "" {
		return hint, nil
	}

	// Try sibling of cli/ binary location
	exe, err := os.Executable()
	if err == nil {
		// cli/ is inside the project, so parent is project root
		parent := filepath.Dir(filepath.Dir(exe))
		if _, err := os.Stat(filepath.Join(parent, "pyproject.toml")); err == nil {
			return parent, nil
		}
		// Try immediate parent
		parent2 := filepath.Dir(exe)
		if _, err := os.Stat(filepath.Join(parent2, "pyproject.toml")); err == nil {
			return parent2, nil
		}
	}

	// Fall back: walk up from cwd
	cwd, _ := os.Getwd()
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "pyproject.toml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("could not locate DigiStock project root (no pyproject.toml found); use --dir to specify it")
}

// resolvePython finds the best Python interpreter to use
func resolvePython(root string) string {
	// Prefer uv run python, then virtualenv, then system python3
	candidates := []string{
		filepath.Join(root, ".venv", "bin", "python"),
		filepath.Join(root, "venv", "bin", "python"),
		"python3",
		"python",
	}
	for _, c := range candidates {
		if path, err := exec.LookPath(c); err == nil {
			return path
		}
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return "python3"
}

func runAnalyze(query string) error {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed, color.Bold)
	white := color.New(color.FgWhite)
	dim := color.New(color.Faint)
	magenta := color.New(color.FgMagenta)

	cyan.Println("\n╔══════════════════════════════════════════════════╗")
	cyan.Println("║          DigiStock Multi-Agent Analysis          ║")
	cyan.Println("╚══════════════════════════════════════════════════╝")
	fmt.Printf("\n%s %s\n\n", yellow.Sprint("📊 Query:"), white.Sprint(query))

	// ── Locate project root ────────────────────────────────────────────────
	root, err := findProjectRoot(projectDir)
	if err != nil {
		red.Println("✗ " + err.Error())
		return err
	}
	dim.Printf("Project root: %s\n\n", root)

	runner := filepath.Join(root, "runner.py")
	if _, err := os.Stat(runner); err != nil {
		red.Printf("✗ runner.py not found at %s\n", runner)
		return fmt.Errorf("runner.py missing")
	}

	python := resolvePython(root)

	// ── Spinner ─────────────────────────────────────────────────────────────
	sp := spinner.New(spinner.CharSets[14], 80*time.Millisecond)
	sp.Suffix = "  Initialising agent pipeline..."
	sp.Color("cyan")
	sp.Start()

	// ── Spawn Python subprocess ──────────────────────────────────────────────
	cmd := exec.Command(python, runner, query)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1") // force line-by-line flushing

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		sp.Stop()
		return fmt.Errorf("failed to attach stdout: %w", err)
	}
	cmd.Stderr = os.Stderr // pass Python errors through directly

	if err := cmd.Start(); err != nil {
		sp.Stop()
		red.Printf("✗ Failed to start Python: %v\n", err)
		return err
	}

	// ── Stream output ────────────────────────────────────────────────────────
	firstChunk := true
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1024*64), 1024*64)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			fmt.Println()
			continue
		}

		if firstChunk {
			sp.Stop()
			green.Println("✔ Pipeline started — streaming output...\n")
			dim.Println(strings.Repeat("─", 52))
			firstChunk = false
		}

		switch {
		case strings.Contains(line, "Calling Sub-Agent:"):
			fmt.Println()
			cyan.Println(line)
		case strings.Contains(line, "Responding as Supervisor:"):
			fmt.Println()
			green.Println(line)
		case strings.Contains(line, "Response from"):
			magenta.Println(line)
		case strings.HasPrefix(line, "**") && strings.HasSuffix(line, "**"):
			yellow.Println(line)
		default:
			white.Println(line)
		}
	}

	if err := scanner.Err(); err != nil {
		red.Printf("\n✗ Stream error: %v\n", err)
	}

	if err := cmd.Wait(); err != nil {
		if !firstChunk {
			red.Printf("\n✗ Agent pipeline exited with error: %v\n", err)
		}
		return err
	}

	if firstChunk {
		sp.Stop()
		yellow.Println("⚠  No output received from the agent pipeline.")
		return nil
	}

	fmt.Println()
	dim.Println(strings.Repeat("─", 52))
	green.Println("\n✔ Analysis complete.")
	return nil
}
