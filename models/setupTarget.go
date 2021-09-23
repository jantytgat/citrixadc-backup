package models

type SetupTarget struct {
	Target BackupTarget
	Username string
	Password string

	CmdPolicyName string
}
