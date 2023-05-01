package simplecobra_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/bep/simplecobra"
	qt "github.com/frankban/quicktest"
	"github.com/spf13/cobra"
)

func testCommands() *rootCommand {
	return &rootCommand{name: "root",
		commands: []simplecobra.Commander{
			&lvl1Command{name: "foo"},
			&lvl1Command{name: "bar",
				commands: []simplecobra.Commander{
					&lvl2Command{name: "baz"},
				},
			},
		},
	}

}

func TestSimplecobra(t *testing.T) {
	c := qt.New(t)

	rootCmd := testCommands()

	x, err := simplecobra.New(rootCmd)
	c.Assert(err, qt.IsNil)
	// This can be anything, just used to make sure the same context is passed all the way.
	type key string
	ctx := context.WithValue(context.Background(), key("foo"), "bar")
	// Execute the root command.
	args := []string{"--localFlagName", "root_local", "--persistentFlagName", "root_persistent"}
	cd, err := x.Execute(ctx, args)
	c.Assert(err, qt.IsNil)
	c.Assert(cd.Command.Name(), qt.Equals, "root")
	tc := cd.Command.(*rootCommand)
	c.Assert(tc, qt.Equals, rootCmd)
	c.Assert(tc.ctx, qt.Equals, ctx)
	c.Assert(tc.localFlagName, qt.Equals, "root_local")
	c.Assert(tc.persistentFlagName, qt.Equals, "root_persistent")
	c.Assert(tc.persistentFlagNameC, qt.Equals, "root_persistent_rootCommand_compiled")
	c.Assert(tc.localFlagNameC, qt.Equals, "root_local_rootCommand_compiled")

	// Execute a level 1 command.
	// This may not be very realistic, but it works. The common use case for a CLI app is to run one command and then exit.
	args = []string{"bar", "--localFlagName", "bar_local", "--persistentFlagName", "bar_persistent"}
	ctx = context.WithValue(context.Background(), key("bar"), "baz")
	cd2, err := x.Execute(ctx, args)
	c.Assert(err, qt.IsNil)
	c.Assert(cd2.Command.Name(), qt.Equals, "bar")
	tc2 := cd2.Command.(*lvl1Command)
	c.Assert(tc2.rootCmd, qt.Equals, rootCmd)
	c.Assert(tc2.ctx, qt.Equals, ctx)
	c.Assert(tc2.localFlagName, qt.Equals, "bar_local")
	c.Assert(tc2.localFlagNameC, qt.Equals, "bar_local_lvl1Command_compiled")
	c.Assert(tc.persistentFlagName, qt.Equals, "bar_persistent")
	c.Assert(tc.persistentFlagNameC, qt.Equals, "bar_persistent_rootCommand_compiled")

	// Execute a level 2 command.
	args = []string{"bar", "baz", "--persistentFlagName", "baz_persistent"}
	ctx = context.WithValue(context.Background(), key("baz"), "qux")
	cd3, err := x.Execute(ctx, args)
	c.Assert(err, qt.IsNil)
	c.Assert(cd3.Command.Name(), qt.Equals, "baz")
	tc3 := cd3.Command.(*lvl2Command)
	c.Assert(tc3.rootCmd, qt.Equals, rootCmd)
	c.Assert(tc3.parentCmd, qt.Equals, tc2)
	c.Assert(tc3.ctx, qt.Equals, ctx)

}

func TestInitAncestorsOnly(t *testing.T) {
	c := qt.New(t)

	rootCmd := testCommands()
	x, err := simplecobra.New(rootCmd)
	c.Assert(err, qt.IsNil)
	args := []string{"bar", "baz", "--persistentFlagName", "baz_persistent"}
	cd3, err := x.Execute(context.Background(), args)
	c.Assert(err, qt.IsNil)
	c.Assert(cd3.Command.Name(), qt.Equals, "baz")
	c.Assert(rootCmd.isInit, qt.IsTrue)
	c.Assert(rootCmd.commands[0].(*lvl1Command).isInit, qt.IsFalse)
	c.Assert(rootCmd.commands[1].(*lvl1Command).isInit, qt.IsTrue)
	c.Assert(cd3.Command.(*lvl2Command).isInit, qt.IsTrue)
}

func TestErrors(t *testing.T) {
	c := qt.New(t)

	c.Run("unknown similar command", func(c *qt.C) {
		x, err := simplecobra.New(testCommands())
		c.Assert(err, qt.IsNil)
		_, err = x.Execute(context.Background(), []string{"fooo"})
		c.Assert(err, qt.Not(qt.IsNil))
		c.Assert(err.Error(), qt.Contains, "unknown")
		c.Assert(err.Error(), qt.Contains, "Did you mean this?")
		c.Assert(simplecobra.IsCommandError(err), qt.Equals, true)
	})

	c.Run("unknown similar sub command", func(c *qt.C) {
		x, err := simplecobra.New(testCommands())
		c.Assert(err, qt.IsNil)
		_, err = x.Execute(context.Background(), []string{"bar", "bazz"})
		c.Assert(err, qt.Not(qt.IsNil))
		c.Assert(err.Error(), qt.Contains, "unknown")
		c.Assert(err.Error(), qt.Contains, "Did you mean this?")
		c.Assert(simplecobra.IsCommandError(err), qt.Equals, true)
	})

	c.Run("unknown flag", func(c *qt.C) {
		x, err := simplecobra.New(testCommands())
		c.Assert(err, qt.IsNil)
		_, err = x.Execute(context.Background(), []string{"bar", "--unknown"})
		c.Assert(err, qt.Not(qt.IsNil))
		c.Assert(err.Error(), qt.Contains, "unknown")
		c.Assert(simplecobra.IsCommandError(err), qt.Equals, true)
	})

}

