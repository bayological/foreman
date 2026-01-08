package telegram

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		max      int
		expected string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "he..."},
		{"", 5, ""},
		{"abc", 3, "abc"},
		{"abcd", 3, "..."},
		{"hello world", 11, "hello world"},
		{"hello world!", 11, "hello wo..."},
	}

	for _, tc := range tests {
		result := truncate(tc.input, tc.max)
		if result != tc.expected {
			t.Errorf("truncate(%q, %d) = %q, expected %q", tc.input, tc.max, result, tc.expected)
		}
	}
}

func TestBotStructInitialization(t *testing.T) {
	// Test that Bot struct has expected fields
	bot := &Bot{
		chatID:    12345,
		commands:  make(map[string]CommandHandler),
		callbacks: make(map[string]CallbackHandler),
	}

	if bot.chatID != 12345 {
		t.Errorf("Expected chatID 12345, got %d", bot.chatID)
	}
	if bot.commands == nil {
		t.Error("Expected commands map to be initialized")
	}
	if bot.callbacks == nil {
		t.Error("Expected callbacks map to be initialized")
	}
}

func TestRegisterCommand(t *testing.T) {
	bot := &Bot{
		commands: make(map[string]CommandHandler),
	}

	called := false
	handler := func(args string) {
		called = true
	}

	bot.RegisterCommand("test", handler)

	if _, ok := bot.commands["test"]; !ok {
		t.Error("Expected command 'test' to be registered")
	}

	// Call the handler to verify it was registered correctly
	bot.commands["test"]("")
	if !called {
		t.Error("Expected handler to be called")
	}
}

func TestRegisterCallback(t *testing.T) {
	bot := &Bot{
		callbacks: make(map[string]CallbackHandler),
	}

	called := false
	handler := func(data string) {
		called = true
	}

	bot.RegisterCallback("approve", handler)

	if _, ok := bot.callbacks["approve"]; !ok {
		t.Error("Expected callback 'approve' to be registered")
	}

	// Call the handler to verify it was registered correctly
	bot.callbacks["approve"]("approve:task-1")
	if !called {
		t.Error("Expected handler to be called")
	}
}

func TestRegisterMessageHandler(t *testing.T) {
	bot := &Bot{}

	called := false
	handler := func(text string) {
		called = true
	}

	bot.RegisterMessageHandler(handler)

	if bot.messageHandler == nil {
		t.Error("Expected message handler to be registered")
	}

	// Call the handler to verify it was registered correctly
	bot.messageHandler("test message")
	if !called {
		t.Error("Expected handler to be called")
	}
}

func TestIsAuthorized(t *testing.T) {
	bot := &Bot{
		chatID: 12345,
	}

	tests := []struct {
		chatID     int64
		authorized bool
	}{
		{12345, true},
		{99999, false},
		{0, false},
		{-1, false},
	}

	for _, tc := range tests {
		result := bot.isAuthorized(tc.chatID)
		if result != tc.authorized {
			t.Errorf("isAuthorized(%d) = %v, expected %v", tc.chatID, result, tc.authorized)
		}
	}
}

// Note: Testing Send, RequestApproval, Escalate, RequestPhaseApproval
// would require mocking the Telegram API, which is beyond the scope of
// unit tests. These functions are tested manually or via integration tests.
//
// The bot API (tgbotapi.BotAPI) would need to be abstracted behind an
// interface to enable proper unit testing with mocks.
