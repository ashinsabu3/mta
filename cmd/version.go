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
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays version.",
	Long:  `This command will display the version of the CLI in json format`,
	Run: func(cmd *cobra.Command, args []string) {
		versionMap := map[string]string{rootCmd.Use: rootCmd.Version}
		versionJson, _ := json.Marshal(versionMap)
		fmt.Println(string(versionJson))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