func Example() {
	rootCmd := &rootCommand{name: "root",
		commands: []simplecobra.Commander{
			&lvl1Command{name: "foo"},
			&lvl1Command{name: "bar",
				commands: []simplecobra.Commander{
					&lvl2Command{name: "baz"},
				},
			},
		},
	}

	x, err := simplecobra.New(rootCmd)
	if err != nil {
		log.Fatal(err)
	}
	cd, err := x.Execute(context.Background(), []string{"bar", "baz", "--localFlagName", "baz_local", "--persistentFlagName", "baz_persistent"})
	if err != nil {
		log.Fatal(err)
	}

	// These are wired up in Init().
	lvl2 := cd.Command.(*lvl2Command)
	lvl1 := lvl2.parentCmd
	root := lvl1.rootCmd

	fmt.Printf("Executed %s.%s.%s with localFlagName %s and and persistentFlagName %s.\n", root.name, lvl1.name, lvl2.name, lvl2.localFlagName, root.persistentFlagName)
	// Output: Executed root.bar.baz with localFlagName baz_local and and persistentFlagName baz_persistent.

}

type rootCommand struct {
	name   string
	isInit bool

	// Flags
	persistentFlagName string
	localFlagName      string

	// Compiled flags.
	persistentFlagNameC string
	localFlagNameC      string

	// For testing.
	ctx context.Context

	// Sub commands.
	commands []simplecobra.Commander
}

func (c *rootCommand) Commands() []simplecobra.Commander {
	return c.commands
}

func (c *rootCommand) Init(*simplecobra.Commandeer) error {
	c.isInit = true
	c.persistentFlagNameC = c.persistentFlagName + "_rootCommand_compiled"
	c.localFlagNameC = c.localFlagName + "_rootCommand_compiled"
	return nil
}

func (c *rootCommand) Name() string {
	return c.name
}

func (c *rootCommand) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	c.ctx = ctx
	return nil
}

func (c *rootCommand) WithCobraCommand(cmd *cobra.Command) error {
	localFlags := cmd.Flags()
	persistentFlags := cmd.PersistentFlags()

	localFlags.StringVar(&c.localFlagName, "localFlagName", "", "set localFlagName")
	persistentFlags.StringVar(&c.persistentFlagName, "persistentFlagName", "", "set persistentFlagName")

	return nil
}

type lvl1Command struct {
	name   string
	isInit bool

	localFlagName  string
	localFlagNameC string

	rootCmd *rootCommand

	commands []simplecobra.Commander

	ctx context.Context
}

func (c *lvl1Command) Commands() []simplecobra.Commander {
	return c.commands
}

func (c *lvl1Command) Init(cd *simplecobra.Commandeer) error {
	c.isInit = true
	c.localFlagNameC = c.localFlagName + "_lvl1Command_compiled"
	c.rootCmd = cd.Root.Command.(*rootCommand)
	return nil
}

func (c *lvl1Command) Name() string {
	return c.name
}

func (c *lvl1Command) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	c.ctx = ctx
	return nil
}

func (c *lvl1Command) WithCobraCommand(cmd *cobra.Command) error {
	localFlags := cmd.Flags()
	localFlags.StringVar(&c.localFlagName, "localFlagName", "", "set localFlagName for lvl1Command")
	return nil
}

type lvl2Command struct {
	name          string
	isInit        bool
	localFlagName string

	ctx       context.Context
	rootCmd   *rootCommand
	parentCmd *lvl1Command
}

func (c *lvl2Command) Commands() []simplecobra.Commander {
	return nil
}

func (c *lvl2Command) Init(cd *simplecobra.Commandeer) error {
	c.isInit = true
	c.rootCmd = cd.Root.Command.(*rootCommand)
	c.parentCmd = cd.Parent.Command.(*lvl1Command)
	return nil
}

func (c *lvl2Command) Name() string {
	return c.name
}

func (c *lvl2Command) Run(ctx context.Context, cd *simplecobra.Commandeer, args []string) error {
	c.ctx = ctx
	return nil
}

func (c *lvl2Command) WithCobraCommand(cmd *cobra.Command) error {
	localFlags := cmd.Flags()
	localFlags.StringVar(&c.localFlagName, "localFlagName", "", "set localFlagName for lvl2Command")
	return nil
}
