package cobrakai_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/bep/cobrakai"
	qt "github.com/frankban/quicktest"
	"github.com/spf13/cobra"
)

func TestCobraKai(t *testing.T) {

	var (
		fooCommand    = &testComand1{name: "foo"}
		barCommand    = &testComand1{name: "bar"}
		fooBazCommand = &testComand2{name: "foo_baz"}
	)

	c := qt.New(t)
	r, err := cobrakai.R(
		&testComand1{name: "hugo"}, // The root command.
		cobrakai.C(
			fooCommand,
			cobrakai.C(
				fooBazCommand),
		),
		cobrakai.C(barCommand),
	)
	c.Assert(err, qt.IsNil)

	// This can be anything, just used to make sure the same context is passed all the way.
	type key string
	ctx := context.WithValue(context.Background(), key("foo"), "bar")
	args := []string{"foo", "--localFlagName", "foo_local", "--persistentFlagName", "foo_persistent"}
	cdeer, err := r.Execute(ctx, args)
	c.Assert(err, qt.IsNil)
	c.Assert(cdeer.Command.Name(), qt.Equals, "foo")
	tc := cdeer.Command.(*testComand1)
	c.Assert(tc.ctx, qt.Equals, ctx)
	c.Assert(tc.localFlagName, qt.Equals, "foo_local")
	c.Assert(tc.persistentFlagName, qt.Equals, "foo_persistent")

	args = []string{"foo", "foo_baz", "--localFlagName", "foo_local2", "--persistentFlagName", "foo_persistent2"}
	ctx = context.WithValue(context.Background(), key("bar"), "baz")
	cdeer2, err := r.Execute(ctx, args)
	c.Assert(err, qt.IsNil)
	c.Assert(cdeer2.Command.Name(), qt.Equals, "foo_baz")
	tc2 := cdeer2.Command.(*testComand2)
	c.Assert(tc2.ctx, qt.Equals, ctx)
	c.Assert(tc2.localFlagName, qt.Equals, "foo_local2")
	c.Assert(tc.persistentFlagName, qt.Equals, "foo_persistent2")

}

func ExampleSimpleCommand() {
	r, err := cobrakai.R(
		// If you need flags, implement cobrakai.Commander.
		cobrakai.SimpleCommand("root", func(ctx context.Context, args []string) error { fmt.Print("run root "); return nil }),
		cobrakai.C(cobrakai.SimpleCommand("sub1", func(ctx context.Context, args []string) error { fmt.Print("run sub1"); return nil })),
		cobrakai.C(cobrakai.SimpleCommand("sub2", func(ctx context.Context, args []string) error { fmt.Print("run sub2"); return nil })),
	)

	if err != nil {
		log.Fatal(err)
	}

	if _, err := r.Execute(context.Background(), []string{""}); err != nil {
		log.Fatal(err)
	}
	if _, err := r.Execute(context.Background(), []string{"sub1"}); err != nil {
		log.Fatal(err)
	}
	// Output: run root run sub1

}

type testComand1 struct {
	persistentFlagName string
	localFlagName      string

	ctx  context.Context
	name string
}

func (c *testComand1) Run(ctx context.Context, args []string) error {
	c.ctx = ctx
	fmt.Println("testComand.Run", c.name, args)
	return nil
}

func (c *testComand1) Name() string {
	return c.name
}

func (c *testComand1) WithCobraCommand(cmd *cobra.Command) error {
	localFlags := cmd.Flags()
	persistentFlags := cmd.PersistentFlags()

	localFlags.StringVar(&c.localFlagName, "localFlagName", "", "set localFlagName")
	persistentFlags.StringVar(&c.persistentFlagName, "persistentFlagName", "", "set persistentFlagName")

	return nil
}

type testComand2 struct {
	localFlagName string

	ctx  context.Context
	name string
}

func (c *testComand2) Run(ctx context.Context, args []string) error {
	c.ctx = ctx
	fmt.Println("testComand2.Run", c.name, args)
	return nil
}

func (c *testComand2) Name() string {
	return c.name
}

func (c *testComand2) WithCobraCommand(cmd *cobra.Command) error {
	localFlags := cmd.Flags()
	localFlags.StringVar(&c.localFlagName, "localFlagName", "", "set localFlagName for testCommand2")
	return nil
}
