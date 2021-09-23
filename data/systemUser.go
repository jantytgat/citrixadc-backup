package data

import (
	"github.com/citrix/adc-nitro-go/resource/config/system"
)

func GetSystemUserCreateData(username string, password string) system.Systemuser {
	return system.Systemuser{
		Username:     username,
		Password:     password,
		Externalauth: "disabled",
		Timeout:      60,
	}
}


