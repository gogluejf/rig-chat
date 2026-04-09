package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"rig-chat/internal/chat"
	"rig-chat/internal/config"
	"rig-chat/internal/ui"
)

// streamEventMsg wraps a StreamEvent for the Bubble Tea message loop
type streamEventMsg chat.StreamEvent

// modelsLoadedMsg signals that model scanning completed
type modelsLoadedMsg struct {
	models []chat.ModelEntry
}

// Model is the top-level Bubble Tea model
type Model struct {
	// UI components
	textarea textarea.Model
	viewport viewport.Model
	mode     Mode
	ready    bool

	// Command palette
	cmdPalette ui.CommandPalette
	cmdInput   string // text after / in command mode

	// Pickers
	modelPicker    ui.PickerList
	sessionPicker  ui.PickerList
	filePicker     ui.PickerList
	thinkingToggle ui.ThinkingToggle
	savePrompt     ui.SavePrompt
	filePickerFor  string // "image" or "system"

	// Chat state
	messages       []config.DisplayMessage
	streamText     strings.Builder
	streamThinking strings.Builder
	inThinking     bool
	streaming      bool
	cancelFn       context.CancelFunc
	streamCh       <-chan chat.StreamEvent
	streamStart    time.Time
	firstTokenTime time.Time
	tokenCount     int

	// Config
	settings  config.Settings
	endpoints config.EndpointsConfig
	paths     config.Paths
	history   config.History
	session   config.SessionFile

	// Model cache
	modelCache *chat.ModelCache

	// Prompt history navigation
	historyIdx int // -1 = draft, 0..n = browsing history
	draft      string

	// Dimensions
	width  int
	height int

	// Image attachment
	attachedImage string

	// Total tokens across session
	totalTokens int

	// Error display
	lastError string
}

// New creates a new app Model
func New(paths config.Paths, settings config.Settings, endpoints config.EndpointsConfig, history config.History) Model {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.ShowLineNumbers = false
	ta.SetHeight(4)
	ta.Focus()
	ta.CharLimit = 0

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Base = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)
	ta.BlurredStyle.Base = ta.FocusedStyle.Base

	vp := viewport.New(80, 20)
	session := config.NewSessionFile(settings.Provider, settings.Model, settings.Thinking, settings.SystemPromptFile)

	return Model{
		textarea:   ta,
		viewport:   vp,
		mode:       ModeChat,
		settings:   settings,
		endpoints:  endpoints,
		paths:      paths,
		history:    history,
		session:    session,
		historyIdx: -1,
		cmdPalette: ui.NewCommandPalette(),
		modelCache: chat.NewModelCache(5 * time.Minute),
	}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func waitForStreamEvent(ch <-chan chat.StreamEvent) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-ch
		if !ok {
			return streamEventMsg(chat.StreamEvent{Done: true})
		}
		return streamEventMsg(event)
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.recalcLayout()
		m.updateViewportContent()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case streamEventMsg:
		return m.handleStreamEvent(chat.StreamEvent(msg))

	case modelsLoadedMsg:
		ids := chat.ModelIDs(msg.models)
		m.modelPicker = ui.NewPickerList("Select Model", ids)
		m.mode = ModeModelPicker
		return m, nil
	}

	if m.mode == ModeChat {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) recalcLayout() {
	inputHeight := 6
	footerHeight := 1
	overlayHeight := 0

	switch m.mode {
	case ModeCommand:
		overlayHeight = len(m.cmdPalette.FilteredItems()) + 1
		if overlayHeight > 10 {
			overlayHeight = 10
		}
	}

	attachHeight := 0
	if m.attachedImage != "" {
		attachHeight = 1
	}
	vpHeight := m.height - inputHeight - footerHeight - attachHeight - overlayHeight
	if vpHeight < 3 {
		vpHeight = 3
	}
	m.viewport.Width = m.width
	m.viewport.Height = vpHeight
	m.textarea.SetWidth(m.width)
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {

	case ModeChat:
		return m.handleChatKey(msg)

	case ModeStreaming:
		return m.handleStreamingKey(msg)

	case ModeHelp:
		if key.Matches(msg, keys.Help) || key.Matches(msg, keys.Cancel) || key.Matches(msg, keys.Escape) {
			m.mode = ModeChat
			m.textarea.Focus()
			return m, nil
		}

	case ModeCommand:
		return m.handleCommandKey(msg)

	case ModeModelPicker:
		return m.handlePickerKey(msg, "model")

	case ModeSessionPicker:
		return m.handlePickerKey(msg, "session")

	case ModeFilePicker:
		return m.handlePickerKey(msg, m.filePickerFor)
	}

	return m, nil
}

