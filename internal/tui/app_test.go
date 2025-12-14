package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/altmueller/xstream-tui/internal/xc"
)

func TestNewApp(t *testing.T) {
	app := NewApp()
	if app == nil {
		t.Fatal("NewApp returned nil")
	}
	if app.screen != LoginScreen {
		t.Errorf("expected LoginScreen, got %v", app.screen)
	}
	if app.player == nil {
		t.Error("player manager not initialized")
	}
}

func TestAppUpdate_WindowSize(t *testing.T) {
	app := NewApp()
	msg := tea.WindowSizeMsg{Width: 80, Height: 24}

	model, _ := app.Update(msg)
	updatedApp := model.(*App)

	if updatedApp.width != 80 {
		t.Errorf("expected width 80, got %d", updatedApp.width)
	}
	if updatedApp.height != 24 {
		t.Errorf("expected height 24, got %d", updatedApp.height)
	}
}

func TestAppUpdate_AuthSuccess(t *testing.T) {
	app := NewApp()
	client, _ := xc.NewClient("http://test.com", "user", "pass")
	msg := AuthSuccessMsg{
		Client:   client,
		UserInfo: xc.UserInfo{Username: "testuser"},
	}

	model, _ := app.Update(msg)
	updatedApp := model.(*App)

	if updatedApp.screen != ContentTypeScreen {
		t.Errorf("expected ContentTypeScreen, got %v", updatedApp.screen)
	}
	if updatedApp.client != client {
		t.Error("client not set")
	}
	if updatedApp.loading {
		t.Error("loading should be false after auth success")
	}
}

func TestAppUpdate_ContentTypeSelected(t *testing.T) {
	app := NewApp()
	app.screen = ContentTypeScreen
	msg := ContentTypeSelectedMsg{Type: LiveContent}

	model, _ := app.Update(msg)
	updatedApp := model.(*App)

	if updatedApp.currentType != LiveContent {
		t.Errorf("expected LiveContent, got %v", updatedApp.currentType)
	}
	if updatedApp.screen != CategoriesScreen {
		t.Errorf("expected CategoriesScreen, got %v", updatedApp.screen)
	}
	if !updatedApp.loading {
		t.Error("loading should be true after content type selection")
	}
}

func TestAppUpdate_CategorySelected(t *testing.T) {
	app := NewApp()
	app.screen = CategoriesScreen
	cat := xc.Category{Name: "Sports"}
	msg := CategorySelectedMsg{Category: cat}

	model, _ := app.Update(msg)
	updatedApp := model.(*App)

	if updatedApp.currentCat.Name != "Sports" {
		t.Errorf("expected Sports category, got %v", updatedApp.currentCat.Name)
	}
	if updatedApp.screen != StreamsScreen {
		t.Errorf("expected StreamsScreen, got %v", updatedApp.screen)
	}
}

func TestAppUpdate_ErrorMsg(t *testing.T) {
	app := NewApp()
	app.loading = true
	msg := ErrorMsg{Err: xc.ErrInvalidConfig}

	model, _ := app.Update(msg)
	updatedApp := model.(*App)

	if updatedApp.errorMsg == "" {
		t.Error("error message not set")
	}
	if updatedApp.loading {
		t.Error("loading should be false after error")
	}
}

func TestAppNavigateBack(t *testing.T) {
	app := NewApp()
	// Simulate navigation: Login -> ContentType -> Categories
	app.navStack = []Screen{LoginScreen, ContentTypeScreen}
	app.screen = CategoriesScreen

	updatedApp, _ := app.navigateBack()

	if updatedApp.screen != ContentTypeScreen {
		t.Errorf("expected ContentTypeScreen, got %v", updatedApp.screen)
	}
	if len(updatedApp.navStack) != 1 {
		t.Errorf("expected 1 item in nav stack, got %d", len(updatedApp.navStack))
	}
}

func TestAppNavigateBack_EmptyStack(t *testing.T) {
	app := NewApp()
	app.navStack = []Screen{}
	app.screen = LoginScreen

	updatedApp, _ := app.navigateBack()

	if updatedApp.screen != LoginScreen {
		t.Errorf("expected to stay on LoginScreen, got %v", updatedApp.screen)
	}
}

func TestContentTypeString(t *testing.T) {
	tests := []struct {
		ct       ContentType
		expected string
	}{
		{LiveContent, "Live TV"},
		{VODContent, "Movies"},
		{SeriesContent, "Series"},
	}

	for _, tt := range tests {
		if got := tt.ct.String(); got != tt.expected {
			t.Errorf("ContentType %d: expected %q, got %q", tt.ct, tt.expected, got)
		}
	}
}

func TestPlayerStartedMsg(t *testing.T) {
	app := NewApp()
	app.loading = true
	msg := PlayerStartedMsg{PlayerType: "mpv"}

	model, _ := app.Update(msg)
	updatedApp := model.(*App)

	if updatedApp.loading {
		t.Error("loading should be false after player started")
	}
	if updatedApp.errorMsg == "" {
		t.Error("expected status message about playing")
	}
}

func TestPlayerStoppedMsg(t *testing.T) {
	app := NewApp()
	app.errorMsg = "Playing..."
	msg := PlayerStoppedMsg{Err: nil}

	model, _ := app.Update(msg)
	updatedApp := model.(*App)

	if updatedApp.errorMsg != "" {
		t.Errorf("expected empty error message on clean exit, got %q", updatedApp.errorMsg)
	}
}
