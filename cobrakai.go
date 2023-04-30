package cobrakai

import (
	"context"

	"github.com/spf13/cobra"
)

// Commander is the interface that must be implemented by all commands.
type Commander interface {
	// The name of the command.
	Name() string

	// The command execution.
	Run(ctx context.Context, args []string) error

	// Init called on all commands in this tree, before execution, starting from the root.
	// This is the place to evaluate flags and set up the command.
	Init(*Commandeer) error

	// WithCobraCommand is called when the cobra command is created.
	// This is where the flags, short and long description etc. are added.
	WithCobraCommand(*cobra.Command) error

	// Commands returns the sub commands, if any.
	Commands() []Commander
}

// Executer is the execution entry point.
// The args are usually filled with os.Args[1:].
type Executer interface {
	Execute(ctx context.Context, args []string) (*Commandeer, error)
}

// New creates a new Executer from the command tree in Commander.
func New(rootCmd Commander) (Executer, error) {
	rootCd := &Commandeer{
		Command: rootCmd,
	}
	rootCd.Root = rootCd

	// Add all commands recursively.
	var addCommands func(cd *Commandeer, cmd Commander)
	addCommands = func(cd *Commandeer, cmd Commander) {
		cd2 := &Commandeer{
			Root:    rootCd,
			Parent:  cd,
			Command: cmd,
		}
		cd.commandeers = append(cd.commandeers, cd2)
		for _, c := range cmd.Commands() {
			addCommands(cd2, c)
		}

	}

	for _, cmd := range rootCmd.Commands() {
		addCommands(rootCd, cmd)
	}

	if err := rootCd.compile(); err != nil {
		return nil, err
	}

	return &root{c: rootCd}, nil

}

// Commandeer holds the state of a command and its subcommands.
type Commandeer struct {
	Command      Commander
	CobraCommand *cobra.Command

	Root        *Commandeer
	Parent      *Commandeer
	commandeers []*Commandeer
}

func (c *Commandeer) init() error {
	// Start from the root and initialize all commands recursively.
	// root is always set.
	cd := c.Root
	var initc func(*Commandeer) error
	initc = func(cd *Commandeer) error {
		if err := cd.Command.Init(cd); err != nil {
			return err
		}
		for _, cc := range cd.commandeers {
			if err := initc(cc); err != nil {
				return err
			}
		}
		return nil
	}
	return initc(cd)
}

func (c *Commandeer) compile() error {
	c.CobraCommand = &cobra.Command{
		Use: c.Command.Name(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.Command.Run(cmd.Context(), args)
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return c.init()
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
