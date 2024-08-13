package mount

import (
	"fmt"
	"log"
	"os"
	"path"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	cliflags "github.com/RogueTeam/guardian/cmd/guardian/flags"
	"github.com/RogueTeam/guardian/cmd/guardian/utils"
	"github.com/RogueTeam/guardian/crypto"
	"github.com/RogueTeam/guardian/database"
	"github.com/RogueTeam/guardian/internal/commands"
	"github.com/RogueTeam/guardian/internal/utils/cli"
	"github.com/RogueTeam/guardian/mount"
)

var defaultArgon = crypto.DefaultArgon()

var MountCommand = &commands.Command{
	Name:        "mount",
	Description: "Experimental mount utility",
	Args: commands.Values{
		{Type: commands.TypeString, Name: cliflags.MountPoint, Description: "Mount point"},
	},
	Flags: commands.Values{
		{Type: commands.TypeString, Name: cliflags.Secrets, Description: "Secrets database to use", Default: path.Join(cli.Home(), "guardian.json")},
		{Type: commands.TypeInt, Name: cliflags.SaltSize, Description: "Size of the random salt to read", Default: crypto.DefaultSaltSize},
		{Type: commands.TypeInt, Name: cliflags.ArgonTime, Description: "Argon time config", Default: int(defaultArgon.Time)},
		{Type: commands.TypeInt, Name: cliflags.ArgonMemory, Description: "Argon memory config", Default: int(defaultArgon.Memory)},
		{Type: commands.TypeInt, Name: cliflags.ArgonThreads, Description: "Argon threads config", Default: int(defaultArgon.Threads)},
		{Type: commands.TypeBool, Name: cliflags.NoPrompt, Description: "No password prompt", Default: false},
	},
	Setup: func(ctx *commands.Context, flags map[string]any) (err error) {
		ctx.Set(cliflags.Secrets, flags[cliflags.Secrets])
		ctx.Set(cliflags.SaltSize, flags[cliflags.SaltSize])
		ctx.Set(cliflags.ArgonTime, flags[cliflags.ArgonTime])
		ctx.Set(cliflags.ArgonMemory, flags[cliflags.ArgonMemory])
		ctx.Set(cliflags.ArgonThreads, flags[cliflags.ArgonThreads])
		ctx.Set(cliflags.NoPrompt, flags[cliflags.NoPrompt])

		err = utils.SetupDB(ctx, flags)
		if err != nil {
			err = fmt.Errorf("failed to setup database: %w", err)
			return
		}

		return
	},
	Callback: func(ctx *commands.Context, flags, args map[string]any) (result any, err error) {
		var config mount.Config
		config.Database = ctx.MustGet(cliflags.Db).(*database.Database)
		config.File = ctx.MustGet(cliflags.File).(*os.File)
		f, err := mount.New(config)
		if err != nil {
			err = fmt.Errorf("failed to create fs: %w", err)
			return
		}

		mountPoint := args[cliflags.MountPoint].(string)
		c, err := fuse.Mount(
			mountPoint,
			fuse.FSName(mount.Name),
			fuse.Subtype(mount.Type),
		)
		if err != nil {
			err = fmt.Errorf("failed to mount location: %w", err)
			return
		}
		defer func() {
			go c.Close()
			go fuse.Unmount(mountPoint)
		}()

		log.Println("Mounting")
		err = fs.Serve(c, f)
		if err != nil {
			err = fmt.Errorf("failed to serve mount point: %w", err)
			return
		}

		return
	},
}
