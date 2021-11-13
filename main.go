package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/ProtoconNet/mitum-did/cmds"
	currencycmds "github.com/spikeekips/mitum-currency/cmds"
	mitumcmds "github.com/spikeekips/mitum/launch/cmds"
	"github.com/spikeekips/mitum/util"
)

var (
	Version = "v0.0.0"
	options = []kong.Option{
		kong.Name("mitum-currency"),
		kong.Description("mitum-currency tool"),
		currencycmds.KeyAddressVars,
		currencycmds.SendVars,
		mitumcmds.BlockDownloadVars,
	}
)

type mainflags struct {
	Version VersionCommand   `cmd:"" help:"version"`
	Node    cmds.NodeCommand `cmd:"" help:"node"`
	// TODO Blocks mitumcmds.BlocksCommand `cmd:"" help:"get block data from node"`
	Key     currencycmds.KeyCommand     `cmd:"" help:"key"`
	Seal    cmds.SealCommand            `cmd:"" help:"seal"`
	Storage currencycmds.StorageCommand `cmd:"" help:"storage"`
	Deploy  currencycmds.DeployCommand  `cmd:"" help:"deploy"`
}

func main() {
	nodeCommand, err := cmds.NewNodeCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err) // revive:disable-line:unhandled-error

		os.Exit(1)
	}

	flags := mainflags{
		Node:    nodeCommand,
		Key:     currencycmds.NewKeyCommand(),
		Seal:    cmds.NewSealCommand(),
		Storage: currencycmds.NewStorageCommand(),
		Deploy:  currencycmds.NewDeployCommand(),
	}

	kctx, err := mitumcmds.Context(os.Args[1:], &flags, options...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %+v\n", err) // revive:disable-line:unhandled-error

		os.Exit(1)
	}

	version := util.Version(Version)
	if err := version.IsValid(nil); err != nil {
		kctx.FatalIfErrorf(err)
	}

	if err := kctx.Run(version); err != nil {
		kctx.FatalIfErrorf(err)
	}

	os.Exit(0)
}

type VersionCommand struct{}

func (*VersionCommand) Run() error {
	version := util.Version(Version)

	_, _ = fmt.Fprintln(os.Stdout, version)

	return nil
}
