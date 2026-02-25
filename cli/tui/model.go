package tui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── State ─────────────────────────────────────────────────────────────────────

type appState int

const (
	stateMenu    appState = iota
	stateForm             // input fields before running
	stateRunning          // subprocess streaming
	stateResult           // scrollable result
)

// ── Menu Entries ──────────────────────────────────────────────────────────────

type fieldSpec struct {
	label       string
	placeholder string
	defaultVal  string
	hint        string // shown beneath the field in dim
	examples    string // comma-separated examples
}

type menuEntry struct {
	icon   string
	label  string
	desc   string
	note   string      // one-line context shown at top of form
	fields []fieldSpec // nil = no form (health check)
}

var entries = []menuEntry{
	{
		"🤖", "Analyze",
		"AI multi-agent Buy/Sell/Hold recommendation",
		"Ask a question about NSE stocks. The AI team will research and give a Buy/Sell/Hold call. Leave blank to auto-pick 2 stocks.",
		[]fieldSpec{{
			"Your query",
			"e.g. Should I buy TCS or INFY today?",
			"",
			"Leave blank to let the AI choose stocks automatically",
			"Should I buy RELIANCE or HDFCBANK?  ·  Top momentum stocks this week  ·  Short-term trade for tomorrow",
		}},
	},
	{
		"📊", "Indicators",
		"RSI, MACD, SMA-50/200, EMA-20 for a stock",
		"Fetches live technical indicator data from Yahoo Finance. No API key required.",
		[]fieldSpec{{
			"NSE ticker",
			"e.g. TCS",
			"",
			"Just the symbol — .NS suffix is added for you automatically",
			"TCS  ·  RELIANCE  ·  INFY  ·  HDFCBANK  ·  ICICIBANK",
		}},
	},
	{
		"🔮", "Predict",
		"Next-N-day price prediction (XGBoost)",
		"Uses a trained XGBoost model to predict the closing price N trading days from today.",
		[]fieldSpec{
			{
				"NSE ticker",
				"e.g. RELIANCE",
				"",
				"Just the symbol — .NS suffix is added for you automatically",
				"TCS  ·  RELIANCE  ·  INFY  ·  HDFCBANK  ·  ICICIBANK",
			},
			{
				"Forecast horizon",
				"Number of trading days ahead (default: 5)",
				"5",
				"How many trading days ahead to predict",
				"5  (1 week)  ·  10  (2 weeks)  ·  21  (1 month)",
			},
		},
	},
	{
		"📈", "Backtest",
		"SMA-50/200 crossover vs buy-and-hold",
		"Runs a 1-year Golden Cross / Death Cross strategy and compares performance to buy-and-hold.",
		[]fieldSpec{
			{
				"NSE ticker",
				"e.g. INFY",
				"",
				"Just the symbol — .NS suffix is added for you automatically",
				"TCS  ·  RELIANCE  ·  INFY  ·  HDFCBANK  ·  ICICIBANK",
			},
			{
				"Strategy",
				"sma_cross",
				"sma_cross",
				"Only sma_cross (SMA-50/200 crossover) is currently supported",
				"",
			},
		},
	},
	{
		"🔍", "Health Check",
		"Verify Python, env files, and API keys",
		"",
		nil,
	},
}

// ── Messages ──────────────────────────────────────────────────────────────────

type lineMsg string
type doneMsg struct{}

// ── Config ────────────────────────────────────────────────────────────────────

type Config struct {
	Root   string
	Python string
}

// ── Model ─────────────────────────────────────────────────────────────────────

type Model struct {
	cfg    Config
	state  appState
	cursor int

	// form
	selIdx int
	inputs []textinput.Model
	focus  int

	// running / result
	sp     spinner.Model
	vp     viewport.Model
	lines  []string
	lineCh chan string

	width  int
	height int
}

func New(cfg Config) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Points
	sp.Style = lipgloss.NewStyle().Foreground(cyan)
	return Model{cfg: cfg, sp: sp}
}

func (m Model) Init() tea.Cmd { return nil }

