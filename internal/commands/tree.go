package commands

type node struct {
	Name        string
	Description string
	Args        []Value
	Flags       map[string]Value
	Setup       Setup
	Callback    Callback
	Defer       Defer
	SubCommands map[string]*node
}

const HelpCommand = "help"

func (c *Command) tree() (n *node) {
	n = new(node)

	if c.Name != HelpCommand {
		c.setHelp()
	}

	n.Name = c.Name
	n.Description = c.Description

	// Arguments
	n.Args = c.Args

	// Flags
	n.Flags = make(map[string]Value, len(c.Flags))
	for _, flag := range c.Flags {
		n.Flags[flag.Name] = flag
	}

	// Setup
	n.Setup = c.Setup

	// Callback
	n.Callback = c.Callback

	// Defer
	n.Defer = c.Defer

	// SubCommands
	n.SubCommands = make(map[string]*node, len(c.SubCommands))
	for _, cmd := range c.SubCommands {
		n.SubCommands[cmd.Name] = cmd.tree()
	}

	return n
}
