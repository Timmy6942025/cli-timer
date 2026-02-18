package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	defaultFont              = "Standard"
	defaultTickRateMs        = 100
	minTickRateMs            = 50
	maxTickRateMs            = 1000
	defaultCompletionMessage = "Time is up!"
)

type keybindings struct {
	PauseKey    string `json:"pauseKey"`
	PauseAltKey string `json:"pauseAltKey"`
	RestartKey  string `json:"restartKey"`
	ExitKey     string `json:"exitKey"`
	ExitAltKey  string `json:"exitAltKey"`
}

var defaultKeybindings = keybindings{
	PauseKey:    "p",
	PauseAltKey: "space",
	RestartKey:  "r",
	ExitKey:     "q",
	ExitAltKey:  "e",
}

type config struct {
	Font              string      `json:"font"`
	CenterDisplay     bool        `json:"centerDisplay"`
	ShowHeader        bool        `json:"showHeader"`
	ShowControls      bool        `json:"showControls"`
	TickRateMs        int         `json:"tickRateMs"`
	CompletionMessage string      `json:"completionMessage"`
	NotifyOnComplete  bool        `json:"notifyOnComplete"`
	Keybindings       keybindings `json:"keybindings"`
}

type statePayload struct {
	ConfigPath string   `json:"configPath"`
	Config     config   `json:"config"`
	Fonts      []string `json:"fonts"`
}

type menuEntry struct {
	id          string
	title       string
	description string
}

func (m menuEntry) Title() string       { return m.title }
func (m menuEntry) Description() string { return m.description }
func (m menuEntry) FilterValue() string { return m.title + " " + m.description }

type fontEntry struct {
	name string
}

func (f fontEntry) Title() string       { return f.name }
func (f fontEntry) Description() string { return "Press Enter to select" }
func (f fontEntry) FilterValue() string { return f.name }

type keyEntry struct {
	token string
}

func (k keyEntry) Title() string       { return keyTokenLabel(k.token) }
func (k keyEntry) Description() string { return "Press Enter to select" }
func (k keyEntry) FilterValue() string { return k.token + " " + keyTokenLabel(k.token) }

type screen int

const (
	screenMain screen = iota
	screenFontPicker
	screenKeyPicker
	screenTickRateEditor
	screenMessageEditor
)

type model struct {
	payload      statePayload
	menu         list.Model
	fontList     list.Model
	keyList      list.Model
	tickInput    textinput.Model
	messageInput textinput.Model
	screen       screen
	keyTarget    string
	quitting     bool
	cancelled    bool
	err          error
}

func boolText(v bool) string {
	if v {
		return "On"
	}
	return "Off"
}

func keyTokenLabel(token string) string {
	if token == "space" {
		return "Spacebar"
	}
	return token
}

func summarizeMessage(text string) string {
	if strings.TrimSpace(text) == "" {
		return "(empty)"
	}
	compact := strings.ReplaceAll(strings.ReplaceAll(text, "\r", ""), "\n", " ")
	if len(compact) > 44 {
		return compact[:41] + "..."
	}
	return compact
}

func supportedKeyTokens() []string {
	tokens := []string{"space"}
	for ch := 'a'; ch <= 'z'; ch++ {
		tokens = append(tokens, string(ch))
	}
	for ch := '0'; ch <= '9'; ch++ {
		tokens = append(tokens, string(ch))
	}
	tokens = append(tokens, "`", "-", "=", "[", "]", "\\", ";", "'", ",", ".", "/")
	return tokens
}

