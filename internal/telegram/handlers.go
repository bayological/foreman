package telegram

// Handlers manages command handlers for the Telegram bot.
type Handlers struct {
	bot      *Bot
	commands map[string]CommandHandler
}

// NewHandlers creates a new Handlers instance.
func NewHandlers(bot *Bot) *Handlers {
	h := &Handlers{
		bot:      bot,
		commands: make(map[string]CommandHandler),
	}
	h.registerDefaultHandlers()
	return h
}

// registerDefaultHandlers registers the default command handlers.
func (h *Handlers) registerDefaultHandlers() {
	h.Register("start", h.handleStart)
	h.Register("status", h.handleStatus)
	h.Register("tasks", h.handleTasks)
	h.Register("help", h.handleHelp)
}

// Register registers a new command handler.
func (h *Handlers) Register(command string, handler CommandHandler) {
	h.commands[command] = handler
}

// Handle handles an incoming command.
func (h *Handlers) Handle(command, args string) {
	handler, ok := h.commands[command]
	if !ok {
		h.bot.Send("Unknown command: " + command)
		return
	}
	handler(args)
}

func (h *Handlers) handleStart(args string) {
	h.bot.Send("Foreman bot started. Use /help for available commands.")
}

func (h *Handlers) handleStatus(args string) {
	// TODO: Implement status command
	h.bot.Send("Status: Running")
}

func (h *Handlers) handleTasks(args string) {
	// TODO: Implement tasks command
	h.bot.Send("No active tasks")
}

func (h *Handlers) handleHelp(args string) {
	help := `Available commands:
/start - Start the bot
/status - Show current status
/tasks - List active tasks
/help - Show this help message`
	h.bot.Send(help)
}
