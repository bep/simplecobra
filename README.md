[![Tests on Linux, MacOS and Windows](https://github.com/bep/simplecobra/workflows/Test/badge.svg)](https://github.com/bep/simplecobra/actions?query=workflow:Test)
[![Go Report Card](https://goreportcard.com/badge/github.com/bep/simplecobra)](https://goreportcard.com/report/github.com/bep/simplecobra)
[![codecov](https://codecov.io/gh/bep/simplecobra/branch/master/graph/badge.svg)](https://codecov.io/gh/bep/simplecobra)
[![GoDoc](https://godoc.org/github.com/bep/simplecobra?status.svg)](https://godoc.org/github.com/bep/simplecobra)

So, [Cobra](https://github.com/spf13/cobra) is a Go CLI library with a feature set that's hard to resist for bigger applications (autocompletion, docs and man pages auto generation etc.). But it's also complex to use beyond the simplest of applications. This package was built to help rewriting [Hugo's](https://github.com/gohugoio/hugo) commands package to something that's easier to understand and maintain.

I welcome suggestions to improve/simplify this further, but the core idea is that the command graph gets built in one go with a tree of struct pointers implementing a simple `Commander` interface:

```go
// Commander is the interface that must be implemented by all commands.
type Commander interface {
	// The name of the command.
	Name() string

	// Init is called when the cobra command is created.
	// This is where the flags, short and long description etc. can be added.
	Init(*Commandeer) error

	// PreRun called on all ancestors and the executing command itself, before execution, starting from the root.
	// This is the place to evaluate flags and set up the this Commandeer.
	// The runner Commandeer holds the currently running command, which will be PreRun last.
	PreRun(this, runner *Commandeer) error

	// The command execution.
	Run(ctx context.Context, cd *Commandeer, args []string) error

	// Commands returns the sub commands, if any.
	Commands() []Commander
}
```

The `Init` method allows for flag compilation, referencing the parent and root etc. If needed, the full Cobra command is still available.

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


## Differences to Cobra

You have access to the `*cobra.Command` pointer so there's not much you cannot do with this project compared to the more low-level Cobra, but there's one small, but imortant difference:

Cobra only treats the first level of misspelled commands as an `unknown command` with "Did you mean this?" suggestions, see [see this issue](https://github.com/spf13/cobra/pull/1500) for more context. The reason this is, is because of the ambiguity between sub command names and command arguments, but that is throwing away a very useful feature for a not very good reason. We recently rewrote [Hugo's CLI](https://github.com/gohugoio/hugo) using this package, and found only one sub command that needed to be adjusted to avoid this ambiguity.



