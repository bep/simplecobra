[![Tests on Linux, MacOS and Windows](https://github.com/bep/simplecobra/workflows/Test/badge.svg)](https://github.com/bep/simplecobra/actions?query=workflow:Test)
[![Go Report Card](https://goreportcard.com/badge/github.com/bep/simplecobra)](https://goreportcard.com/report/github.com/bep/simplecobra)
[![GoDoc](https://godoc.org/github.com/bep/simplecobra?status.svg)](https://godoc.org/github.com/bep/simplecobra)

So, [Cobra](https://github.com/spf13/cobra) is a Go CLI library with a feature set that's hard to resist for bigger applications (autocomplete, docs auto generation etc.). But it's also rather complex to use beyond the simplest of applications. This package is built to aid rewriting [Hugo's](https://github.com/gohugoio/hugo) commands package to something that's easier to understand and maintain.

I welcome suggestions to improve/simplify this further, but the core idea is that the command graph gets built in one go with a tree of struct pointers implementing a simple `Commander` interface:

```go
// Commander is the interface that must be implemented by all commands.
type Commander interface {
	// The name of the command.
	Name() string

	// The command execution.
	Run(ctx context.Context, cd *Commandeer, args []string) error

	// Init called on all ancestors and the executing command itself, before execution, starting from the root.
	// This is the place to evaluate flags and set up the command.
	Init(*Commandeer) error

	// WithCobraCommand is called when the cobra command is created.
	// This is where the flags, short and long description etc. are added.
	WithCobraCommand(*cobra.Command) error

	// Commands returns the sub commands, if any.
	Commands() []Commander
}
```

The `Init` method allows for flag compilation, referencing the parent and root etc. If needed, the full Cobra command is still available for configuration in `WithCobraCommand`.

There's a runnable [example](https://pkg.go.dev/github.com/bep/simplecobra#example-package) in the documentation, but the gist of it is:

```go
func main() {
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
}
```
