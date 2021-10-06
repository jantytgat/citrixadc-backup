package controllers

import (
	"bufio"
	"fmt"
	"github.com/citrix/adc-nitro-go/service"
	"github.com/jantytgat/citrixadc-backup/data"
	"github.com/jantytgat/citrixadc-backup/models"
	"golang.org/x/term"
	"log"
	"os"
	"strings"
	"sync"
)

type SetupController struct{}

type SetupControllerCaller interface {
	RunInstall(s models.BackupConfiguration)
	RunUninstall(s models.BackupConfiguration)

	getSetupTargets() []models.SetupTarget
	getUsernameFromStdin() string
	getPasswordFromStdin() string
	getCmdPolicyNameFromStdin() string

	createSetupNitroClientsForNodes(t models.SetupTarget) (map[string]service.NitroClient, error)
	runInstallCommands(t models.SetupTarget, wg *sync.WaitGroup)
	runUninstallCommands(t models.SetupTarget, wg *sync.WaitGroup)
	createCmdPolicy(nitroClient service.NitroClient, name string) error
	createUser(nitroClient service.NitroClient, username string, password string) error
	bindCmdPolicy(nitroClient service.NitroClient, username string, policyName string) error
	deleteUser(nitroClient service.NitroClient, username string) error
	deleteCmdPolicy(nitroClient service.NitroClient, policyName string) error
	saveConfig(nitroClient service.NitroClient) error
}

func (c *SetupController) RunInstall(s models.BackupConfiguration) {
	var wg sync.WaitGroup
	for _, t := range c.getSetupTargets(s) {
		wg.Add(1)
		go c.runInstallCommands(t, &wg)
	}
	wg.Wait()
}

func (c *SetupController) RunUninstall(s models.BackupConfiguration) {
	var wg sync.WaitGroup
	for _, t := range c.getSetupTargets(s) {
		wg.Add(1)
		go c.runUninstallCommands(t, &wg)
	}
	wg.Wait()
}

func (c *SetupController) getSetupTargets(s models.BackupConfiguration) []models.SetupTarget {
	var setupTargets []models.SetupTarget

	for _, t := range s.Targets {
		fmt.Printf("Configuring target: %s\n", t.Name)
		setupTarget := models.SetupTarget{
			Target:        t,
			Username:      c.getUsernameFromStdin(),
			Password:      c.getPasswordFromStdin(),
			CmdPolicyName: c.getCmdPolicyNameFromStdin(),
		}
		setupTargets = append(setupTargets, setupTarget)
	}
	return setupTargets
}

func (c *SetupController) getUsernameFromStdin() string {
	fmt.Print("Username: ")
	reader := bufio.NewReader(os.Stdin)
	username, _ := reader.ReadString('\n')
	// convert CRLF to LF
	username = strings.Replace(username, "\r\n", "", -1)
	username = strings.Replace(username, "\n", "", -1)
	return username
}

func (c *SetupController) getPasswordFromStdin() string {
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

func (c *SetupController) getCmdPolicyNameFromStdin() string {
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

func (c *SetupController) createSetupNitroClientsForNodes(t models.SetupTarget) (map[string]service.NitroClient, error) {
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

func (c *SetupController) runInstallCommands(t models.SetupTarget, wg *sync.WaitGroup) {
	var primaryNode models.BackupNode
	var err error
	var clients = make(map[string]service.NitroClient, len(t.Target.Nodes))

	clients, err = c.createSetupNitroClientsForNodes(t)
	if err != nil {
		log.Fatal("Error creating nitro clients")
		wg.Done()
		return
	}

	primaryNode, err = getPrimaryNode(clients, t.Target)
	if err != nil {
		wg.Done()
		return
	}
	fmt.Println("Executing commands for", t.Target.Name, "on", primaryNode.Name)

	err = c.createCmdPolicy(clients[primaryNode.Name], t.CmdPolicyName)
	if err != nil {
		fmt.Println(err)
		wg.Done()
		return
	}

	err = c.createUser(clients[primaryNode.Name], t.Target.Username, t.Target.Password)
	if err != nil {
		wg.Done()
		return
	}

	err = c.bindCmdPolicy(clients[primaryNode.Name], t.Target.Username, t.CmdPolicyName)
	if err != nil {
		wg.Done()
		return
	}

	err = c.saveConfig(clients[primaryNode.Name])
	wg.Done()
}

func (c *SetupController) runUninstallCommands(t models.SetupTarget, wg *sync.WaitGroup) {
	var primaryNode models.BackupNode
	var err error
	var nitroClient = make(map[string]service.NitroClient, len(t.Target.Nodes))

	nitroClient, err = c.createSetupNitroClientsForNodes(t)
	if err != nil {
		log.Fatal("Error creating nitro clients")
		wg.Done()
		return
	}

	primaryNode, err = getPrimaryNode(nitroClient, t.Target)
	if err != nil {
		wg.Done()
		return
	}
	fmt.Println("Executing commands for", t.Target.Name, "on", primaryNode.Name)

	err = c.deleteUser(nitroClient[primaryNode.Name], t.Target.Username)
	if err != nil {
		wg.Done()
		return
	}

	err = c.deleteCmdPolicy(nitroClient[primaryNode.Name], t.CmdPolicyName)
	if err != nil {
		wg.Done()
		return
	}

	err = c.saveConfig(nitroClient[primaryNode.Name])
	wg.Done()
}

func (c *SetupController) createCmdPolicy(nitroClient service.NitroClient, name string) error {
	fmt.Println("Creating system command policy")
	request := data.GetSystemCmdPolicyCreateData(name)
	response, err := nitroClient.AddResource(service.Systemcmdpolicy.Type(), name, request)
	if err == nil {
		fmt.Println(response)
	}
	return err
}

func (c *SetupController) createUser(nitroClient service.NitroClient, username string, password string) error {
	fmt.Println("Creating system user")
	request := data.GetSystemUserCreateData(username, password)
	response, err := nitroClient.AddResource(service.Systemuser.Type(), username, request)
	if err == nil {
		fmt.Println(response)
	}
	return err
}

func (c *SetupController) bindCmdPolicy(nitroClient service.NitroClient, username string, policyName string) error {
	fmt.Println("Binding command policy to user")
	request := data.GetSystemCmdPolicyBindingCreateData(policyName, username)
	response, err := nitroClient.AddResource(service.Systemuser_binding.Type(), username, request)
	if err == nil {
		fmt.Println(response)
	}
	return err
}

func (c *SetupController) deleteUser(nitroClient service.NitroClient, username string) error {
	fmt.Println("Deleting system user")
	err := nitroClient.DeleteResource(service.Systemuser.Type(), username)
	return err
}

func (c *SetupController) deleteCmdPolicy(nitroClient service.NitroClient, policyName string) error {
	fmt.Println("Deleting system command policy")
	err := nitroClient.DeleteResource(service.Systemcmdpolicy.Type(), policyName)
	return err
}

func (c *SetupController) saveConfig(nitroClient service.NitroClient) error {
	return saveConfig(nitroClient)
}
