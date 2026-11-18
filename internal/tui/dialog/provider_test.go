package dialog

import (
	"context"
	"errors"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/GarroshIcecream/yummy/internal/llms"
	messages "github.com/GarroshIcecream/yummy/internal/models/msg"
	themes "github.com/GarroshIcecream/yummy/internal/themes"
	"github.com/charmbracelet/x/ansi"
)

// runCmd executes cmd and returns every message it produces, expanding
// batches. Commands that block on ctx are given a short deadline by the
// fake deps, so this terminates.
func runCmd(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		var out []tea.Msg
		for _, c := range batch {
			out = append(out, runCmd(c)...)
		}
		return out
	}
	if msg == nil {
		return nil
	}
	return []tea.Msg{msg}
}

func findMsg[T any](msgs []tea.Msg) (T, bool) {
	for _, m := range msgs {
		if v, ok := m.(T); ok {
			return v, true
		}
	}
	var zero T
	return zero, false
}

type fakeDeps struct {
	signedIn    bool
	loginErr    error
	opened      []string
	prepareErr  error
	preparedFor []string
}

func (f *fakeDeps) deps() providerDeps {
	return providerDeps{
		signedIn: func() bool { return f.signedIn },
		login: func(_ context.Context, openBrowser func(string) error) (string, error) {
			_ = openBrowser("https://auth.example/authorize?x=1")
			if f.loginErr != nil {
				return "", f.loginErr
			}
			f.signedIn = true
			return "acc_1", nil
		},
		browser: func(u string) error { f.opened = append(f.opened, u); return nil },
		prepare: func(name string) llms.ProviderState {
			f.preparedFor = append(f.preparedFor, name)
			if f.prepareErr != nil {
				return llms.ProviderState{Err: f.prepareErr}
			}
			return llms.ProviderState{ModelName: "model-1"}
		},
	}
}

func newTestProviderDialog(current string, f *fakeDeps) *ProviderDialogCmp {
	theme := themes.NewDefaultTheme()
	d := newProviderDialog(current, &theme, f.deps())
	d.SetSize(100, 40)
	return d
}

func press(d *ProviderDialogCmp, key string) []tea.Msg {
	var k tea.KeyPressMsg
	switch key {
	case "enter":
		k = tea.KeyPressMsg{Code: tea.KeyEnter}
	case "down":
		k = tea.KeyPressMsg{Code: tea.KeyDown}
	case "esc":
		k = tea.KeyPressMsg{Code: tea.KeyEscape}
	}
	_, cmd := d.Update(k)
	return runCmd(cmd)
}

// feed delivers msgs to the dialog (as the TUI would) and returns what the
// resulting commands produce.
func feed(d *ProviderDialogCmp, msgs []tea.Msg) []tea.Msg {
	var out []tea.Msg
	for _, m := range msgs {
		_, cmd := d.Update(m)
		out = append(out, runCmd(cmd)...)
	}
	return out
}

func TestProviderDialogSwitchToOllama(t *testing.T) {
	f := &fakeDeps{}
	d := newTestProviderDialog("chatgpt", f)
	if d.selected != 1 {
		t.Fatalf("cursor = %d, want it on the current provider (chatgpt)", d.selected)
	}
	d.selected = 0
	msgs := feed(d, press(d, "enter"))
	applied, ok := findMsg[ProviderAppliedMsg](msgs)
	if !ok || applied.Name != "ollama" || applied.State.ModelName != "model-1" {
		t.Fatalf("msgs = %#v, want ProviderAppliedMsg for ollama", msgs)
	}
	if !strings.Contains(ansi.Strip(d.View().Content), "Now using Ollama") {
		t.Errorf("view does not confirm the switch:\n%s", ansi.Strip(d.View().Content))
	}
}

func TestProviderDialogChatGPTNeedsSignIn(t *testing.T) {
	f := &fakeDeps{}
	d := newTestProviderDialog("ollama", f)
	if !strings.Contains(ansi.Strip(d.View().Content), "not signed in") {
		t.Errorf("chooser should show ChatGPT as not signed in:\n%s", ansi.Strip(d.View().Content))
	}
	press(d, "down")
	// enter → sign-in → (login result) → switch → (prepared) → applied
	msgs := feed(d, press(d, "enter"))
	msgs = append(msgs, feed(d, msgs)...)
	if len(f.opened) != 1 {
		t.Errorf("browser opened %d times, want 1", len(f.opened))
	}
	applied, ok := findMsg[ProviderAppliedMsg](msgs)
	if !ok || applied.Name != "chatgpt" {
		t.Fatalf("msgs = %#v, want ProviderAppliedMsg for chatgpt", msgs)
	}
}

func TestProviderDialogShowsAuthURLWhileWaiting(t *testing.T) {
	theme := themes.NewDefaultTheme()
	d := newProviderDialog("ollama", &theme, (&fakeDeps{}).deps())
	d.SetSize(100, 40)
	d.phase = providerPhaseSigningIn
	d.Update(providerLoginURLMsg{url: "https://auth.example/authorize?x=1"})
	if !strings.Contains(ansi.Strip(d.View().Content), "auth.example") {
		t.Errorf("view does not show the sign-in URL:\n%s", ansi.Strip(d.View().Content))
	}
}

func TestProviderDialogSignInFailureShowsError(t *testing.T) {
	f := &fakeDeps{loginErr: errors.New("access_denied")}
	d := newTestProviderDialog("ollama", f)
	d.selected = 1
	feed(d, press(d, "enter"))
	if d.phase != providerPhaseFailed || !strings.Contains(d.errMsg, "access_denied") {
		t.Fatalf("phase=%v err=%q, want failed with access_denied", d.phase, d.errMsg)
	}
	if len(f.preparedFor) != 0 {
		t.Errorf("switched despite failed sign-in: %v", f.preparedFor)
	}
}

func TestProviderDialogSwitchFailureDoesNotApply(t *testing.T) {
	f := &fakeDeps{signedIn: true, prepareErr: errors.New("catalog down")}
	d := newTestProviderDialog("ollama", f)
	d.selected = 1
	msgs := feed(d, press(d, "enter"))
	if _, ok := findMsg[ProviderAppliedMsg](msgs); ok {
		t.Fatal("ProviderAppliedMsg sent for a provider that failed to prepare")
	}
	if !strings.Contains(ansi.Strip(d.View().Content), "catalog down") {
		t.Errorf("view does not show the failure:\n%s", ansi.Strip(d.View().Content))
	}
}

func TestProviderDialogEscCancelsSignIn(t *testing.T) {
	f := &fakeDeps{}
	d := newTestProviderDialog("ollama", f)
	d.selected = 1
	_, _ = d.Update(tea.KeyPressMsg{Code: tea.KeyEnter}) // starts sign-in, don't run it
	msgs := press(d, "esc")
	if _, ok := findMsg[messages.CloseModalViewMsg](msgs); !ok {
		t.Fatalf("esc did not close the dialog: %#v", msgs)
	}
	if d.cancel == nil {
		t.Fatal("sign-in had no cancel func")
	}
}
