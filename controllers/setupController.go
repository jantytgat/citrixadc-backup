package controllers

import (
	"bufio"
	"fmt"
	"github.com/citrix/adc-nitro-go/service"
	"github.com/jantytgat/citrixadc-backup/data"
	"github.com/jantytgat/citrixadc-backup/models"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"log"
	"os"
	"strings"
	"sync"
)

type SetupController struct{}

type SetupControllerCaller interface {
	ExecuteInstall()
	ExecuteUninstall()
}

func (s *SetupController) ExecuteInstall() {
	var wg sync.WaitGroup
	for _, t := range s.getSetupTargets() {
		wg.Add(1)
		go s.runInstallCommands(t, &wg)
	}
	wg.Wait()
}

func (s *SetupController) ExecuteUninstall() {
	var wg sync.WaitGroup
	for _, t := range s.getSetupTargets() {
		wg.Add(1)
		go s.runUninstallCommands(t, &wg)
	}
	wg.Wait()
}

func (s *SetupController) getSetupTargets() []models.SetupTarget {
	var config models.BackupConfiguration
	var setupTargets []models.SetupTarget

	err := viper.Unmarshal(&config)
	if err != nil {
		panic(err)
	}

	for _, t := range config.Targets {
		fmt.Printf("Configuring target: %s\n", t.Name)
		setupTarget := models.SetupTarget{
			Target:        t,
			Username:      s.getUsernameFromStdin(),
			Password:      s.getPasswordFromStdin(),
			CmdPolicyName: s.getCmdPolicyNameFromStdin(),
		}
		setupTargets = append(setupTargets, setupTarget)
	}
	return setupTargets
}

func (s *SetupController) getUsernameFromStdin() string {
	fmt.Print("Username: ")
	reader := bufio.NewReader(os.Stdin)
	username, _ := reader.ReadString('\n')
	// convert CRLF to LF
	username = strings.Replace(username, "\r\n", "", -1)
	username = strings.Replace(username, "\n", "", -1)
	return username
}

func (s *SetupController) getPasswordFromStdin() string {
	fmt.Print("Password: ")
	// https://pkg.go.dev/golang.org/x/term
	// terminal.ReadPassword accepts file descriptor as argument, returns byte slice and error.
	password, e := term.ReadPassword(int(os.Stdin.Fd()))
	if e != nil {
		log.Fatal(e)
	}
	fmt.Println()
	// Type cast byte slice to string.
	output := strings.Replace(string(password), "\r\n", "", -1)
	return output
}

func (s *SetupController) getCmdPolicyNameFromStdin() string {
	fmt.Print("Policy Name [leave empty for default value: CMD_CITRIXADCBACKUP]: ")
	reader := bufio.NewReader(os.Stdin)
	policyName, _ := reader.ReadString('\n')
	// convert CRLF to LF
	policyName = strings.Replace(policyName, "\r\n", "", -1)
	policyName = strings.Replace(policyName, "\n", "", -1)
	if policyName == "" {
		policyName = "CMD_CITRIXADCBACKUP"
	}
	return policyName
}

func (s *SetupController) createSetupNitroClientsForNodes(t models.SetupTarget) (map[string]service.NitroClient, error) {
	nitroClient := make(map[string]service.NitroClient, len(t.Target.Nodes))
	var err error
	for _, n := range t.Target.Nodes {
		client, err := service.NewNitroClientFromParams(
			service.NitroParams{
				Url:       n.Address,
				Username:  t.Username,
				Password:  t.Password,
				SslVerify: t.Target.ValidateCertificate,
			})
		if err != nil {
			log.Fatal("Could not create client for target ", t.Target.Name, " node ", n.Name)
		}

		nitroClient[n.Name] = *client
	}
	return nitroClient, err
}

func (s *SetupController) runInstallCommands(t models.SetupTarget, wg *sync.WaitGroup) {
	var sharedController = SharedController{}
	var primaryNode models.BackupNode
	var err error
	var clients = make(map[string]service.NitroClient, len(t.Target.Nodes))

	clients, err = s.createSetupNitroClientsForNodes(t)
	if err != nil {
		log.Fatal("Error creating nitro clients")
		wg.Done()
		return
	}

	primaryNode, err = sharedController.GetPrimaryNode(clients, t.Target)
	if err != nil {
		wg.Done()
		return
	}
	fmt.Println("Executing commands for", t.Target.Name, "on", primaryNode.Name)

	err = s.createCmdPolicy(clients[primaryNode.Name], t.CmdPolicyName)
	if err != nil {
		fmt.Println(err)
		wg.Done()
		return
	}

	err = s.createUser(clients[primaryNode.Name], t.Target.Username, t.Target.Password)
	if err != nil {
		wg.Done()
		return
	}

	err = s.bindCmdPolicy(clients[primaryNode.Name], t.Target.Username, t.CmdPolicyName)
	if err != nil {
		wg.Done()
		return
	}

	err = s.saveConfig(clients[primaryNode.Name])
	wg.Done()
}

func (s *SetupController) runUninstallCommands(t models.SetupTarget, wg *sync.WaitGroup) {
	var sharedController = SharedController{}
	var primaryNode models.BackupNode
	var err error
	var nitroClient = make(map[string]service.NitroClient, len(t.Target.Nodes))

	nitroClient, err = s.createSetupNitroClientsForNodes(t)
	if err != nil {
		log.Fatal("Error creating nitro clients")
		wg.Done()
		return
	}

	primaryNode, err = sharedController.GetPrimaryNode(nitroClient, t.Target)
	if err != nil {
		wg.Done()
		return
	}
	fmt.Println("Executing commands for", t.Target.Name, "on", primaryNode.Name)

	err = s.deleteUser(nitroClient[primaryNode.Name], t.Target.Username)
	if err != nil {
		wg.Done()
		return
	}

	err = s.deleteCmdPolicy(nitroClient[primaryNode.Name], t.CmdPolicyName)
	if err != nil {
		wg.Done()
		return
	}

	err = s.saveConfig(nitroClient[primaryNode.Name])
	wg.Done()
}

func (s *SetupController) createCmdPolicy(c service.NitroClient, name string) error {
	fmt.Println("Creating system command policy")
	request := data.GetSystemCmdPolicyCreateData(name)
	response, err := c.AddResource(service.Systemcmdpolicy.Type(), name, request)
	if err == nil {
		fmt.Println(response)
	}
	return err
}

func (s *SetupController) createUser(c service.NitroClient, username string, password string) error {
	fmt.Println("Creating system user")
	request := data.GetSystemUserCreateData(username, password)
	response, err := c.AddResource(service.Systemuser.Type(), username, request)
	if err == nil {
		fmt.Println(response)
	}
	return err
}

func (s *SetupController) bindCmdPolicy(c service.NitroClient, username string, policyName string) error {
	fmt.Println("Binding command policy to user")
	request := data.GetSystemCmdPolicyBindingCreateData(policyName, username)
	response, err := c.AddResource(service.Systemuser_binding.Type(), username, request)
	if err == nil {
		fmt.Println(response)
	}
	return err
}

func (s *SetupController) deleteUser(c service.NitroClient, username string) error {
	fmt.Println("Deleting system user")
	err := c.DeleteResource(service.Systemuser.Type(), username)
	return err
}

func (s *SetupController) deleteCmdPolicy(c service.NitroClient, policyName string) error {
	fmt.Println("Deleting system command policy")
	err := c.DeleteResource(service.Systemcmdpolicy.Type(), policyName)
	return err
}

func (s *SetupController) saveConfig(c service.NitroClient) error {
	sharedController := SharedController{}
	return sharedController.SaveConfig(c)
}