func (m Model) handleChatKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel):
		_ = config.SaveHistory(m.paths, m.history)
		return m, tea.Quit

	case key.Matches(msg, keys.Help):
		m.mode = ModeHelp
		return m, nil

	case key.Matches(msg, keys.ExpandThinking):
		m.toggleLastThinking()
		m.updateViewportContent()
		return m, nil

	case key.Matches(msg, keys.Save):
		return m.startSave()

	case key.Matches(msg, keys.Load):
		return m.startLoad()

	case key.Matches(msg, keys.Send):
		return m.sendMessage()

	case key.Matches(msg, keys.Up):
		return m.historyUp()

	case key.Matches(msg, keys.Down):
		return m.historyDown()

	default:
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)

		// Check for slash command trigger: "/" as the only content
		val := m.textarea.Value()
		if val == "/" {
			m.textarea.SetValue("")
			m.cmdPalette = ui.NewCommandPalette()
			m.cmdPalette.Visible = true
			m.cmdInput = ""
			m.mode = ModeCommand
			m.recalcLayout()
			return m, nil
		}

		return m, cmd
	}
}

func (m Model) handleStreamingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Cancel):
		if m.cancelFn != nil {
			m.cancelFn()
		}
		if len(m.messages) > 0 && m.messages[len(m.messages)-1].Role == "user" {
			m.messages = m.messages[:len(m.messages)-1]
		}
		// Remove from prompt history too
		if len(m.history.Entries) > 0 {
			m.history.Entries = m.history.Entries[:len(m.history.Entries)-1]
			_ = config.SaveHistory(m.paths, m.history)
		}
		m.streaming = false
		m.mode = ModeChat
		m.textarea.Focus()
		m.streamText.Reset()
		m.streamThinking.Reset()
		m.updateViewportContent()
		return m, nil

	case key.Matches(msg, keys.ExpandThinking):
		// Could toggle live thinking visibility in the future
		return m, nil
	}
	return m, nil
}

func (m Model) handleCommandKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape), key.Matches(msg, keys.Cancel):
		m.cmdPalette.Reset()
		m.mode = ModeChat
		m.textarea.Focus()
		m.recalcLayout()
		return m, nil

	case key.Matches(msg, keys.Up):
		m.cmdPalette.MoveUp()
		return m, nil

	case key.Matches(msg, keys.Down):
		m.cmdPalette.MoveDown()
		return m, nil

	case key.Matches(msg, keys.Send):
		return m.executeCommand(m.cmdPalette.SelectedCommand())

	default:
		// Type to filter
		s := msg.String()
		if len(s) == 1 {
			m.cmdInput += s
			m.cmdPalette.Filter = m.cmdInput
			m.cmdPalette.Selected = 0
		} else if s == "backspace" && len(m.cmdInput) > 0 {
			m.cmdInput = m.cmdInput[:len(m.cmdInput)-1]
			m.cmdPalette.Filter = m.cmdInput
			m.cmdPalette.Selected = 0
		}
		m.recalcLayout()
		return m, nil
	}
}

func (m Model) handlePickerKey(msg tea.KeyMsg, pickerType string) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape), key.Matches(msg, keys.Cancel):
		m.mode = ModeChat
		m.textarea.Focus()
		return m, nil

	case key.Matches(msg, keys.Up):
		switch pickerType {
		case "model":
			m.modelPicker.MoveUp()
		case "session":
			m.sessionPicker.MoveUp()
		case "image", "system":
			m.filePicker.MoveUp()
		}
		return m, nil

	case key.Matches(msg, keys.Down):
		switch pickerType {
		case "model":
			m.modelPicker.MoveDown()
		case "session":
			m.sessionPicker.MoveDown()
		case "image", "system":
			m.filePicker.MoveDown()
		}
		return m, nil

	case key.Matches(msg, keys.Send):
		return m.confirmPicker(pickerType)

	case key.Matches(msg, keys.Tab):
		switch pickerType {
		case "model":
			m.modelPicker.MoveDown()
		case "session":
			m.sessionPicker.MoveDown()
		case "image", "system":
			m.filePicker.MoveDown()
		}
		return m, nil

	default:
		// Type to filter
		s := msg.String()
		switch pickerType {
		case "model":
			if len(s) == 1 {
				m.modelPicker.Filter += s
				m.modelPicker.Selected = 0
			} else if s == "backspace" && len(m.modelPicker.Filter) > 0 {
				m.modelPicker.Filter = m.modelPicker.Filter[:len(m.modelPicker.Filter)-1]
				m.modelPicker.Selected = 0
			}
		case "session":
			if len(s) == 1 {
				m.sessionPicker.Filter += s
				m.sessionPicker.Selected = 0
			} else if s == "backspace" && len(m.sessionPicker.Filter) > 0 {
				m.sessionPicker.Filter = m.sessionPicker.Filter[:len(m.sessionPicker.Filter)-1]
				m.sessionPicker.Selected = 0
			}
		case "image", "system":
			if len(s) == 1 {
				m.filePicker.Filter += s
				m.filePicker.Selected = 0
			} else if s == "backspace" && len(m.filePicker.Filter) > 0 {
				m.filePicker.Filter = m.filePicker.Filter[:len(m.filePicker.Filter)-1]
				m.filePicker.Selected = 0
			}
		}
		return m, nil
	}
}

