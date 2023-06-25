package serve

import (
	"context"
	"flag"

	"github.com/google/subcommands"
	"go.seankhliao.com/svcrunner/v2/basehttp"
)

type Cmd struct {
	basehttp basehttp.Config

	host   string
	source string
}

func (c *Cmd) Name() string     { return `serve` }
func (c *Cmd) Synopsis() string { return `start server` }
func (c *Cmd) Usage() string {
	return `serve [options...]

Starts a server managing listening records

Flags:
`
}

func (c *Cmd) SetFlags(f *flag.FlagSet) {
	c.basehttp.SetFlags(f)
	f.StringVar(&c.host, "vanity.host", "go.seankhliao.com", "host this server runs on")
	f.StringVar(&c.source, "vanity.source", "github.com/seankhliao", "where the code is hosted")
}

func (c *Cmd) Execute(ctx context.Context, f *flag.FlagSet, args ...any) subcommands.ExitStatus {
	err := New(ctx, c).Run(ctx)
	if err != nil {
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
