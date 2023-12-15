package commands

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

type (
	Init     func(set *flag.FlagSet) any
	Callback func(config any, args []string) (result any, err error)
	Command  struct {
		Name        string
		Aliases     []string
		Description string
		Init        Init
		Callback    Callback
	}
	Commands []Command
)

func (c Commands) Map() (m map[string]Command) {
	m = make(map[string]Command, len(c))
	for _, cmd := range c {
		m[cmd.Name] = cmd
		for _, alias := range cmd.Aliases {
			m[alias] = cmd
		}
	}
	return m
}

func Usage(commands []Command, executable string) (s string) {
	table := struct {
		CommandMax   int
		Commands     []string
		Descriptions []string
	}{}

	for _, command := range commands {
		entry := fmt.Sprintf("%s %s", command.Name, strings.Join(command.Aliases, " "))
		table.CommandMax = max(table.CommandMax, len(entry))
		table.Commands = append(table.Commands, entry)
		table.Descriptions = append(table.Descriptions, command.Description)
	}

	var builder strings.Builder

	fmt.Fprintf(&builder, "Usage: %s SUBCOMMAND [-FLAGS] [ARGUMENTS]\n\nAvailable subcommands:\n", executable)

	for index, entry := range table.Commands {
		description := table.Descriptions[index]
		fmt.Fprintf(&builder, "  %s : %s\n", entry+strings.Repeat(" ", table.CommandMax-len(entry)), description)
	}

	s = builder.String()
	return s
}

var ErrUnknownCommand = errors.New("unknown command")

func Handle(commands Commands, args []string) (result any, err error) {
	m := commands.Map()
	cmd, found := m[args[0]]
	if !found {
		err = fmt.Errorf("%w: %s", ErrUnknownCommand, args[0])
		return
	}

	set := flag.NewFlagSet(cmd.Name, flag.ExitOnError)
	var config any
	if cmd.Init != nil {
		config = cmd.Init(set)
	}
	err = set.Parse(args[1:])
	if err == nil {
		result, err = cmd.Callback(config, set.Args())
	} else if errors.Is(err, flag.ErrHelp) {
		err = nil
	}
	return
}
