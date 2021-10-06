package controllers

import (
	"fmt"
	"github.com/jantytgat/citrixadc-backup/models"
	"github.com/spf13/viper"
	"os"
	"runtime"
)

type ScheduleController struct{}

type ScheduleControllerCaller interface {
	Run()

	addSchedule()
	removeSchedule()

	addScheduleForWindows()
	addScheduleForCron()

	removeScheduleForWindows()
	removeScheduleForCron()
}

func (c *ScheduleController) Run(s models.BackupConfiguration, configFile string) {
	fmt.Println("schedule called")

	fmt.Println(os.Getwd())
	fmt.Println(configFile)
	fmt.Println(viper.AllKeys())
	fmt.Println(s)
}

func (c *ScheduleController) addSchedule() {
	if runtime.GOOS == "windows" {
		c.addScheduleForWindows()
	} else {
		c.addScheduleForCron()
	}
}

func (c *ScheduleController) removeSchedule() {
	if runtime.GOOS == "windows" {
		c.addScheduleForWindows()
	} else {
		c.addScheduleForCron()
	}
}

func (c *ScheduleController) addScheduleForWindows() {}
func (c *ScheduleController) addScheduleForCron() {}

func (c *ScheduleController) removeScheduleForWindows() {}
func (c *ScheduleController) removeScheduleForCron() {}