func (m Model) confirmPicker(pickerType string) (tea.Model, tea.Cmd) {
	switch pickerType {
	case "model":
		selected := m.modelPicker.SelectedItem()
		if selected != "" {
			m.settings.Model = selected
			// Find the provider for this model
			if models, ok := m.modelCache.Get(); ok {
				for _, me := range models {
					if me.ID == selected {
						m.settings.Provider = me.Provider
						break
					}
				}
			}
			_ = config.SaveSettings(m.paths, m.settings)
		}

	case "session":
		selected := m.sessionPicker.SelectedItem()
		if selected != "" {
			sf, err := config.LoadSession(m.paths, selected)
			if err == nil {
				m.session = sf
				m.messages = make([]config.DisplayMessage, len(sf.Messages))
				for i, msg := range sf.Messages {
					m.messages[i] = config.DisplayMessage{Message: msg}
				}
				m.settings.Model = sf.Session.Model
				m.settings.Provider = sf.Session.Provider
				m.settings.Thinking = sf.Session.Thinking
				m.settings.LastSessionName = selected
				_ = config.SaveSettings(m.paths, m.settings)
				m.updateViewportContent()
			} else {
				m.lastError = fmt.Sprintf("load session: %v", err)
			}
		}

	case "image":
		selected := m.filePicker.SelectedItem()
		if selected != "" {
			m.attachedImage = selected
			m.recalcLayout()
		}

	case "system":
		selected := m.filePicker.SelectedItem()
		if selected != "" {
			m.settings.SystemPromptFile = selected
			_ = config.SaveSettings(m.paths, m.settings)
		}
	}

	m.mode = ModeChat
	m.textarea.Focus()
	return m, nil
}

func (m Model) executeCommand(name string) (tea.Model, tea.Cmd) {
	m.cmdPalette.Reset()

	switch name {
	case "exit":
		_ = config.SaveHistory(m.paths, m.history)
		return m, tea.Quit

	case "help":
		m.mode = ModeHelp
		return m, nil

	case "model":
		// Scan models asynchronously
		m.mode = ModeChat // temporarily back to chat while loading
		return m, m.scanModelsCmd()

	case "thinking":
		m.thinkingToggle = ui.NewThinkingToggle(m.settings.Thinking)
		m.mode = ModeChat
		// Simple toggle for now
		m.settings.Thinking = !m.settings.Thinking
		_ = config.SaveSettings(m.paths, m.settings)
		m.textarea.Focus()
		return m, nil

	case "image":
		// List image files — for now just let user type a path
		m.filePicker = ui.NewPickerList("Attach Image (type path)", []string{})
		m.filePickerFor = "image"
		m.mode = ModeFilePicker
		return m, nil

	case "save":
		return m.startSave()

	case "load":
		return m.startLoad()

	case "system":
		prompts := config.ListSystemPrompts(m.paths)
		m.filePicker = ui.NewPickerList("System Prompt", prompts)
		m.filePickerFor = "system"
		m.mode = ModeFilePicker
		return m, nil
	}

	m.mode = ModeChat
	m.textarea.Focus()
	return m, nil
}

func (m Model) startSave() (Model, tea.Cmd) {
	name := m.settings.LastSessionName
	if name == "" {
		name = time.Now().Format("2006-01-02_15-04")
	}
	m.session.Messages = m.extractSessionMessages()
	err := config.SaveSession(m.paths, name, m.session)
	if err != nil {
		m.lastError = fmt.Sprintf("save: %v", err)
	} else {
		m.settings.LastSessionName = name
		_ = config.SaveSettings(m.paths, m.settings)
		m.lastError = "" // clear any previous error
	}
	return m, nil
}

