package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func testPayload() statePayload {
	return statePayload{
		ConfigPath: "/tmp/cli-timer-test-config.json",
		Config: config{
			Font:              "Standard",
			CenterDisplay:     true,
			ShowHeader:        true,
			ShowControls:      true,
			TickRateMs:        100,
			CompletionMessage: "Time is up!",
			Keybindings:       defaultKeybindings,
		},
		Fonts: []string{"Standard", "Big"},
	}
}

func TestEnterSelectsMainMenuAction(t *testing.T) {
	m := newModel(testPayload())
	if m.screen != screenMain {
		t.Fatalf("expected screenMain at init, got %v", m.screen)
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next := updated.(model)

	if next.screen != screenFontPicker {
		t.Fatalf("expected enter to open font picker, got screen %v", next.screen)
	}
}

func TestEnterSelectsFontInPicker(t *testing.T) {
	m := newModel(testPayload())
	m.screen = screenFontPicker
	m.fontList.Select(1)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	next := updated.(model)

	if next.screen != screenMain {
		t.Fatalf("expected return to main screen, got %v", next.screen)
	}
	if next.payload.Config.Font != "Big" {
		t.Fatalf("expected selected font Big, got %q", next.payload.Config.Font)
	}
}

func TestCtrlMAlsoConfirmsSelection(t *testing.T) {
	m := newModel(testPayload())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlM})
	next := updated.(model)

	if next.screen != screenFontPicker {
		t.Fatalf("expected ctrl+m to open font picker, got %v", next.screen)
	}
}

func TestRuneCarriageReturnConfirmsSelection(t *testing.T) {
	m := newModel(testPayload())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\r'}})
	next := updated.(model)

	if next.screen != screenFontPicker {
		t.Fatalf("expected \\r rune to open font picker, got %v", next.screen)
	}
}

func TestRuneLineFeedConfirmsSelection(t *testing.T) {
	m := newModel(testPayload())

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\n'}})
	next := updated.(model)

	if next.screen != screenFontPicker {
		t.Fatalf("expected \\n rune to open font picker, got %v", next.screen)
	}
}