func buildMenuItems(cfg config) []list.Item {
	return []list.Item{
		menuEntry{id: "font", title: "Font", description: cfg.Font},
		menuEntry{id: "center", title: "Center display", description: boolText(cfg.CenterDisplay)},
		menuEntry{id: "header", title: "Show header", description: boolText(cfg.ShowHeader)},
		menuEntry{id: "controls", title: "Show controls", description: boolText(cfg.ShowControls)},
		menuEntry{id: "tickRate", title: "Tick rate", description: fmt.Sprintf("%d ms", cfg.TickRateMs)},
		menuEntry{id: "message", title: "Completion message", description: summarizeMessage(cfg.CompletionMessage)},
		menuEntry{id: "notify", title: "System notification", description: boolText(cfg.NotifyOnComplete)},
		menuEntry{id: "pauseKey", title: "Pause key", description: keyTokenLabel(cfg.Keybindings.PauseKey)},
		menuEntry{id: "pauseAltKey", title: "Pause alt key", description: keyTokenLabel(cfg.Keybindings.PauseAltKey)},
		menuEntry{id: "restartKey", title: "Restart key", description: keyTokenLabel(cfg.Keybindings.RestartKey)},
		menuEntry{id: "exitKey", title: "Exit key", description: keyTokenLabel(cfg.Keybindings.ExitKey)},
		menuEntry{id: "exitAltKey", title: "Exit alt key", description: keyTokenLabel(cfg.Keybindings.ExitAltKey)},
		menuEntry{id: "save", title: "Save and exit", description: "Write settings and close"},
		menuEntry{id: "cancel", title: "Cancel", description: "Discard changes"},
	}
}

func buildFontItems(fonts []string) []list.Item {
	items := make([]list.Item, 0, len(fonts))
	for _, font := range fonts {
		items = append(items, fontEntry{name: font})
	}
	return items
}

func buildKeyItems() []list.Item {
	items := make([]list.Item, 0, len(supportedKeyTokens()))
	for _, token := range supportedKeyTokens() {
		items = append(items, keyEntry{token: token})
	}
	return items
}

func sanitizeTickRate(value int) int {
	if value < minTickRateMs {
		return minTickRateMs
	}
	if value > maxTickRateMs {
		return maxTickRateMs
	}
	return value
}

func normalizeCompletionMessage(value string) string {
	compact := strings.ReplaceAll(strings.ReplaceAll(value, "\r", ""), "\n", " ")
	if len(compact) > 240 {
		return compact[:240]
	}
	return compact
}

func validKeyToken(value string) bool {
	if value == "space" {
		return true
	}
	if len(value) != 1 {
		return false
	}
	ch := value[0]
	return ch >= 33 && ch <= 126
}

func normalizeKeyToken(value, fallback string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if !validKeyToken(normalized) {
		return fallback
	}
	return normalized
}

func normalizeKeybindings(cfg keybindings) keybindings {
	result := defaultKeybindings
	result.PauseKey = normalizeKeyToken(cfg.PauseKey, result.PauseKey)
	result.PauseAltKey = normalizeKeyToken(cfg.PauseAltKey, result.PauseAltKey)
	result.RestartKey = normalizeKeyToken(cfg.RestartKey, result.RestartKey)
	result.ExitKey = normalizeKeyToken(cfg.ExitKey, result.ExitKey)
	result.ExitAltKey = normalizeKeyToken(cfg.ExitAltKey, result.ExitAltKey)
	return result
}

func normalizeConfig(cfg config) config {
	result := config{
		Font:              defaultFont,
		CenterDisplay:     true,
		ShowHeader:        true,
		ShowControls:      true,
		TickRateMs:        defaultTickRateMs,
		CompletionMessage: defaultCompletionMessage,
		NotifyOnComplete:  true,
		Keybindings:       defaultKeybindings,
	}

	if strings.TrimSpace(cfg.Font) != "" {
		result.Font = cfg.Font
	}
	result.CenterDisplay = cfg.CenterDisplay
	result.ShowHeader = cfg.ShowHeader
	result.ShowControls = cfg.ShowControls
	if cfg.TickRateMs != 0 {
		result.TickRateMs = sanitizeTickRate(cfg.TickRateMs)
	}
	if cfg.CompletionMessage != "" {
		result.CompletionMessage = normalizeCompletionMessage(cfg.CompletionMessage)
	}
	result.NotifyOnComplete = cfg.NotifyOnComplete
	result.Keybindings = normalizeKeybindings(cfg.Keybindings)
	return result
}

func isConfirmKey(msg tea.KeyMsg) bool {
	if msg.Type == tea.KeyEnter || msg.Type == tea.KeyCtrlM || msg.Type == tea.KeyCtrlJ {
		return true
	}
	if len(msg.Runes) == 1 {
		if msg.Runes[0] == '\r' || msg.Runes[0] == '\n' {
			return true
		}
	}
	s := msg.String()
	return s == "enter" || s == "ctrl+m" || s == "ctrl+j" || s == "return"
}

