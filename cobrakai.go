package cobrakai

import (
	"context"

	"github.com/spf13/cobra"
)

// Executer is the execution entry point.
// The args are usually filled with os.Args[1:].
type Executer interface {
	Execute(ctx context.Context, args []string) (*Commandeer, error)
}

// Commander is the interface that must be implemented by all commands.
type Commander interface {
	Name() string
	Run(ctx context.Context, args []string) error
	WithCobraCommand(*cobra.Command) error
}

type root struct {
	c *Commandeer
}

func (r *root) Execute(ctx context.Context, args []string) (*Commandeer, error) {
	r.c.CobraCommand.SetArgs(args)
	cobraCommand, err := r.c.CobraCommand.ExecuteContextC(ctx)
	if err != nil {
		return nil, err
	}
	// Find the commandeer that was executed.
	var find func(*cobra.Command, *Commandeer) *Commandeer
	find = func(what *cobra.Command, in *Commandeer) *Commandeer {
		if in.CobraCommand == what {
			return in
		}
		for _, in2 := range in.commandeers {
			if found := find(what, in2); found != nil {
				return found
			}
		}
		return nil
	}
	return find(cobraCommand, r.c), nil
}

// Commandeer holds the state of a command and its subcommands.
type Commandeer struct {
	Command      Commander
	CobraCommand *cobra.Command
	commandeers  []*Commandeer
}

func (c *Commandeer) compile() error {
	c.CobraCommand = &cobra.Command{
		Use: c.Command.Name(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Command.Run(cmd.Context(), args)
		},
	}

	// This is where the flags, short and long description etc. are added
	c.Command.WithCobraCommand(c.CobraCommand)

	// Add commands recursively.
	for _, cc := range c.commandeers {
		if err := cc.compile(); err != nil {
			return err
		}
		c.CobraCommand.AddCommand(cc.CobraCommand)
	}

	return nil
}

// WithCommandeer allows chaining of commandeers.
type WithCommandeer func(*Commandeer)

// R creates the execution entry poing given a root command and a chain of nested commands.
func R(command Commander, wcs ...WithCommandeer) (Executer, error) {
	c := &Commandeer{
		Command: command,
	}
	for _, wc := range wcs {
		wc(c)
	}
	if err := c.compile(); err != nil {
		return nil, err
	}
	return &root{c: c}, nil
}

// C creates nested commands.
func C(command Commander, wcs ...WithCommandeer) WithCommandeer {
	return func(parent *Commandeer) {
		cd := &Commandeer{
			Command: command,
		}
		parent.commandeers = append(parent.commandeers, cd)
		for _, wc := range wcs {
			wc(cd)
		}
	}
}

// SimpleCommand creates a simple command that does not take any flags.
func SimpleCommand(name string, run func(ctx context.Context, args []string) error) Commander {
	return &simpleCommand{
		name: name,
		run:  run,
	}
}

type simpleCommand struct {
	name string
	run  func(ctx context.Context, args []string) error
}

func (c *simpleCommand) Name() string {
	return c.name
}

func (c *simpleCommand) Run(ctx context.Context, args []string) error {
	return c.run(ctx, args)
}

func (c *simpleCommand) WithCobraCommand(cmd *cobra.Command) error {
	return nil
}
