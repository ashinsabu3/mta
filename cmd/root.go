/*
Copyright © 2022 Christian Hernandez christian@chernand.io

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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// set up the global config file
var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "mta",
	Version: "v0.0.1",
	Short:   "This commands turns Flux Kustomizations and HelmReleases into Argo CD Applications",
	Long: `This is a migration tool that helps you move your Flux Kustomizations and HelmReleases
into an Argo CD ApplicationSet or Application.

Kustomization example:

	mta kustomization --name=mykustomization --namespace=flux-system | kubectl apply -n argocd -f -

HelmRelease example:

	mta helmrelease --name=myhelmrelease --namespace=flux-system | kubectl apply -n argocd -f -

This utilty exports the named Kustomization or HelmRelease and the source Git repo or Helm repo and
creates a manifests to stdout, which you can pipe into an apply command
with kubectl.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mta.yaml)")

	rootCmd.PersistentFlags().String("kubeconfig", "", "Path to the kubeconfig file to use (if not the standard one).")
	rootCmd.PersistentFlags().String("name", "", "Name of Kustomization or HelmRelease to export")
	rootCmd.PersistentFlags().String("namespace", "flux-system", "Namespace of where the Kustomization or HelmRelease is")
	rootCmd.PersistentFlags().String("argocd-namespace", "argocd", "Namespace where Argo CD is installed")
	rootCmd.PersistentFlags().String("argoproject", "default", "Argo CD project to use for the migrated Application")
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
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".mta" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".mta")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
