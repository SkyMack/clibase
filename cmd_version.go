package clibase

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	flagPackageScopeName = "package-prefix"
)

func version(name string, flags *pflag.FlagSet) error {
	packPrefix, err := flags.GetString(flagPackageScopeName)
	if err != nil {
		LogFlagError(flagPackageScopeName, err)
		return err
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		log.Debug("binary not built with module support")
		return nil
	}
	fmt.Printf("%s (%s %s)\n", name, buildInfo.Main.Path, buildInfo.Main.Version)

	fmt.Printf("\n")
	fmt.Printf("  Compiled with: %s\n", runtime.Compiler)
	fmt.Printf("         GOARCH: %s\n", runtime.GOARCH)
	fmt.Printf("           GOOS: %s\n", runtime.GOOS)
	fmt.Printf("     Go Version: %s\n", runtime.Version())
	fmt.Printf("\n")

	for _, pkg := range buildInfo.Deps {
		if !strings.HasPrefix(pkg.Path, packPrefix) {
			continue
		}
		output := fmt.Sprintf("%s %s", pkg.Path, pkg.Version)
		if pkg.Replace != nil {
			var struckthrough string
			for _, r := range output {
				struckthrough += "\u0336" + string(r)
			}
			output = fmt.Sprintf("%s\u0336  => %s", struckthrough, pkg.Replace.Path)
		}
		fmt.Printf("  %s\n", output)
	}
	return nil
}

func addVersionFlags(flags *pflag.FlagSet) {
	verFlags := &pflag.FlagSet{}

	verFlags.String(flagPackageScopeName, "github.com/SkyMack", "only introspect packages under this prefix")

	SetFlagsFromEnv(cliBaseFlagPrefix, verFlags)

	flags.AddFlagSet(verFlags)
}

func addVersionCmd(rootCmd *cobra.Command) {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "output the binary version and dependency details",
		RunE: func(cmd *cobra.Command, args []string) error {
			return version(rootCmd.Name(), cmd.Flags())
		},
	}

	addVersionFlags(versionCmd.Flags())
	rootCmd.AddCommand(versionCmd)
}
