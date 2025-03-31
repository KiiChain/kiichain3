package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/sirkon/goproxy/gomod"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	flagFormat = "format"

	pathCosmosSDK = "github.com/cosmos/cosmos-sdk"
)

var (
	// Version defines the application version (defined at compile time)
	Version = ""

	// Commit defines the application commit hash (defined at compile time)
	Commit = ""

	versionFormat string
)

// versionInfo represent the object which store the project's version info
type versionInfo struct {
	Version string `json:"version" yaml:"version"`
	Commit  string `json:"commit" yaml:"commit"`
	SDK     string `json:"sdk" yaml:"sdk"`
	Go      string `json:"go" yaml:"go"`
}

// CmdgetVersion is the command executed when users will type "version subcommand"
func CmdgetVersion() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print binary version information",
		RunE:  getVersionCmdHandler,
	}

	versionCmd.Flags().StringVar(&versionFormat, flagFormat, "text", "Print the version in the given format (text|json)")

	return versionCmd
}

// getVersionCmdHandler gets the project and go version from the go.mod file
func getVersionCmdHandler(cmd *cobra.Command, args []string) error {
	// get go.mod file
	modBz, err := os.ReadFile("go.mod")
	if err != nil {
		return err
	}

	mod, err := gomod.Parse("go.mod", modBz)
	if err != nil {
		return err
	}

	// set version info
	verInfo := versionInfo{
		Version: Version,
		Commit:  Commit,
		SDK:     mod.Require[pathCosmosSDK],
		Go:      fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
	}

	var bz []byte

	// print on the selected log format
	switch versionFormat {
	case "json":
		bz, err = json.Marshal(verInfo)

	default:
		bz, err = yaml.Marshal(&verInfo)
	}
	if err != nil {
		return err
	}

	// print on the console
	_, err = fmt.Println(string(bz))
	return err
}
