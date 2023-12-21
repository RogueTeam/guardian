package commands

import (
	"bytes"
	"fmt"
	"strings"
)

func (vs Values) Table(prefix string) (s string) {
	var longest int
	for _, v := range vs {
		longest = max(longest, len(v.Name))
	}

	var buf bytes.Buffer
	for _, v := range vs {
		blank := strings.Repeat(" ", longest-len(v.Name))
		fmt.Fprintf(&buf, "%s%s%s : %s\n", prefix, v.Name, blank, v.Description)
	}

	s = buf.String()
	return s
}

func (cs Commands) Table(prefix string) (s string) {
	var longest int
	for _, c := range cs {
		longest = max(longest, len(c.Name))
	}

	var buf bytes.Buffer
	for _, c := range cs {
		blank := strings.Repeat(" ", longest-len(c.Name))
		fmt.Fprintf(&buf, "%s%s%s : %s\n", prefix, c.Name, blank, c.Description)
	}

	s = buf.String()
	return s
}

func (c *Command) setHelp() {
	help := Command{
		Name:        HelpCommand,
		Description: "Prints this help message",
		Callback: func(ctx *Context, flags, args map[string]any) (result any, err error) {
			var buf bytes.Buffer

			fmt.Fprintf(&buf, "Usage: %s [-FLAGS] [SUBCOMMAND] [-FLAGS] [ARGS]\n\n", c.Name)
			fmt.Fprintf(&buf, "Description: %s\n\n", c.Description)
			fmt.Fprintf(&buf, "Flags:\n\n%s\n\n", c.Flags.Table("\t-"))
			fmt.Fprintf(&buf, "Arguments:\n\n%s\n\n", c.Args.Table("\t"))
			fmt.Fprintf(&buf, "Subcommands:\n\n%s\n\n", c.SubCommands.Table("\t"))

			result = buf.String()
			return
		},
	}
	c.SubCommands = append(c.SubCommands, help)
}