// ── Update ────────────────────────────────────────────────────────────────────

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.vp.Width = msg.Width - 4
		m.vp.Height = msg.Height - 9
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.sp, cmd = m.sp.Update(msg)
		return m, cmd

	case lineMsg:
		m.lines = append(m.lines, string(msg))
		m.vp.SetContent(renderLines(m.lines))
		m.vp.GotoBottom()
		return m, waitLine(m.lineCh)

	case doneMsg:
		m.state = stateResult
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateMenu:
			return m.updateMenu(msg)
		case stateForm:
			return m.updateForm(msg)
		case stateRunning:
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		case stateResult:
			return m.updateResult(msg)
		}
	}
	return m, nil
}

func (m Model) updateMenu(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(entries)-1 {
			m.cursor++
		}
	case "enter":
		m.selIdx = m.cursor
		e := entries[m.selIdx]

		if e.fields == nil {
			// health — no form, go straight to running
			return m.startRunning(nil)
		}

		// build text inputs
		m.inputs = make([]textinput.Model, len(e.fields))
		for i, f := range e.fields {
			ti := textinput.New()
			ti.Placeholder = f.placeholder
			ti.SetValue(f.defaultVal)
			ti.Width = 54
			ti.PromptStyle = labelStyle
			ti.TextStyle = lipgloss.NewStyle().Foreground(white)
			m.inputs[i] = ti
		}
		m.focus = 0
		m.inputs[0].Focus()
		m.state = stateForm
	}
	return m, nil
}

func (m Model) updateForm(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.state = stateMenu
		return m, nil
	case "tab", "shift+tab":
		if len(m.inputs) > 1 {
			m.inputs[m.focus].Blur()
			if k.String() == "tab" {
				m.focus = (m.focus + 1) % len(m.inputs)
			} else {
				m.focus = (m.focus - 1 + len(m.inputs)) % len(m.inputs)
			}
			m.inputs[m.focus].Focus()
		}
		return m, nil
	case "enter":
		if m.focus < len(m.inputs)-1 {
			// advance to next field
			m.inputs[m.focus].Blur()
			m.focus++
			m.inputs[m.focus].Focus()
			return m, nil
		}
		return m.startRunning(m.inputs)
	}
	// pass to focused input
	var cmd tea.Cmd
	m.inputs[m.focus], cmd = m.inputs[m.focus].Update(k)
	return m, cmd
}

func (m Model) updateResult(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "q", "esc":
		m.lines = nil
		m.state = stateMenu
		return m, nil
	case "j", "down":
		m.vp.LineDown(3)
	case "k", "up":
		m.vp.LineUp(3)
	case "g":
		m.vp.GotoTop()
	case "G":
		m.vp.GotoBottom()
	}
	return m, nil
}

// startRunning wires up the subprocess and transitions to stateRunning.
func (m Model) startRunning(inputs []textinput.Model) (tea.Model, tea.Cmd) {
	m.lines = nil
	m.lineCh = make(chan string, 512)
	m.vp = viewport.New(m.width-4, m.height-9)
	m.state = stateRunning

	if entries[m.selIdx].fields == nil {
		go m.runHealth()
	} else {
		go m.runPython(inputs)
	}
	return m, tea.Batch(m.sp.Tick, waitLine(m.lineCh))
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m Model) View() string {
	switch m.state {
	case stateMenu:
		return m.viewMenu()
	case stateForm:
		return m.viewForm()
	case stateRunning:
		return m.viewStream(false)
	case stateResult:
		return m.viewStream(true)
	}
	return ""
}

func (m Model) viewMenu() string {
	var b strings.Builder
	banner := titleStyle.Render("  ✦ DigiStock — NSE Trading Bot ✦  ")
	b.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, banner))
	b.WriteString("\n\n")

	for i, e := range entries {
		var line string
		if i == m.cursor {
			line = cursorStyle.Render("  ❯ ") +
				selectedStyle.Render(e.icon+" "+e.label) +
				"   " + descStyle.Render(e.desc)
		} else {
			line = "    " + normalStyle.Render(e.icon+" "+e.label) +
				"   " + descStyle.Render(e.desc)
		}
		b.WriteString(line + "\n")
	}

	b.WriteString("\n")
	b.WriteString(hintStyle.Render("  ↑/↓  navigate   ↵  select   q  quit"))
	return b.String()
}

