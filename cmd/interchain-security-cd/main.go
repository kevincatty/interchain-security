package main

import (
	"fmt"
	"os"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	app "github.com/cosmos/interchain-security/v7/app/consumer"
	appparams "github.com/cosmos/interchain-security/v7/app/params"
	"github.com/cosmos/interchain-security/v7/cmd/interchain-security-cd/cmd"
)

func main() {
	appparams.SetAddressPrefixes(app.Bech32MainPrefix)

	rootCmd := cmd.NewRootCmd()

	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		fmt.Fprintln(rootCmd.OutOrStderr(), err)
		os.Exit(1)
	}
}
