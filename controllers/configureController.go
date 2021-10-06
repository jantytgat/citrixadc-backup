package controllers

import (
	"fmt"
	"github.com/jantytgat/citrixadc-backup/models"
)

type ConfigureController struct {}

type ConfigureControllerCaller interface {
	Run(s models.BackupConfiguration)
}

func (c *ConfigureController) Run(s models.BackupConfiguration) {
	fmt.Println("Configure called")
	fmt.Println(s)
}