func (m Model) startLoad() (Model, tea.Cmd) {
	sessions := config.ListSessions(m.paths)
	m.sessionPicker = ui.NewPickerList("Load Session", sessions)
	m.mode = ModeSessionPicker
	return m, nil
}

func (m Model) scanModelsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		models := chat.ScanModels(ctx, m.endpoints, m.modelCache)
		return modelsLoadedMsg{models: models}
	}
}

func (m Model) sendMessage() (tea.Model, tea.Cmd) {
	text := strings.TrimSpace(m.textarea.Value())
	if text == "" {
		return m, nil
	}

	config.AddHistoryEntry(&m.history, text, m.settings.MaxHistory)
	_ = config.SaveHistory(m.paths, m.history)
	m.historyIdx = -1
	m.draft = ""

	userMsg := config.DisplayMessage{
		Message: config.Message{
			ID:        fmt.Sprintf("msg_%d", len(m.messages)+1),
			Role:      "user",
			CreatedAt: time.Now(),
			Text:      text,
			ImagePath: m.attachedImage,
		},
	}
	m.messages = append(m.messages, userMsg)

	m.textarea.SetValue("")
	m.textarea.Blur()

	apiMsgs := m.buildAPIMessages()
	m.attachedImage = ""

	m.streaming = true
	m.mode = ModeStreaming
	m.streamText.Reset()
	m.streamThinking.Reset()
	m.inThinking = false
	m.tokenCount = 0
	m.streamStart = time.Now()
	m.firstTokenTime = time.Time{}
	m.lastError = ""

	chatURL := "http://localhost/v1/chat/completions"
	for _, p := range m.endpoints.Providers {
		if p.Name == m.settings.Provider {
			chatURL = p.ChatURL
			break
		}
	}

	engine := chat.NewEngine(chatURL, m.settings.Model, m.settings.Thinking)

	ctx, cancel := context.WithCancel(context.Background())
	m.cancelFn = cancel

	ch := engine.Stream(ctx, apiMsgs)
	m.streamCh = ch

	m.updateViewportContent()
	return m, waitForStreamEvent(ch)
}

func (m Model) handleStreamEvent(event chat.StreamEvent) (tea.Model, tea.Cmd) {
	if event.Error != nil {
		m.lastError = event.Error.Error()
		m.streaming = false
		m.mode = ModeChat
		m.textarea.Focus()
		m.updateViewportContent()
		return m, nil
	}

	if event.Done {
		assistantMsg := config.DisplayMessage{
			Message: config.Message{
				ID:              fmt.Sprintf("msg_%d", len(m.messages)+1),
				Role:            "assistant",
				CreatedAt:       time.Now(),
				Text:            m.streamText.String(),
				ThinkingText:    m.streamThinking.String(),
				OutputTokens:    m.tokenCount,
				TokensPerSecond: m.calcTokPerSec(),
				StopReason:      event.StopReason,
			},
		}
		m.messages = append(m.messages, assistantMsg)
		m.session.Messages = m.extractSessionMessages()
		m.totalTokens += m.tokenCount
		m.streaming = false
		m.mode = ModeChat
		m.textarea.Focus()
		m.updateViewportContent()
		return m, nil
	}

	if event.Text != "" {
		m.streamText.WriteString(event.Text)
		m.tokenCount += countTokensApprox(event.Text)
		if m.firstTokenTime.IsZero() {
			m.firstTokenTime = time.Now()
		}
	}
	if event.Thinking != "" {
		m.streamThinking.WriteString(event.Thinking)
		m.tokenCount += countTokensApprox(event.Thinking)
		if m.firstTokenTime.IsZero() {
			m.firstTokenTime = time.Now()
		}
	}
	m.inThinking = event.InThinking
	m.updateViewportContent()
	return m, waitForStreamEvent(m.streamCh)
}

func (m *Model) updateViewportContent() {
	var b strings.Builder

	for _, msg := range m.messages {
		b.WriteString(ui.RenderMessage(msg, m.width, msg.ThinkingExpanded))
	}

	if m.streaming {
		b.WriteString(ui.RenderStreamingMessage(
			m.streamText.String(),
			m.streamThinking.String(),
			m.inThinking,
			m.width,
		))
	}

	if m.lastError != "" {
		b.WriteString(ui.ErrorStyle.Render("Error: " + m.lastError))
		b.WriteString("\n")
	}

	m.viewport.SetContent(b.String())
	m.viewport.GotoBottom()
}

