package data

import (
	"github.com/citrix/adc-nitro-go/resource/config/system"
	"strings"
)

var cmdPolicyHaNodeGet = "(^show\\s+ha\\s+node\\s+0)"
var cmdPolicySystemBackupGet = "(^show\\s+system\\s+backup\\s+\\d{8}_\\d{6})"
var cmdPolicySystemBackupCreate = "(^create\\s+system\\s+backup\\s+\\d{8}_\\d{6})"
var cmdPolicySystemBackupDelete = "(^rm\\s+system\\s+backup\\s+\\d{8}_\\d{6}\\.tgz)"
var cmdPolicySystemFileDownload = "(^show\\s+system\\s+file\\s+\\d{8}_\\d{6}\\.tgz\\s+-fileLocation\\s+\"/var/ns_sys_backup\")"

func getSystemCmdPolicySpecification() string {
	cmdPolicies := []string{
		cmdPolicyHaNodeGet,
		cmdPolicySystemBackupGet,
		cmdPolicySystemBackupCreate,
		cmdPolicySystemBackupDelete,
		cmdPolicySystemFileDownload,
	}

	return strings.Join(cmdPolicies, "|")
}

func GetSystemCmdPolicyCreateData(name string) system.Systemcmdpolicy {
	return system.Systemcmdpolicy{
		Policyname: name,
		Action:     "ALLOW",
		Cmdspec:    getSystemCmdPolicySpecification(),
	}
}
