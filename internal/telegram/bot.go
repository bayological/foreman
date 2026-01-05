package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api       *tgbotapi.BotAPI
	chatID    int64
	commands  map[string]CommandHandler
	callbacks map[string]CallbackHandler
}

type CommandHandler func(args string)
type CallbackHandler func(data string)

func NewBot(token string, chatID int64) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	return &Bot{
		api:       api,
		chatID:    chatID,
		commands:  make(map[string]CommandHandler),
		callbacks: make(map[string]CallbackHandler),
	}, nil
}

func (b *Bot) RegisterCommand(name string, handler CommandHandler) {
	b.commands[name] = handler
}

func (b *Bot) RegisterCallback(prefix string, handler CallbackHandler) {
	b.callbacks[prefix] = handler
}

func (b *Bot) Listen(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return
		case update := <-updates:
			b.handleUpdate(update)
		}
	}
}

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	// Verify the message comes from the authorized chat
	var chatID int64
	if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
	} else if update.Message != nil {
		chatID = update.Message.Chat.ID
	}

	if !b.isAuthorized(chatID) {
		log.Printf("Unauthorized access attempt from chat ID: %d", chatID)
		return
	}

	// Handle callback queries (button presses)
	if update.CallbackQuery != nil {
		data := update.CallbackQuery.Data

		// Find matching callback handler
		for prefix, handler := range b.callbacks {
			if strings.HasPrefix(data, prefix+":") || data == prefix {
				handler(data)

				// Acknowledge callback
				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
				b.api.Request(callback)
				return
			}
		}
	}

	// Handle commands
	if update.Message != nil && update.Message.IsCommand() {
		cmd := update.Message.Command()
		args := update.Message.CommandArguments()

		if handler, ok := b.commands[cmd]; ok {
			handler(args)
		} else {
			b.Send(fmt.Sprintf("Unknown command: /%s\nUse /help to see available commands", cmd))
		}
	}
}

// isAuthorized checks if the chat ID matches the configured authorized chat.
func (b *Bot) isAuthorized(chatID int64) bool {
	return chatID == b.chatID
}

func (b *Bot) Send(message string) error {
	msg := tgbotapi.NewMessage(b.chatID, message)
	msg.ParseMode = "Markdown"
	_, err := b.api.Send(msg)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
	return err
}

func (b *Bot) RequestApproval(taskID, summary, prURL string) error {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Approve", fmt.Sprintf("approve:%s", taskID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Reject", fmt.Sprintf("reject:%s", taskID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Request Changes", fmt.Sprintf("changes:%s", taskID)),
		),
	)

	text := fmt.Sprintf(
		"🚦 *Approval Required*\n\nTask: `%s`\n\n%s\n\n[View Changes](%s)",
		taskID,
		truncate(summary, 500),
		prURL,
	)

	msg := tgbotapi.NewMessage(b.chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	_, err := b.api.Send(msg)
	return err
}

func (b *Bot) Escalate(taskID, reason, details string) error {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 Retry", fmt.Sprintf("retry:%s", taskID)),
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Abandon", fmt.Sprintf("reject:%s", taskID)),
		),
	)

	text := fmt.Sprintf(
		"🚨 *Escalation Required*\n\nTask: `%s`\nReason: %s\n\n%s",
		taskID,
		reason,
		truncate(details, 500),
	)

	msg := tgbotapi.NewMessage(b.chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	_, err := b.api.Send(msg)
	return err
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}