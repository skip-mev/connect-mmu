package main

import (
	"errors"
	"os"

	"github.com/skip-mev/connect-mmu/cmd/mmu/cmd"
	"github.com/skip-mev/connect-mmu/signing"
	"github.com/skip-mev/connect-mmu/signing/simulate"
)

func main() {
	r := signing.NewRegistry()
	err := errors.Join(
		r.RegisterSigner(simulate.TypeName, simulate.NewSigningAgent),
	)
	if err != nil {
		panic(err)
	}
	if err := cmd.RootCmd(r).Execute(); err != nil {
		os.Exit(1)
	}
}
