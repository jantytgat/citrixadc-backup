package data

import "github.com/citrix/adc-nitro-go/resource/config/system"

func GetSystemCmdPolicyBindingCreateData(policy string, username string) system.Systemusercmdpolicybinding {
	return system.Systemusercmdpolicybinding{
		Policyname: policy,
		Priority:   100,
		Username:   username,
	}
}
