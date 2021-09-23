package models

type BackupConfiguration struct {
	Targets  []BackupTarget `yaml: targets`
	Settings BackupSettings `yaml: settings`
}
