package controllers

import (
	"fmt"
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

func (s *ScheduleController) Run() {
	fmt.Println("schedule called")

	fmt.Println(os.Getwd())
	fmt.Println(viper.AllKeys())
}

func (s *ScheduleController) addSchedule() {
	if runtime.GOOS == "windows" {
		s.addScheduleForWindows()
	} else {
		s.addScheduleForCron()
	}
}

func (s *ScheduleController) removeSchedule() {
	if runtime.GOOS == "windows" {
		s.addScheduleForWindows()
	} else {
		s.addScheduleForCron()
	}
}

func (s *ScheduleController) addScheduleForWindows() {}
func (s *ScheduleController) addScheduleForCron() {}

func (s *ScheduleController) removeScheduleForWindows() {}
func (s *ScheduleController) removeScheduleForCron() {}