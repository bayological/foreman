package telegram

import "fmt"

// Notifier handles sending notifications via Telegram.
type Notifier struct {
	bot *Bot
}

// NewNotifier creates a new Notifier.
func NewNotifier(bot *Bot) *Notifier {
	return &Notifier{bot: bot}
}

// TaskStarted notifies that a task has started.
func (n *Notifier) TaskStarted(taskID, description string) error {
	msg := fmt.Sprintf("🚀 Task started: %s\n%s", taskID, description)
	return n.bot.Send(msg)
}

// TaskCompleted notifies that a task has completed.
func (n *Notifier) TaskCompleted(taskID, result string) error {
	msg := fmt.Sprintf("✅ Task completed: %s\n%s", taskID, result)
	return n.bot.Send(msg)
}

// TaskFailed notifies that a task has failed.
func (n *Notifier) TaskFailed(taskID string, err error) error {
	msg := fmt.Sprintf("❌ Task failed: %s\nError: %v", taskID, err)
	return n.bot.Send(msg)
}

// ReviewReady notifies that a review is ready.
func (n *Notifier) ReviewReady(taskID, prURL string) error {
	msg := fmt.Sprintf("📝 Review ready: %s\nPR: %s", taskID, prURL)
	return n.bot.Send(msg)
}

// Error notifies of a general error.
func (n *Notifier) Error(message string, err error) error {
	msg := fmt.Sprintf("⚠️ Error: %s\n%v", message, err)
	return n.bot.Send(msg)
}