func isBackKey(msg tea.KeyMsg) bool {
	s := msg.String()
	return s == "esc" || s == "q"
}

func isQuitKey(msg tea.KeyMsg) bool {
	s := msg.String()
	return s == "ctrl+c" || s == "q"
}

func isSaveKey(msg tea.KeyMsg) bool {
	s := msg.String()
	return s == "ctrl+s"
}

func newModel(payload statePayload) model {
	menuModel := list.New(buildMenuItems(payload.Config), list.NewDefaultDelegate(), 0, 0)
	menuModel.Title = "Timer Settings"
	menuModel.SetShowHelp(true)
	menuModel.SetFilteringEnabled(false)
	menuModel.DisableQuitKeybindings()
	menuModel.SetShowStatusBar(false)
	menuModel.SetSize(100, 20)

	fontModel := list.New(buildFontItems(payload.Fonts), list.NewDefaultDelegate(), 0, 0)
	fontModel.Title = "Select Font"
	fontModel.SetShowHelp(true)
	fontModel.SetFilteringEnabled(true)
	fontModel.DisableQuitKeybindings()
	fontModel.SetSize(100, 20)

	keyModel := list.New(buildKeyItems(), list.NewDefaultDelegate(), 0, 0)
	keyModel.Title = "Select Key"
	keyModel.SetShowHelp(true)
	keyModel.SetFilteringEnabled(true)
	keyModel.DisableQuitKeybindings()
	keyModel.SetSize(100, 20)

	tickInput := textinput.New()
	tickInput.Prompt = "Tick rate (ms): "
	tickInput.CharLimit = 4
	tickInput.SetValue(strconv.Itoa(payload.Config.TickRateMs))
	tickInput.Blur()

	messageInput := textinput.New()
	messageInput.Prompt = "Completion message: "
	messageInput.CharLimit = 240
	messageInput.SetValue(payload.Config.CompletionMessage)
	messageInput.Blur()

	return model{
		payload:      payload,
		menu:         menuModel,
		fontList:     fontModel,
		keyList:      keyModel,
		tickInput:    tickInput,
		messageInput: messageInput,
		screen:       screenMain,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) refreshMenu() {
	m.menu.SetItems(buildMenuItems(m.payload.Config))
}

func (m *model) save() error {
	if m.payload.ConfigPath == "" {
		return errors.New("config path is missing")
	}
	text, err := json.MarshalIndent(m.payload.Config, "", "  ")
	if err != nil {
		return err
	}
	text = append(text, '\n')
	return os.WriteFile(m.payload.ConfigPath, text, 0644)
}

func (m *model) selectFontItem(font string) {
	for idx, item := range m.fontList.Items() {
		entry, ok := item.(fontEntry)
		if ok && entry.name == font {
			m.fontList.Select(idx)
			return
		}
	}
}

func (m *model) selectKeyItem(token string) {
	for idx, item := range m.keyList.Items() {
		entry, ok := item.(keyEntry)
		if ok && entry.token == token {
			m.keyList.Select(idx)
			return
		}
	}
}

func (m *model) keyTokenForTarget(target string) string {
	switch target {
	case "pauseKey":
		return m.payload.Config.Keybindings.PauseKey
	case "pauseAltKey":
		return m.payload.Config.Keybindings.PauseAltKey
	case "restartKey":
		return m.payload.Config.Keybindings.RestartKey
	case "exitKey":
		return m.payload.Config.Keybindings.ExitKey
	case "exitAltKey":
		return m.payload.Config.Keybindings.ExitAltKey
	default:
		return defaultKeybindings.PauseKey
	}
}

func (m *model) setKeyTokenForTarget(target string, token string) {
	switch target {
	case "pauseKey":
		m.payload.Config.Keybindings.PauseKey = token
	case "pauseAltKey":
		m.payload.Config.Keybindings.PauseAltKey = token
	case "restartKey":
		m.payload.Config.Keybindings.RestartKey = token
	case "exitKey":
		m.payload.Config.Keybindings.ExitKey = token
	case "exitAltKey":
		m.payload.Config.Keybindings.ExitAltKey = token
	}
}

func (m *model) saveAndQuit() tea.Cmd {
	if err := m.save(); err != nil {
		m.err = err
		return nil
	}
	m.quitting = true
	return tea.Quit
}

func (m *model) openKeyPicker(target string, title string) {
	m.keyTarget = target
	m.keyList.Title = title
	m.selectKeyItem(m.keyTokenForTarget(target))
	m.screen = screenKeyPicker
}

func (m *model) applyMenuAction() tea.Cmd {
	selected, ok := m.menu.SelectedItem().(menuEntry)
	if !ok {
		items := m.menu.Items()
		index := m.menu.Index()
		if index >= 0 && index < len(items) {
			selected, ok = items[index].(menuEntry)
		}
	}
	if !ok {
		return nil
	}
	m.err = nil

	switch selected.id {
	case "font":
		m.selectFontItem(m.payload.Config.Font)
		m.screen = screenFontPicker
		return nil
	case "center":
		m.payload.Config.CenterDisplay = !m.payload.Config.CenterDisplay
		m.refreshMenu()
		return nil
	case "header":
		m.payload.Config.ShowHeader = !m.payload.Config.ShowHeader
		m.refreshMenu()
		return nil
	case "controls":
		m.payload.Config.ShowControls = !m.payload.Config.ShowControls
		m.refreshMenu()
		return nil
	case "tickRate":
		m.tickInput.SetValue(strconv.Itoa(m.payload.Config.TickRateMs))
		m.tickInput.CursorEnd()
		m.tickInput.Focus()
		m.screen = screenTickRateEditor
		return nil
	case "message":
		m.messageInput.SetValue(m.payload.Config.CompletionMessage)
		m.messageInput.CursorEnd()
		m.messageInput.Focus()
		m.screen = screenMessageEditor
		return nil
	case "notify":
		m.payload.Config.NotifyOnComplete = !m.payload.Config.NotifyOnComplete
		m.refreshMenu()
		return nil
	case "pauseKey":
		m.openKeyPicker("pauseKey", "Select Pause Key")
		return nil
	case "pauseAltKey":
		m.openKeyPicker("pauseAltKey", "Select Pause Alt Key")
		return nil
	case "restartKey":
		m.openKeyPicker("restartKey", "Select Restart Key")
		return nil
	case "exitKey":
		m.openKeyPicker("exitKey", "Select Exit Key")
		return nil
	case "exitAltKey":
		m.openKeyPicker("exitAltKey", "Select Exit Alt Key")
		return nil
	case "save":
		return m.saveAndQuit()
	case "cancel":
		m.cancelled = true
		m.quitting = true
		return tea.Quit
	default:
		return nil
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.menu.SetSize(msg.Width, msg.Height-4)
		m.fontList.SetSize(msg.Width, msg.Height-4)
		m.keyList.SetSize(msg.Width, msg.Height-4)
		if msg.Width > 26 {
			m.tickInput.Width = msg.Width - 26
			m.messageInput.Width = msg.Width - 26
		}
		return m, nil
	case tea.KeyMsg:
		switch m.screen {
		case screenMain:
			if isQuitKey(msg) {
				m.cancelled = true
				m.quitting = true
				return m, tea.Quit
			}
			if isSaveKey(msg) {
				return m, m.saveAndQuit()
			}
			if isConfirmKey(msg) {
				cmd := m.applyMenuAction()
				return m, cmd
			}
		case screenFontPicker:
			if isBackKey(msg) {
				m.screen = screenMain
				return m, nil
			}
			if isConfirmKey(msg) {
				item, ok := m.fontList.SelectedItem().(fontEntry)
				if !ok {
					items := m.fontList.Items()
					index := m.fontList.Index()
					if index >= 0 && index < len(items) {
						item, ok = items[index].(fontEntry)
					}
				}
				if ok {
					m.payload.Config.Font = item.name
					m.screen = screenMain
					m.refreshMenu()
				}
				return m, nil
			}
		case screenKeyPicker:
			if isBackKey(msg) {
				m.screen = screenMain
				return m, nil
			}
			if isConfirmKey(msg) {
				item, ok := m.keyList.SelectedItem().(keyEntry)
				if !ok {
					items := m.keyList.Items()
					index := m.keyList.Index()
					if index >= 0 && index < len(items) {
						item, ok = items[index].(keyEntry)
					}
				}
				if ok {
					m.setKeyTokenForTarget(m.keyTarget, item.token)
					m.screen = screenMain
					m.refreshMenu()
				}
				return m, nil
			}
		case screenTickRateEditor:
			if isBackKey(msg) {
				m.tickInput.Blur()
				m.screen = screenMain
				return m, nil
			}
			if isConfirmKey(msg) {
				value, err := strconv.Atoi(strings.TrimSpace(m.tickInput.Value()))
				if err != nil {
					m.err = errors.New("tick rate must be an integer")
					return m, nil
				}
				if value < minTickRateMs || value > maxTickRateMs {
					m.err = fmt.Errorf("tick rate must be between %d and %d", minTickRateMs, maxTickRateMs)
					return m, nil
				}
				m.payload.Config.TickRateMs = value
				m.err = nil
				m.tickInput.Blur()
				m.screen = screenMain
				m.refreshMenu()
				return m, nil
			}
		case screenMessageEditor:
			if isBackKey(msg) {
				m.messageInput.Blur()
				m.screen = screenMain
				return m, nil
			}
			if isConfirmKey(msg) {
				m.payload.Config.CompletionMessage = normalizeCompletionMessage(m.messageInput.Value())
				m.err = nil
				m.messageInput.Blur()
				m.screen = screenMain
				m.refreshMenu()
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	switch m.screen {
	case screenMain:
		m.menu, cmd = m.menu.Update(msg)
	case screenFontPicker:
		m.fontList, cmd = m.fontList.Update(msg)
	case screenKeyPicker:
		m.keyList, cmd = m.keyList.Update(msg)
	case screenTickRateEditor:
		m.tickInput, cmd = m.tickInput.Update(msg)
	case screenMessageEditor:
		m.messageInput, cmd = m.messageInput.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		if m.err != nil {
			return fmt.Sprintf("Error: %v\n", m.err)
		}
		if m.cancelled {
			return "Cancelled\n"
		}
		return "Saved\n"
	}

	errorLine := ""
	if m.err != nil {
		errorLine = fmt.Sprintf("\nError: %v\n", m.err)
	}

	switch m.screen {
	case screenMain:
		return m.menu.View() + errorLine + "\nEnter: select/edit | Ctrl+S: save and exit | q: cancel | Ctrl+C: cancel"
	case screenFontPicker:
		return m.fontList.View() + "\nEnter: choose font | /: filter | esc: back"
	case screenKeyPicker:
		return m.keyList.View() + "\nEnter: choose key | /: filter | esc: back"
	case screenTickRateEditor:
		return fmt.Sprintf("Tick rate (%d-%d ms)\n\n%s%s\n\nEnter: save | esc: back", minTickRateMs, maxTickRateMs, m.tickInput.View(), errorLine)
	case screenMessageEditor:
		return fmt.Sprintf("Completion message\n\n%s%s\n\nEnter: save | esc: back", m.messageInput.View(), errorLine)
	default:
		return ""
	}
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func loadPayload(statePath string) (statePayload, error) {
	if statePath == "" {
		return statePayload{}, errors.New("--state is required")
	}
	text, err := os.ReadFile(statePath)
	if err != nil {
		return statePayload{}, err
	}
	var payload statePayload
	if err := json.Unmarshal(text, &payload); err != nil {
		return statePayload{}, err
	}
	if strings.TrimSpace(payload.ConfigPath) == "" {
		return statePayload{}, errors.New("configPath is missing in state payload")
	}
	if len(payload.Fonts) == 0 {
		payload.Fonts = []string{defaultFont}
	}
	payload.Config = normalizeConfig(payload.Config)
	if !containsString(payload.Fonts, payload.Config.Font) {
		payload.Config.Font = payload.Fonts[0]
	}
	return payload, nil
}

func main() {
	statePath := flag.String("state", "", "Path to JSON state file")
	flag.Parse()

	payload, err := loadPayload(*statePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load state: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(newModel(payload), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Settings UI failed: %v\n", err)
		os.Exit(1)
	}

	m := finalModel.(model)
	if m.err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save settings: %v\n", m.err)
		os.Exit(1)
	}
	if m.cancelled {
		os.Exit(2)
	}
}
