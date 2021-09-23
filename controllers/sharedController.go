package controllers

import (
	"fmt"
	"github.com/citrix/adc-nitro-go/service"
	"github.com/jantytgat/citrixadc-backup/models"
)

type SharedController struct {}

type SharedControllerLauncher interface {
	CheckNodeIsPrimary(c service.NitroClient) (bool, error)
	GetPrimaryNode(c map[string]service.NitroClient, t models.BackupTarget) (models.BackupNode, error)
	SaveConfig(c service.NitroClient) error
}

func (s *SharedController) CheckNodeIsPrimary(c service.NitroClient) (bool, error) {
	response, err := c.FindResource(service.Hanode.Type(), "0")
	if err == nil {
		if response["state"] == "Primary" {
			return true, err
		}
	}
	return false, err
}

func (s *SharedController) GetPrimaryNode(c map[string]service.NitroClient, t models.BackupTarget) (models.BackupNode, error) {
	var output models.BackupNode
	var err error
	if t.Type != "standalone" {
		fmt.Println("Detecting primary node for", t.Name)
		for _, n := range t.Nodes {
			// TODO - Check reverse err == nil
			if _, err := s.CheckNodeIsPrimary(c[n.Name]); err == nil {
				output = n
				break
			}
		}
	} else {
		output = t.Nodes[0]
	}
	return output, err
}

func (s *SharedController) SaveConfig(c service.NitroClient) error {
	err := c.SaveConfig()
	return err
}