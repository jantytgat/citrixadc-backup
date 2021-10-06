package controllers

import (
	"fmt"
	"github.com/citrix/adc-nitro-go/service"
	"github.com/jantytgat/citrixadc-backup/models"
	"log"
)

func createNitroClientsForNodes(t models.BackupTarget) (map[string]service.NitroClient, error) {
	nitroClient := make(map[string]service.NitroClient, len(t.Nodes))
	var err error
	for _, n := range t.Nodes {
		client, err := service.NewNitroClientFromParams(
			service.NitroParams{
				Url:       n.Address,
				Username:  t.Username,
				Password:  t.Password,
				SslVerify: t.ValidateCertificate,
			})
		if err != nil {
			log.Fatal("Could not create client for target", t.Name, "node", n.Name)
		}
		nitroClient[n.Name] = *client
	}
	return nitroClient, err
}


func checkNodeIsPrimary(nitroClient service.NitroClient) (bool, error) {
	response, err := nitroClient.FindResource(service.Hanode.Type(), "0")
	if err == nil {
		if response["state"] == "Primary" {
			return true, err
		}
	}
	return false, err
}

func getPrimaryNode(nitroClients map[string]service.NitroClient, t models.BackupTarget) (models.BackupNode, error) {
	var output models.BackupNode
	var err error
	if t.Type != "standalone" {
		fmt.Println("Detecting primary node for", t.Name)
		for _, n := range t.Nodes {
			// TODO - Check reverse err == nil
			if _, err := checkNodeIsPrimary(nitroClients[n.Name]); err == nil {
				output = n
				break
			}
		}
	} else {
		output = t.Nodes[0]
	}
	return output, err
}

func saveConfig(nitroClient service.NitroClient) error {
	err := nitroClient.SaveConfig()
	return err
}