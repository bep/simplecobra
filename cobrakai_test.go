package cobrakai_test

import (
	"context"
	"fmt"
	"testing"

	ck "github.com/bep/cobrakai"
	qt "github.com/frankban/quicktest"
	"github.com/spf13/pflag"
)

func TestCobraKai(t *testing.T) {

	var (
		fooCommand    = &testComand1{name: "foo"}
		barCommand    = &testComand1{name: "bar"}
		fooBazCommand = &testComand2{name: "foo_baz"}
	)

	c := qt.New(t)
	r, err := ck.R(
		&testComand1{name: "hugo"},
		ck.C(fooCommand,
			ck.C(fooBazCommand),
		),
		ck.C(barCommand),
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

func (c *testComand1) AddFlagsLocal(flags *pflag.FlagSet) {
	flags.StringVar(&c.localFlagName, "localFlagName", "", "set localFlagName")
}

func (c *testComand1) AddFlagsPersistent(flags *pflag.FlagSet) {
	flags.StringVar(&c.persistentFlagName, "persistentFlagName", "", "set persistentFlagName")
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

func (c *testComand2) AddFlagsLocal(flags *pflag.FlagSet) {
	flags.StringVar(&c.localFlagName, "localFlagName", "", "set localFlagName for testCommand2")
}

func (c *testComand2) AddFlagsPersistent(flags *pflag.FlagSet) {

}
