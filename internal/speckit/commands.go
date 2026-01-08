package speckit

// Command represents a SpecKit slash command
type Command string

const (
	CmdConstitution Command = "speckit.constitution"
	CmdSpecify      Command = "speckit.specify"
	CmdClarify      Command = "speckit.clarify"
	CmdPlan         Command = "speckit.plan"
	CmdTasks        Command = "speckit.tasks"
)

type CommandInfo struct {
	Name        Command
	Description string
	NeedsArgs   bool
	ArgsHint    string
}

var Commands = map[Command]CommandInfo{
	CmdConstitution: {
		Name:        CmdConstitution,
		Description: "Create or update project governing principles",
		NeedsArgs:   true,
		ArgsHint:    "Describe principles for code quality, testing, UX, performance",
	},
	CmdSpecify: {
		Name:        CmdSpecify,
		Description: "Define what you want to build",
		NeedsArgs:   true,
		ArgsHint:    "Describe the feature you want to build",
	},
	CmdClarify: {
		Name:        CmdClarify,
		Description: "Clarify underspecified areas with Q&A",
		NeedsArgs:   false,
		ArgsHint:    "",
	},
	CmdPlan: {
		Name:        CmdPlan,
		Description: "Create technical implementation plan",
		NeedsArgs:   true,
		ArgsHint:    "Describe your tech stack and architecture choices",
	},
	CmdTasks: {
		Name:        CmdTasks,
		Description: "Generate actionable task breakdown",
		NeedsArgs:   false,
		ArgsHint:    "",
	},
}
