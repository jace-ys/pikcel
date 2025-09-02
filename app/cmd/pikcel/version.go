package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/jace-ys/pikcel/internal/versioninfo"
)

type VersionCmd struct{}

func (c *VersionCmd) Run(_ context.Context, _ *Globals) error {
	fmt.Printf( //nolint:forbidigo
		"Version: %v\nCommit SHA: %v\nGo Version: %v\nGo OS/Arch: %v/%v\n",
		versioninfo.Version, versioninfo.CommitSHA, runtime.Version(), runtime.GOOS, runtime.GOARCH,
	)
	return nil
}
