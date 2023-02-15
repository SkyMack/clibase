package clibase

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	cliBaseFlagPrefix = "BASE_"
)

var (
	// ErrorFlagCannotRetrieve is the error logged when attempting to retrieve the value of a flag fails
	ErrorFlagCannotRetrieve = fmt.Errorf("cannot retrieve flag value")
)

// AddTopLevelFlags takes a pointer to an existing pflag.FlagSet and adds the default top level flags to it
func AddTopLevelFlags(flags *pflag.FlagSet) {
	topLevelFlags := &pflag.FlagSet{}

	addLogFlags(topLevelFlags)
	flags.AddFlagSet(topLevelFlags)
}

// New returns a new Cobra root command with all the defaults
func New(cmdName, cmdDescription string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdName,
		Short: cmdDescription,
	}
	return NewUsingCmd(cmd)
}

// NewUsingCmd takes an existing Cobra command and adds in the default flags, subcommands, and Init/Run entries
func NewUsingCmd(rootCmd *cobra.Command) *cobra.Command {
	var persistentPreRunE func(*cobra.Command, []string) error
	if rootCmd.PersistentPreRunE != nil {
		existingPersistentPreRunE := rootCmd.PersistentPreRunE
		persistentPreRunE = func(cmd *cobra.Command, args []string) error {
			if err := rootPersistentPreRunE(cmd, args); err != nil {
				return err
			}
			return existingPersistentPreRunE(cmd, args)
		}
	} else {
		persistentPreRunE = rootPersistentPreRunE
	}

	rootCmd.PersistentPreRunE = persistentPreRunE
	AddTopLevelFlags(rootCmd.PersistentFlags())
	addVersionCmd(rootCmd)
	return rootCmd
}

// LogFlagError generates a log entry for an error related to retrieving a flag value
func LogFlagError(flagName string, err error) {
	log.WithFields(log.Fields{
		"flag.name": flagName,
		"error":     err,
	}).Error(ErrorFlagCannotRetrieve.Error())
}

func rootPersistentPreRunE(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	logFormat, err := flags.GetString(logFlagFormatName)
	if err != nil {
		LogFlagError(logFlagFormatName, err)
		return err
	}
	logLevel, err := flags.GetString(logFlagLevelName)
	if err != nil {
		LogFlagError(logFlagLevelName, err)
		return err
	}

	checkCobraFlags(flags)

	return configureLogging(logFormat, logLevel)
}

// checkCobraFlags logs warnings for flags that don't follow style conventions
func checkCobraFlags(flags *pflag.FlagSet) {
	flags.VisitAll(func(flag *pflag.Flag) {
		logLn := log.WithField("flag.name", flag.Name)
		logLn.Tracef("checking flag for style")

		// all flag names should be lowercase
		for _, rune := range flag.Name {
			if unicode.IsLetter(rune) && !unicode.IsLower(rune) {
				logLn.WithField("violation", "flag names must be all lower case").Warn("invalid flag name")
			}
		}

		// don't use --foo_bar, use --foo-bar
		if strings.Index(flag.Name, "_") > 0 {
			logLn.WithField("violation", "flag names must use hyphen not underscore").Warn("invalid flag name")
		}
	})
}

// SetFlagsFromEnvWithOverrides sets the default value for each flag in the flagset based on a matching environment variable.
// The expected env var name will be the flag's name, all uppercase, with hyphens replaced by underscores
// (i.e. flag "foo-bar" will be matched with env var "FOO_BAR")
// If set, prefix will be appended to the expected name of each env variable
// (i.e. prefix of baz and flag "foobar" will be marched with env var "BAZ_FOOBAR")
// If overrides has a key that matches the flag name, the value of that key will be used as the expected env var name
// (i.e. override entry of "foobar=MY_SPECIAL_FLAG" will result in the flag "foobar" being matched with the env var "MY_SPECIAL_FLAG")
// Override is for when you want all the flags to have a prefix, but need some specific flags to not use that prefix
// (or have different env vars entirely). For example, having the application recognize a standard env var like "REDIS_URL"
// while prefacing most app specific flags with APPNAME_
func SetFlagsFromEnvWithOverrides(prefix string, flagSet *pflag.FlagSet, overrides map[string]string) {
	flagSet.VisitAll(func(flag *pflag.Flag) {
		envName, ok := overrides[flag.Name]
		if !ok {
			envName = fmt.Sprintf("%s%s", prefix, strings.Replace(strings.ToUpper(flag.Name), "-", "_", -1))
		}

		logLn := log.WithFields(log.Fields{
			"flag.name":  flag.Name,
			"flag.usage": flag.Usage,
			"flag.value": flag.Value,
			"env.name":   envName,
		})

		flag.Usage = fmt.Sprintf("%s (${%s})", flag.Usage, envName)

		value := os.Getenv(envName)
		if value == "" {
			return
		}

		logLn = logLn.WithField("env.value", value)

		logLn.Debug("Updating with the environment value")
		if err := flag.Value.Set(value); err != nil {
			logLn.WithField("error", err).Warn("failed to set flag from environment variable")
		}
	})
}

// SetFlagsFromEnv calls SetFlagsFromEnvWithOverrides and passes an empty overrides map
func SetFlagsFromEnv(prefix string, flags *pflag.FlagSet) {
	SetFlagsFromEnvWithOverrides(prefix, flags, map[string]string{})
}

// EnvNameForFlag generates an expected environment variable name, given the provided prefix and flag
func EnvNameForFlag(prefix string, flag *pflag.Flag) string {
	return fmt.Sprintf("%s%s", prefix, strings.Replace(strings.ToUpper(flag.Name), "-", "_", -1))
}
