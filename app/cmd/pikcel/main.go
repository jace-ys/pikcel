package main

import (
	"context"
	"io"
	"os"

	"github.com/alecthomas/kong"

	"github.com/jace-ys/pikcel/api/v1/gen/api"
	"github.com/jace-ys/pikcel/internal/ctxlog"
)

type RootCmd struct {
	Globals

	Server  ServerCmd  `cmd:"" help:"Run the pikcel server."`
	Version VersionCmd `cmd:"" help:"Show version information."`
}

type Globals struct {
	Debug  bool      `env:"DEBUG" help:"Enable debug logging."`
	Writer io.Writer `kong:"-"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	root := RootCmd{
		Globals: Globals{
			Writer: os.Stdout,
		},
	}

	cli := kong.Parse(&root,
		kong.Name(api.APIName),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:             true,
			FlagsLast:           true,
			NoExpandSubcommands: true,
		}),
	)

	ctx = ctxlog.NewContext(ctx, root.Writer, root.Debug)

	cli.BindTo(ctx, (*context.Context)(nil))
	cli.FatalIfErrorf(cli.Run(&root.Globals))
}
