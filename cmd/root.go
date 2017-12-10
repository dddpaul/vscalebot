package cmd

import (
	"fmt"
	"os"

	"errors"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

type VscaleAccount struct {
	Name  string
	Token string
}

type arrayFlags []string

func (flags *arrayFlags) String() string {
	return strings.Join(*flags, ", ")
}

func (flags *arrayFlags) Set(value string) error {
	*flags = append(*flags, value)
	return nil
}

func (flags *arrayFlags) toMap() ([]*VscaleAccount, error) {
	var accounts []*VscaleAccount
	for _, s := range *flags {
		items := strings.Split(s, "=")
		if len(items) == 0 {
			return nil, errors.New("incorrect Vscale name to token map format")
		}
		accounts = append(accounts, &VscaleAccount{
			Name:  items[0],
			Token: items[1],
		})
	}
	return accounts, nil
}

var (
	cfgFile         string
	verbose         bool
	telegramToken   string
	accountsStrings []string

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "vscalebot",
		Short: "Telegram Vscale Bot",
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.vscalebot.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable bot debug")
	rootCmd.PersistentFlags().StringVarP(&telegramToken, "telegram-token", "t", "", "Telegram API token")
	rootCmd.PersistentFlags().StringSliceVarP(&accountsStrings, "vscale-account", "a", nil, "List of Vscale name to token maps, i.e. 'swarm=123456'")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".vscalebot" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".vscalebot")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
