/*
Copyright © 2021 Jan Tytgat

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
	"bufio"
	"bytes"
	"fmt"
	"github.com/jantytgat/citrixadc-backup/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var configFile string
var yamlExample = []byte(`
Targets:
  - Name: HighAvailableTarget
    Type: hapair
    Username: nsroot
    Password: nsroot
    Level: full
    ValidateCertificate: false
    Nodes:
      - name: dummy-vpx-001
        address: http://169.254.254.254
      - name: dummy-vpx-002
        address: https://dummy-vpx-002.domain.local
  - Name: StandaloneTarget
    Type: standalone
    Username: nsroot
    Password: nsroot
    ValidateCertificate: false
    Nodes:
      - name: dummy-vpx-001
        address: http://dummy-vpx-001
Settings:
  OutputBasePath: /var/citrixadc/backup
  FolderPerTarget: true
`)

var yamlTemplate = []byte(`
Targets:
Settings:
  OutputBasePath:
  FolderPerTarget:
  Interval: 6
`)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "citrixadc-backup",
	Short: "Citrix ADC Backup Utility",
	Long:  `Citrix ADC Backup Utility`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//Run: func(cmd *cobra.Command, args []string) {	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "citrixadc-backup.yaml", "config file (default is $PWD/citrixadc-backup.yaml)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	verifyLoading()
}

func verifyLoading() {
	if err := viper.ReadInConfig(); err != nil {
		// Config file was found but another error was produced
		fmt.Printf("Could not find %s, generate empty configuration at specified location? [y/n]: ", configFile)

		reader := bufio.NewReader(os.Stdin)
		generateFile, _ := reader.ReadString('\n')
		// convert CRLF to LF
		generateFile = strings.Replace(generateFile, "\r\n", "", -1)
		generateFile = strings.Replace(generateFile, "\n", "", -1)

		if generateFile == "y" {
			viper.ReadConfig(bytes.NewBuffer(yamlTemplate))
			viper.SafeWriteConfigAs(configFile)
		} else {
			os.Exit(0)
		}
	}
}

func getBackupConfiguration() (models.BackupConfiguration, error) {
	var config models.BackupConfiguration
	err := viper.Unmarshal(&config)

	return config, err
}

