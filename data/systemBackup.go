package data

import "github.com/citrix/adc-nitro-go/resource/config/system"

func GetSystemBackupCreateData(name string, level string) system.Systembackup {
	return system.Systembackup{
		Filename:         name,
		Level:            level,
	}
}
