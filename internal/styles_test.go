package internal

import (
	"testing"
)

func TestStatusBarStyle_Exists(t *testing.T) {
	// Verify StatusBarStyle is defined and renders without panic
	rendered := StatusBarStyle.Render("test")
	if rendered == "" {
		t.Error("StatusBarStyle should render non-empty output")
	}
}

func TestMessageStyles_Exist(t *testing.T) {
	// Verify all message styles are defined and render without panic
	styles := map[string]string{
		"MessageStyle": MessageStyle.Render("test"),
		"ErrorStyle":   ErrorStyle.Render("test"),
		"SuccessStyle": SuccessStyle.Render("test"),
		"WarningStyle": WarningStyle.Render("test"),
	}

	for name, rendered := range styles {
		if rendered == "" {
			t.Errorf("%s should render non-empty output", name)
		}
	}
}

func TestDiffStyles_Exist(t *testing.T) {
	// Verify all diff styles are defined and render without panic
	styles := map[string]string{
		"DiffAddedStyle":    DiffAddedStyle.Render("|"),
		"DiffDeletedStyle":  DiffDeletedStyle.Render("|"),
		"DiffModifiedStyle": DiffModifiedStyle.Render("|"),
	}

	for name, rendered := range styles {
		if rendered == "" {
			t.Errorf("%s should render non-empty output", name)
		}
	}
}
