package models

type BackupTarget struct {
	Name                string       `yaml: name`
	Type                string       `yaml: type`
	Level               string       `yaml: level`
	Nodes               []BackupNode `yaml: nodes`
	UseSsl              bool         `yaml: usessl`
	ValidateCertificate bool         `yaml: validatecertificate`
	Username            string       `yaml: username`
	Password            string       `yaml: password`
}