func (m Model) viewForm() string {
	e := entries[m.selIdx]
	var b strings.Builder

	banner := titleStyle.Render("  ✦ " + e.icon + " " + e.label + "  ")
	b.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, banner))
	b.WriteString("\n")

	// Context note
	if e.note != "" {
		noteStyle := lipgloss.NewStyle().Foreground(dim).Width(m.width - 4)
		b.WriteString(noteStyle.Render("  " + e.note))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	for i, ti := range m.inputs {
		f := e.fields[i]
		active := i == m.focus

		// Field label + active indicator
		if active {
			b.WriteString(labelStyle.Render("  ▶ "+f.label+":") + "\n")
		} else {
			b.WriteString(dimStyle.Render("    "+f.label+":") + "\n")
		}

		// Text input
		if active {
			b.WriteString("    " + ti.View() + "\n")
		} else {
			b.WriteString(dimStyle.Render("    "+ti.Value()) + "\n")
		}

		// Hint
		if f.hint != "" {
			b.WriteString(dimStyle.Render("    ↳  "+f.hint) + "\n")
		}

		// Examples
		if f.examples != "" {
			b.WriteString(dimStyle.Render("    eg: "+f.examples) + "\n")
		}

		b.WriteString("\n")
	}

	hint := "  ↵  run   Esc  back"
	if len(m.inputs) > 1 {
		hint = "  Tab  next field   ↵  run   Esc  back"
	}
	b.WriteString(hintStyle.Render(hint))
	return b.String()
}

func (m Model) viewStream(done bool) string {
	e := entries[m.selIdx]
	var b strings.Builder

	var statusBadge string
	if done {
		statusBadge = "  " + successStyle.Render("✔ Done")
	} else {
		statusBadge = "  " + m.sp.View()
	}
	banner := titleStyle.Render("  ✦ " + e.icon + " " + e.label + statusBadge + "  ")
	b.WriteString(lipgloss.PlaceHorizontal(m.width, lipgloss.Center, banner))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", m.width-2)) + "\n")
	b.WriteString(m.vp.View() + "\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", m.width-2)) + "\n")

	if done {
		b.WriteString(hintStyle.Render("  j/k  scroll   g  top   G  bottom   q  back to menu"))
	} else {
		b.WriteString(hintStyle.Render("  Ctrl+C  quit"))
	}
	return b.String()
}

// ── Subprocess Runners ────────────────────────────────────────────────────────

func waitLine(ch chan string) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return doneMsg{}
		}
		return lineMsg(line)
	}
}

func (m Model) runPython(inputs []textinput.Model) {
	root := m.cfg.Root
	python := m.cfg.Python
	var pyArgs []string

	switch m.selIdx {
	case 0: // analyze
		query := strings.TrimSpace(inputs[0].Value())
		if query == "" {
			query = "Analyze current NSE market and recommend 2 stocks to trade today."
		}
		pyArgs = []string{filepath.Join(root, "runner.py"), query}
	case 1: // indicators
		pyArgs = []string{filepath.Join(root, "runner_tools.py"), "indicators", normTicker(inputs[0].Value())}
	case 2: // predict
		days := strings.TrimSpace(inputs[1].Value())
		if days == "" {
			days = "5"
		}
		pyArgs = []string{filepath.Join(root, "runner_tools.py"), "predict", normTicker(inputs[0].Value()), days}
	case 3: // backtest
		strat := strings.TrimSpace(inputs[1].Value())
		if strat == "" {
			strat = "sma_cross"
		}
		pyArgs = []string{filepath.Join(root, "runner_tools.py"), "backtest", normTicker(inputs[0].Value()), strat}
	}

	cmd := exec.Command(python, pyArgs...)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		m.lineCh <- "ERR: " + err.Error()
		close(m.lineCh)
		return
	}
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		m.lineCh <- "ERR: Failed to start Python: " + err.Error()
		close(m.lineCh)
		return
	}

	sc := bufio.NewScanner(stdout)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		if line := sc.Text(); line != "" {
			m.lineCh <- line
		}
	}

	if err := cmd.Wait(); err != nil {
		if s := strings.TrimSpace(stderrBuf.String()); s != "" {
			for _, l := range strings.Split(s, "\n") {
				if l = strings.TrimSpace(l); l != "" {
					m.lineCh <- "ERR: " + l
				}
			}
		}
	}
	close(m.lineCh)
}

