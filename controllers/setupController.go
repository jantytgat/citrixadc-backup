package controllers

import (
	"fmt"
	"github.com/citrix/adc-nitro-go/resource/config/ha"
	"github.com/citrix/adc-nitro-go/resource/config/system"
	"github.com/jantytgat/citrixadc-backup/models"
	"github.com/spf13/viper"
	"sync"
)

var temaplteHaNode = ha.Hanode{
	Id: 0,
}

var templateCmdPolicy = system.Systemcmdpolicy{
	Policyname: "",
	Action:     "ALLOW",
	Cmdspec:    "(^(show\\s+ha\\+node.*))|(^(show\\s+system\\s+backup)|^(create|rm)\\s+system\\s+backup\\s+.*)|(^show\\ssystem\\sfile\\s[\\w\\.-]+\\s-fileLocation\\s\"/var/ns_sys_backup\")",
}

var templateCmdUser = system.Systemuser{
	Username:                       "",
	Password:                       "",
	Externalauth:                   "false",
	Timeout:                        60,
}

var templateCmdUserBinding = system.Systemusercmdpolicybinding{
	Policyname: "",
	Priority:   100,
	Username:   "",
}

var config models.BackupConfiguration

func SetupTarget() {
	err := viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	for _, t := range config.Targets {
		wg.Add(1)
		go runCommands(t, &wg)
	}
	wg.Wait()
}

func runCommands(target models.BackupTarget, wg *sync.WaitGroup) {
	if target.Type == "HighAvailablePair" {
		fmt.Println("Detecting primary node for", target.Name)
		for _, node := range target.Nodes {
			go detectHaState(node, wg)
		}
	}
	fmt.Println("Executing commands for", target.Name)
	for _, node := range target.Nodes {
		createCmdPolicy(target.Name, target.Username, target.Password, target.UseSsl, target.ValidateCertificate, node)
		createUser(target.Name, target.Username, target.Password, target.UseSsl, target.ValidateCertificate, node)
		bindCmdPolicy(target.Name, target.Username, target.Password, target.UseSsl, target.ValidateCertificate, node)
	}
	wg.Done()
}

func detectHaState(node models.BackupNode, wg *sync.WaitGroup) {
	wg.Add(1)
	fmt.Println("Detecting node state for", node.Name)
	wg.Done()
}


func createCmdPolicy(target string, username string, password string, useSsl bool, validateCertificate bool, node models.BackupNode) {
	fmt.Println("Creating command policy for", target, node.Address)
	request := templateCmdPolicy
	request.Policyname = "CMD_BACKUP_" + target
	fmt.Println(request)
}

func createUser(target string, username string, password string, useSsl bool, validateCertificate bool, node models.BackupNode) {
	fmt.Println("Creating system user for", target, node.Name)
	request := templateCmdUser
	request.Username = username
	request.Password = password
	fmt.Println(request)
}

func bindCmdPolicy(target string, username string, password string, useSsl bool, validateCertificate bool, node models.BackupNode) {
	fmt.Println("Binding command policy to user for", target, node.Name)
	request := templateCmdUserBinding
	request.Policyname = "CMD_BACKUP_" + target
	request.Username = username
	fmt.Println(request)
}
