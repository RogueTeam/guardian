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
					Args: []commands.Value{{commands.TypeString, "string", "description"}},
				},
				Args:   []string{"help"},
				Expect: "Usage:",
			},
			{
				Name: "Simple command with flags in root and arguments in subcommand",
				Root: commands.Command{
					Flags: []commands.Value{{commands.TypeString, "file", ""}},
					Setup: func(ctx *commands.Context, flags map[string]any) (err error) {
						ctx.Set("file", flags["file"])

						return err
					},
					SubCommands: []commands.Command{
						{
							Name:  "init",
							Flags: []commands.Value{{commands.TypeBool, "with-db", ""}},
							Args:  []commands.Value{{commands.TypeString, "user", ""}},
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
					SubCommands: []commands.Command{
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
					Flags: []commands.Value{
						{commands.TypeString, "string", ""},
						{commands.TypeBool, "bool", ""},
						{commands.TypeInt, "int", ""},
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
					Args: []commands.Value{
						{commands.TypeString, "string", ""},
						{commands.TypeBool, "bool", ""},
						{commands.TypeInt, "int", ""},
					},
					Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
						result = fmt.Sprintf("%s && %v && %d", args["string"], args["bool"], args["int"])

						return
					},
				},
				Args:   []string{"value", "true", "10"},
				Expect: "value && true && 10",
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
					Flags: []commands.Value{{commands.TypeString, "string", ""}},
				},
				Args: []string{"-string"},
			},
			{
				Name: "Incomplete Int flag",
				Root: commands.Command{
					Flags: []commands.Value{{commands.TypeInt, "int", ""}},
				},
				Args: []string{"-int"},
			},
			{
				Name: "Invalid Int flag",
				Root: commands.Command{
					Flags: []commands.Value{{commands.TypeInt, "int", ""}},
				},
				Args: []string{"-int", "sulcud"},
			},
			{
				Name: "Invalid Type flag",
				Root: commands.Command{
					Flags: []commands.Value{{0xff, "invalid", ""}},
				},
				Args: []string{"-invalid", "sulcud"},
			},
			{
				Name: "Subcommand with parent args",
				Root: commands.Command{
					Args:  []commands.Value{{commands.TypeString, "user", ""}},
					Flags: []commands.Value{{0xff, "invalid", ""}},
					SubCommands: []commands.Command{
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
					SubCommands: []commands.Command{
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
					Args: []commands.Value{{commands.TypeInt, "int", ""}},
				},
				Args: []string{"sulcud"},
			},
			{
				Name: "Invalid Arg Type",
				Root: commands.Command{
					Args: []commands.Value{{commands.Type(0xff), "int", ""}},
				},
				Args: []string{"sulcud"},
			},
			{
				Name: "Callback error",
				Root: commands.Command{
					Args: []commands.Value{{commands.Type(0xff), "int", ""}},
					Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
						err = errors.New("error")
						return
					},
				},
				Args: []string{},
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
