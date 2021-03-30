/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/shipa988/hw_otus_architect/cmd/news/internal/data/app"
	"github.com/shipa988/hw_otus_architect/internal/data/controller/log"
	"github.com/spf13/cobra"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/shipa988/hw_otus_architect/cmd/news/internal/data/config"
	"github.com/spf13/viper"
)

var cfgFile string
var logdest string
var loglevel string
var cfg *config.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "hw_otus_architect",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			if err := recover(); err != nil {
				if e, ok := err.(error); ok {
					log.Fatal(e)
				}

			}
		}()
		a := app.NewNewsService()
		if err := a.Start(cfg); err != nil {
			log.Fatal(err)
		}
	},
}

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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "config file (default is $HOME/config.json)")
	rootCmd.PersistentFlags().StringVar(&logdest, "logdest", "std", "set log destination for service:std,file,graylog")
	rootCmd.PersistentFlags().StringVar(&loglevel, "loglevel", "info", "set log level for service:debug,info,error,disable")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
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

		// Search config in home directory with name "config" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName("config.v")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	cfg = &config.Config{}
	err := viper.Unmarshal(cfg)
	if err != nil {
		log.Fatal(err)
	}
	switch logdest {
	case "std":
		log.InitWithStdout(loglevel, cfg.Name, cfg.Env)
	case "file":
		log.InitWithFile("clicker.log", loglevel, cfg.Name, cfg.Env)
	default:
		log.InitWithStdout(loglevel, cfg.Name, cfg.Env)
		log.Error(errors.New("unknown log destination. should be one of std,file,graylog. std as default"))
	}
}
