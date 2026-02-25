package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Check Python environment and required env vars",
	Long: `Verifies the DigiStock environment is ready to run:
  - Python interpreter is available
  - runner.py exists in the project root
  - Required API keys are set in the environment

Examples:
  digistock health
  digistock health --dir /path/to/DigiStock`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := findProjectRoot(projectDir)
		if err != nil {
			color.New(color.FgRed, color.Bold).Println("✗ " + err.Error())
			return err
		}
		return runHealth(root)
	},
}

func runHealth(root string) error {
	green := color.New(color.FgGreen, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow)
	dim := color.New(color.Faint)
	bold := color.New(color.Bold)

	bold.Println("\n🔍 DigiStock Environment Check")
	dim.Println(strings.Repeat("─", 40))

	allOK := true
	check := func(label, detail string, ok bool) {
		if ok {
			green.Printf("  ✔ %-28s %s\n", label, dim.Sprint(detail))
		} else {
			red.Printf("  ✗ %-28s %s\n", label, detail)
			allOK = false
		}
	}

	// 1. Project root
	check("Project root found", root, true)

	// 2. runner.py
	runnerPath := filepath.Join(root, "runner.py")
	_, runnerErr := os.Stat(runnerPath)
	check("runner.py", runnerPath, runnerErr == nil)

	// 3. agent_logic.py
	agentPath := filepath.Join(root, "agent_logic.py")
	_, agentErr := os.Stat(agentPath)
	check("agent_logic.py", agentPath, agentErr == nil)

	// 4. Python interpreter
	python := resolvePython(root)
	out, err := exec.Command(python, "--version").CombinedOutput()
	pythonVersion := strings.TrimSpace(string(out))
	check("Python interpreter", pythonVersion, err == nil)

	// 5. .env file
	envPath := filepath.Join(root, ".env")
	_, envErr := os.Stat(envPath)
	check(".env file", envPath, envErr == nil)

	// 6. Required env vars (read from environment, .env is loaded by Python)
	required := []string{"OPENAI_API_KEY", "BRIGHT_DATA_API_TOKEN"}
	for _, key := range required {
		val := os.Getenv(key)
		present := val != ""
		detail := "set"
		if !present {
			detail = "NOT SET — add to your .env file"
		}
		check(key, detail, present)
	}

	// 7. Optional env vars
	optional := []string{"WEB_UNLOCKER_ZONE", "BROWSER_ZONE"}
	for _, key := range optional {
		val := os.Getenv(key)
		if val == "" {
			yellow.Printf("  ⚠ %-28s (optional, using default)\n", key)
		} else {
			green.Printf("  ✔ %-28s %s\n", key, dim.Sprint(val))
		}
	}

	dim.Println(strings.Repeat("─", 40))
	fmt.Println()
	if allOK {
		green.Println("✔ All checks passed — run: digistock analyze")
	} else {
		red.Println("✗ Some checks failed — resolve the issues above before running analyze")
	}
	fmt.Println()
	return nil
}