func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	if m.mode == ModeHelp {
		return m.renderHelp()
	}

	var sections []string

	// Viewport (messages)
	sections = append(sections, m.viewport.View())

	// Command palette overlay (between viewport and input)
	switch m.mode {
	case ModeCommand:
		sections = append(sections, m.cmdPalette.Render(m.width))
	case ModeModelPicker:
		sections = append(sections, m.modelPicker.Render(m.width))
	case ModeSessionPicker:
		sections = append(sections, m.sessionPicker.Render(m.width))
	case ModeFilePicker:
		sections = append(sections, m.filePicker.Render(m.width))
	}

	// Attachment chip
	if m.attachedImage != "" {
		sections = append(sections, ui.AttachmentStyle.Render("  attached: "+m.attachedImage))
	}

	// Textarea
	sections = append(sections, m.textarea.View())

	// Footer
	footerData := ui.FooterData{
		Model:       m.settings.Model,
		TotalTokens: m.totalTokens + m.tokenCount,
		Streaming:   m.streaming,
		InThinking:  m.inThinking,
		TokPerSec:   m.calcTokPerSec(),
	}
	sections = append(sections, ui.RenderFooter(footerData, m.width))

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m *Model) toggleLastThinking() {
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].Role == "assistant" && m.messages[i].ThinkingText != "" {
			m.messages[i].ThinkingExpanded = !m.messages[i].ThinkingExpanded
			break
		}
	}
}

func (m Model) historyUp() (Model, tea.Cmd) {
	if len(m.history.Entries) == 0 {
		return m, nil
	}
	if m.historyIdx == -1 {
		m.draft = m.textarea.Value()
		m.historyIdx = len(m.history.Entries) - 1
	} else if m.historyIdx > 0 {
		m.historyIdx--
	}
	m.textarea.SetValue(m.history.Entries[m.historyIdx])
	return m, nil
}

func (m Model) historyDown() (Model, tea.Cmd) {
	if m.historyIdx == -1 {
		return m, nil
	}
	if m.historyIdx < len(m.history.Entries)-1 {
		m.historyIdx++
		m.textarea.SetValue(m.history.Entries[m.historyIdx])
	} else {
		m.historyIdx = -1
		m.textarea.SetValue(m.draft)
	}
	return m, nil
}

func (m Model) buildAPIMessages() []chat.ChatMessage {
	var msgs []chat.ChatMessage

	sysPrompt := config.LoadSystemPrompt(m.paths, m.settings.SystemPromptFile)
	msgs = append(msgs, chat.ChatMessage{Role: "system", Content: sysPrompt})

	for _, msg := range m.messages {
		if msg.Role == "user" {
			if msg.ImagePath != "" {
				parts, err := chat.BuildMultimodalContent(msg.Text, msg.ImagePath)
				if err == nil {
					msgs = append(msgs, chat.ChatMessage{Role: "user", Content: parts})
				} else {
					msgs = append(msgs, chat.ChatMessage{Role: "user", Content: msg.Text})
				}
			} else {
				msgs = append(msgs, chat.ChatMessage{Role: "user", Content: msg.Text})
			}
		} else if msg.Role == "assistant" {
			msgs = append(msgs, chat.ChatMessage{Role: msg.Role, Content: msg.Text})
		}
	}

	return msgs
}

func (m Model) extractSessionMessages() []config.Message {
	msgs := make([]config.Message, len(m.messages))
	for i, dm := range m.messages {
		msgs[i] = dm.Message
	}
	return msgs
}

func (m Model) calcTokPerSec() float64 {
	if m.firstTokenTime.IsZero() || m.tokenCount == 0 {
		return 0
	}
	elapsed := time.Since(m.firstTokenTime).Seconds()
	if elapsed <= 0 {
		return 0
	}
	return float64(m.tokenCount) / elapsed
}

func countTokensApprox(s string) int {
	n := len(s) / 4
	if n == 0 && len(s) > 0 {
		n = 1
	}
	return n
}

// SetAttachedImage sets the image to attach to the next message (from --image flag)
func (m *Model) SetAttachedImage(path string) {
	m.attachedImage = path
}

// SetInitialPrompt sets the textarea content (from --prompt flag)
func (m *Model) SetInitialPrompt(text string) {
	m.textarea.SetValue(text)
}

func (m Model) renderHelp() string {
	return ui.RenderHelp(m.width, m.height)
}
