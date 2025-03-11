package main

import (
	"os"

	"github.com/kiichain/kiichain/app/params"
	"github.com/kiichain/kiichain/cmd/kiichaind/cmd"

	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	"github.com/kiichain/kiichain/app"
)

func main() {
	params.SetAddressPrefixes()
	rootCmd, _ := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, app.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
