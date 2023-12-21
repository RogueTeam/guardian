package commands

import (
	"errors"
	"fmt"
	"strconv"
)

type Type uint8

const (
	TypeString = iota
	TypeBool
	TypeInt
)

type (
	Setup    func(ctx *Context, flags map[string]any) (err error)
	Callback func(ctx *Context, flags map[string]any, args map[string]any) (result any, err error)
	Defer    func(ctx *Context, result any) (finalResult any, err error)
	Value    struct {
		Type        Type
		Name        string
		Description string

		// Reserved for flags
		Default any
	}
	Values  []Value
	Context struct {
		entries map[any]any
	}
	Command struct {
		Name        string
		Description string
		Args        Values
		Flags       Values
		Setup       Setup
		Callback    Callback
		Defer       Defer
		SubCommands Commands
	}
	Commands []*Command
)

func NewContext() *Context {
	return &Context{
		entries: make(map[any]any),
	}
}

func (ctx *Context) Set(key any, value any) {
	ctx.entries[key] = value
}

func (ctx *Context) Get(key any) (value any, found bool) {
	value, found = ctx.entries[key]
	return
}

func (ctx *Context) MustGet(key any) (value any) {
	value, _ = ctx.Get(key)
	return
}

var (
	ErrUnknownType            = errors.New("unknown type")
	ErrInvalidNumberOfArgs    = errors.New("invalid number of arguments")
	ErrArgsNotAllowedInParent = errors.New("arguments not allowed in parent")
	ErrFlagNotFound           = errors.New("flag not found")
	ErrIncompleteFlag         = errors.New("incomplete flag")
)

func (c *Command) Run(args []string) (result any, err error) {
	ctx := NewContext()

	curr := c.tree()

	ctxArgs := make(map[string]any, len(args))
	ctxFlags := make(map[string]any, len(args))
	defers := make([]Defer, 0, len(args))

	// Initialize defaults
	for _, flag := range curr.Flags {
		if flag.Default != nil {
			ctxFlags[flag.Name] = flag.Default
		}
	}

	for index := 0; index < len(args); {

		arg := args[index]
		index++

		switch arg[0] {
		case '-': // Is flag
			flag := arg[1:]
			fDef, found := curr.Flags[flag]
			if !found {
				err = fmt.Errorf("%w: -%s", ErrFlagNotFound, flag)
				return
			}
			switch fDef.Type {
			case TypeString:
				if index >= len(args) {
					err = fmt.Errorf("%w: expecting value for -%s", ErrIncompleteFlag, flag)
					return
				}
				ctxFlags[flag] = args[index]
				index++
			case TypeBool:
				ctxFlags[flag] = true
			case TypeInt:
				if index >= len(args) {
					err = fmt.Errorf("%w: expecting value for -%s", ErrIncompleteFlag, flag)
					return
				}
				var i int
				i, err = strconv.Atoi(args[index])
				index++
				if err != nil {
					err = fmt.Errorf("failed to parse integer for flag -%s: %w", flag, err)
					return
				}
				ctxFlags[flag] = i
			default:
				err = fmt.Errorf("%w: %s", ErrUnknownType, fDef.Name)
				return
			}
		default: // Can be a command or subcommand
			// Check if it is a command
			sub, found := curr.SubCommands[arg]
			if found { // Is subcommand

				// If it is help message return inmediatly
				if sub.Name == HelpCommand && sub.Callback != nil {
					result, err = sub.Callback(ctx, ctxFlags, ctxArgs)
					return
				}
				// Setup should not allow args
				if len(ctxArgs) > 0 {
					err = fmt.Errorf("%w: %s", ErrArgsNotAllowedInParent, curr.Name)
					return
				}

				// Setup parent
				if curr.Setup != nil {
					err = curr.Setup(ctx, ctxFlags)
					if err != nil {
						err = fmt.Errorf("failed to setup parent command %s: %w", curr.Name, err)
						return
					}
				}

				// Prepare defer
				if curr.Defer != nil {
					defers = append(defers, curr.Defer)
				}

				// Clear ctx
				clear(ctxArgs)
				clear(ctxFlags)

				// Make new current
				curr = sub

				// Initialize defaults
				for _, flag := range curr.Flags {
					if flag.Default != nil {
						ctxFlags[flag.Name] = flag.Default
					}
				}
			} else { // Is argument
				argIdx := len(ctxArgs)
				if argIdx >= len(curr.Args) {
					err = fmt.Errorf("%w: %s: expecting %d: %s", ErrInvalidNumberOfArgs, curr.Name, len(curr.Args), arg)
					return
				}
				argEntry := curr.Args[argIdx]
				switch argEntry.Type {
				case TypeString:
					ctxArgs[argEntry.Name] = arg
				case TypeBool:
					ctxArgs[argEntry.Name] = arg == "true"
				case TypeInt:
					var i int
					i, err = strconv.Atoi(arg)
					if err != nil {
						err = fmt.Errorf("failed to parse integer for argument %s: %w", argEntry.Name, err)
						return
					}
					ctxArgs[argEntry.Name] = i
				default:
					err = fmt.Errorf("%w: %s", ErrUnknownType, argEntry.Name)
					return
				}
			}
		}
	}

	if curr.Setup != nil {
		err = curr.Setup(ctx, ctxFlags)
		if err != nil {
			err = fmt.Errorf("failed to setup parent command %s: %w", curr.Name, err)
			return
		}
	}

	if curr.Callback != nil {
		result, err = curr.Callback(ctx, ctxFlags, ctxArgs)
		if err != nil {
			err = fmt.Errorf("failure during command %s: %w", curr.Name, err)
			return
		}
	}

	if curr.Defer != nil {
		result, err = curr.Defer(ctx, result)
		if err != nil {
			err = fmt.Errorf("failed during defer of command %s: %w", curr.Name, err)
			return
		}
	}

	for index := len(defers) - 1; index >= 0; index-- {
		result, err = defers[index](ctx, result)
		if err == nil {
			break
		}
	}

	return
}
