package commands_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/RogueTeam/guardian/internal/commands"
)

func TestCommand_Run(t *testing.T) {
	t.Parallel()

	t.Run("Succeed", func(t *testing.T) {
		t.Parallel()

		type Test struct {
			Name   string
			Root   commands.Command
			Args   []string
			Expect any
		}

		tests := []Test{
			{
				Name: "Help message",
				Root: commands.Command{
					Flags: commands.Values{{commands.TypeString, "string", "description", "this is a default value"}},
					Args:  commands.Values{{commands.TypeString, "string", "description", nil}},
				},
				Args:   []string{"help"},
				Expect: "Usage:",
			},
			{
				Name: "Simple command with flags in root and arguments in subcommand",
				Root: commands.Command{
					Flags: commands.Values{{commands.TypeString, "file", "", nil}},
					Setup: func(ctx *commands.Context, flags map[string]any) (err error) {
						ctx.Set("file", flags["file"])

						return err
					},
					SubCommands: commands.Commands{
						{
							Name:  "init",
							Flags: commands.Values{{commands.TypeBool, "with-db", "", nil}},
							Args:  commands.Values{{commands.TypeString, "user", "", nil}},
							Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
								result = fmt.Sprintf("%s && %v && %s", ctx.MustGet("file"), flags["with-db"], args["user"])

								return
							},
						},
					},
				},
				Args:   []string{"-file", "path/to/file", "init", "-with-db", "root"},
				Expect: "path/to/file && true && root",
			},
			{
				Name: "Direct call to subcommand",
				Root: commands.Command{
					SubCommands: commands.Commands{
						{
							Name:  "init",
							Flags: commands.Values{{commands.TypeString, "string", "description", "this is a default value"}},
							Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
								result = "init"
								return
							},
						},
					},
				},
				Args:   []string{"init"},
				Expect: "init",
			},
			{
				Name: "Subcommand default flags",
				Root: commands.Command{
					SubCommands: commands.Commands{
						{
							Name: "init",
							Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
								result = "init"
								return
							},
						},
					},
				},
				Args:   []string{"init"},
				Expect: "init",
			},
			{
				Name: "All Flags types",
				Root: commands.Command{
					Flags: commands.Values{
						{commands.TypeString, "string", "", nil},
						{commands.TypeBool, "bool", "", nil},
						{commands.TypeInt, "int", "", nil},
					},
					Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
						result = fmt.Sprintf("%s && %v && %d", flags["string"], flags["bool"], flags["int"])

						return
					},
				},
				Args:   []string{"-string", "value", "-bool", "-int", "10"},
				Expect: "value && true && 10",
			},
			{
				Name: "All Args types",
				Root: commands.Command{
					Args: commands.Values{
						{commands.TypeString, "string", "", nil},
						{commands.TypeBool, "bool", "", nil},
						{commands.TypeInt, "int", "", nil},
					},
					Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
						result = fmt.Sprintf("%s && %v && %d", args["string"], args["bool"], args["int"])

						return
					},
				},
				Args:   []string{"value", "true", "10"},
				Expect: "value && true && 10",
			},
			{
				Name: "Defer functions",
				Root: commands.Command{
					Flags: commands.Values{{commands.TypeString, "string", "description", "this is a default value"}},
					Args:  commands.Values{{commands.TypeString, "string", "description", "this is a default value"}},
					Defer: func(ctx *commands.Context, result any) (finalResult any, err error) {
						result = result.(string) + " last"
						return result, err
					},
					SubCommands: commands.Commands{
						{
							Name: "init",
							Defer: func(ctx *commands.Context, result any) (finalResult any, err error) {
								result = result.(string) + " second"
								return result, err
							},
							Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
								result = "first"
								return result, err
							},
						},
					},
				},
				Args:   []string{"init"},
				Expect: "first second last",
			},
		}

		for _, test := range tests {
			test := test
			test.Root.Name = test.Name

			t.Run(test.Name, func(t *testing.T) {
				result, err := test.Root.Run(test.Args)
				if err != nil {
					t.Fatalf("expecting no errors: but obtained: %v", err)
				}

				var comparison bool
				switch expect := test.Expect.(type) {
				case string:
					comparison = strings.Contains(result.(string), expect)
				default:
					comparison = test.Expect == result
				}

				if !comparison {
					t.Fatalf("Expecting(%v) is not equal to Obtained(%v)", test.Expect, result)
				}
			})
		}
	})
	t.Run("Fail", func(t *testing.T) {
		t.Parallel()

		type Test struct {
			Name string
			Root commands.Command
			Args []string
		}

		tests := []Test{
			{
				Name: "Unknown flag",
				Args: []string{"-file"},
			},
			{
				Name: "Incomplete String flag",
				Root: commands.Command{
					Flags: commands.Values{{commands.TypeString, "string", "", nil}},
				},
				Args: []string{"-string"},
			},
			{
				Name: "Incomplete Int flag",
				Root: commands.Command{
					Flags: commands.Values{{commands.TypeInt, "int", "", nil}},
				},
				Args: []string{"-int"},
			},
			{
				Name: "Invalid Int flag",
				Root: commands.Command{
					Flags: commands.Values{{commands.TypeInt, "int", "", nil}},
				},
				Args: []string{"-int", "sulcud"},
			},
			{
				Name: "Invalid Type flag",
				Root: commands.Command{
					Flags: commands.Values{{0xff, "invalid", "", nil}},
				},
				Args: []string{"-invalid", "sulcud"},
			},
			{
				Name: "Subcommand with parent args",
				Root: commands.Command{
					Args:  commands.Values{{commands.TypeString, "user", "", nil}},
					Flags: commands.Values{{0xff, "invalid", "", nil}},
					SubCommands: commands.Commands{
						{
							Name: "init",
						},
					},
				},
				Args: []string{"sulcud", "init"},
			},
			{
				Name: "Setup Parent Error",
				Root: commands.Command{
					Setup: func(ctx *commands.Context, flags map[string]any) (err error) {
						return errors.New("error")
					},
				},
				Args: []string{},
			},
			{
				Name: "Setup SubCommand Error",
				Root: commands.Command{
					Setup: func(ctx *commands.Context, flags map[string]any) (err error) {
						return errors.New("error")
					},
					SubCommands: commands.Commands{
						{
							Name: "init",
						},
					},
				},
				Args: []string{"init"},
			},
			{
				Name: "Invalid number of arguments",
				Root: commands.Command{},
				Args: []string{"init"},
			},
			{
				Name: "Invalid Int arg",
				Root: commands.Command{
					Args: commands.Values{{commands.TypeInt, "int", "", nil}},
				},
				Args: []string{"sulcud"},
			},
			{
				Name: "Invalid Arg Type",
				Root: commands.Command{
					Args: commands.Values{{commands.Type(0xff), "int", "", nil}},
				},
				Args: []string{"sulcud"},
			},
			{
				Name: "Callback error",
				Root: commands.Command{
					Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
						err = errors.New("error")
						return
					},
				},
				Args: []string{},
			},
			{
				Name: "Defer parent",
				Root: commands.Command{
					Defer: func(ctx *commands.Context, result any) (finalResult any, err error) {
						err = errors.New("error")
						return
					},
					SubCommands: commands.Commands{
						{
							Name: "init",
							Defer: func(ctx *commands.Context, result any) (finalResult any, err error) {
								return
							},
							Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
								return
							},
						},
					},
				},
				Args: []string{"init"},
			},
			{
				Name: "Defer subcommand",
				Root: commands.Command{
					SubCommands: commands.Commands{
						{
							Name: "init",
							Defer: func(ctx *commands.Context, result any) (finalResult any, err error) {
								err = errors.New("error")
								return
							},
							Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
								return
							},
						},
					},
				},
				Args: []string{"init"},
			},
		}

		for _, test := range tests {
			test := test
			test.Root.Name = test.Name

			t.Run(test.Name, func(t *testing.T) {
				_, err := test.Root.Run(test.Args)
				if err == nil {
					t.Fatalf("expecting error")
				}
			})
		}
	})
}