func (m Model) runHealth() {
	root := m.cfg.Root
	python := m.cfg.Python

	send := func(line string) { m.lineCh <- line }
	check := func(label string, ok bool, detail string) {
		if ok {
			send(fmt.Sprintf("✔  %-32s %s", label, detail))
		} else {
			send(fmt.Sprintf("✗  %-32s %s", label, detail))
		}
	}

	send("🔍 DigiStock Environment Check")
	send(strings.Repeat("─", 52))

	check("Project root found", true, root)

	for _, f := range []string{"runner.py", "runner_tools.py", "agent_logic.py"} {
		_, err := os.Stat(filepath.Join(root, f))
		check(f, err == nil, "")
	}

	out, err := exec.Command(python, "--version").CombinedOutput()
	check("Python interpreter", err == nil, strings.TrimSpace(string(out)))

	_, err = os.Stat(filepath.Join(root, ".env"))
	check(".env file", err == nil, "")

	for _, key := range []string{"OPENAI_API_KEY", "BRIGHT_DATA_API_TOKEN"} {
		val := os.Getenv(key)
		if val != "" {
			check(key, true, "set")
		} else {
			check(key, false, "NOT SET — add to .env")
		}
	}
	for _, key := range []string{"WEB_UNLOCKER_ZONE", "BROWSER_ZONE"} {
		val := os.Getenv(key)
		if val != "" {
			send(fmt.Sprintf("✔  %-32s %s", key, val))
		} else {
			send(fmt.Sprintf("⚠  %-32s (optional, using default)", key))
		}
	}

	send(strings.Repeat("─", 52))
	close(m.lineCh)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func normTicker(t string) string {
	t = strings.ToUpper(strings.TrimSpace(t))
	if !strings.HasSuffix(t, ".NS") {
		t += ".NS"
	}
	return t
}

func renderLines(lines []string) string {
	out := make([]string, len(lines))
	for i, line := range lines {
		out[i] = colorLine(line)
	}
	return strings.Join(out, "\n")
}

func colorLine(line string) string {
	switch {
	case strings.Contains(line, "Calling Sub-Agent:"):
		return lipgloss.NewStyle().Foreground(cyan).Bold(true).Render(line)
	case strings.Contains(line, "Responding as Supervisor:"):
		return lipgloss.NewStyle().Foreground(green).Bold(true).Render(line)
	case strings.Contains(line, "Response from"):
		return lipgloss.NewStyle().Foreground(magenta).Render(line)
	case strings.HasPrefix(line, "✔"), strings.HasPrefix(line, "✓"):
		return lipgloss.NewStyle().Foreground(green).Render(line)
	case strings.HasPrefix(line, "✗"):
		return lipgloss.NewStyle().Foreground(red).Bold(true).Render(line)
	case strings.HasPrefix(line, "⚠"):
		return lipgloss.NewStyle().Foreground(yellow).Render(line)
	case strings.HasPrefix(line, "ERR:"):
		return lipgloss.NewStyle().Foreground(red).Bold(true).Render(line)
	case strings.HasPrefix(line, "**") && strings.HasSuffix(line, "**"):
		return lipgloss.NewStyle().Foreground(yellow).Bold(true).Render(line)
	case strings.HasPrefix(line, "📊"), strings.HasPrefix(line, "📈"), strings.HasPrefix(line, "🔮"), strings.HasPrefix(line, "🔍"):
		return lipgloss.NewStyle().Foreground(cyan).Bold(true).Render(line)
	case strings.HasPrefix(line, "─"):
		return dimStyle.Render(line)
	default:
		return lipgloss.NewStyle().Foreground(white).Render(line)
	}
}

// Start launches the fullscreen TUI program.
func Start(cfg Config) error {
	p := tea.NewProgram(New(cfg), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
