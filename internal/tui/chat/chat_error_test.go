package chat

import (
	"errors"
	"strings"
	"testing"

	"github.com/GarroshIcecream/yummy/internal/config"
	"github.com/GarroshIcecream/yummy/internal/db"
	"github.com/GarroshIcecream/yummy/internal/llms"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/charmbracelet/x/ansi"
)

func newTestChat(t *testing.T) *ChatModel {
	t.Helper()
	config.SetGlobalConfig(config.NewDefaultConfig())
	sessionLog, err := db.NewSessionLog(t.TempDir(), &config.NewDefaultConfig().Database)
	if err != nil {
		t.Fatal(err)
	}
	ai, err := llms.NewLLMServiceWithProvider(nil, stubProvider{})
	if err != nil {
		t.Fatal(err)
	}
	e, err := NewExecutorService(ai, nil, sessionLog)
	if err != nil {
		t.Fatal(err)
	}
	theme := themes.NewDefaultTheme()
	m, err := NewChatModel(e, &theme)
	if err != nil {
		t.Fatal(err)
	}
	m.SetSize(100, 30)
	return m
}

func TestChatShowsGenerationError(t *testing.T) {
	m := newTestChat(t)
	m.Update(messages.ResponseMsg{GenerationID: m.generationID, Err: errors.New("ChatGPT is not available: not signed in")})
	if err := m.RenderConversationAsMarkdown(); err != nil {
		t.Fatal(err)
	}
	if got := ansi.Strip(m.viewport.View()); !strings.Contains(got, "not signed in") {
		t.Fatalf("viewport does not show the error:\n%s", got)
	}
}
