# Citrix ADC Backup Utility

## Help output
```
   ___ _ _       _         _   ___   ___   ___          _             
  / __(_) |_ _ _(_)_ __   /_\ |   \ / __| | _ ) __ _ __| |___  _ _ __ 
 | (__| |  _| '_| \ \ /  / _ \| |) | (__  | _ \/ _` / _| / / || | '_ \
  \___|_|\__|_| |_/_\_\ /_/ \_\___/ \___| |___/\__,_\__|_\_\\_,_| .__/
                                                                |_|   
Citrix ADC Backup Utility

Usage:
  citrixadc-backup [command]

Available Commands:
  backup      Backup all targets defined in the configuration file
  completion  generate the autocompletion script for the specified shell
  configure   Create a configuration file for citrixadc-backup
  help        Help about any command
  install     Install all targets defined in the configuration file
  uninstall   Uninstall all targets defined in the configuration file

Flags:
      --config string   config file (default is $HOME/.citrixadc-backup.yaml)
  -h, --help            help for citrixadc-backup

Use "citrixadc-backup [command] --help" for more information about a command.

```

## Default configuration file
```
Targets:
  - Name: HighAvailableTarget
    Type: hapair
    Username: nsroot
    Password: nsroot
    Level: full
    ValidateCertificate: false
    Nodes:
      - Name: dummy-vpx-001
        Address: http://169.254.254.254
      - Name: dummy-vpx-002
        Address: https://dummy-vpx-002.domain.local
  - Name: StandaloneTarget
    Type: standalone
    Username: nsroot
    Password: nsroot
    ValidateCertificate: false
    Nodes:
      - Name: dummy-vpx-001
        Address: http://dummy-vpx-001
Settings:
  OutputBasePath: /var/citrixadc/backup
  FolderPerTarget: true

```

## Usage
### No configuration file
Create a default configuration file when you have no file available

Run either
```citrixadc-backup```
or
```citrixadc-backup --config config.yaml```

The first option will create a default file .citrixadc-backup.yaml, either in the current directory or in the user home folder.
The second option will create a file config.yaml in the working directory.

### Edit the configuration file
Targets --> Citrix ADC Setup

For each target, you define the necessary settings:
- Target name --> e.g. <customername>-<production>
- Type: standalone | hapair
- Username: username to be used for backup
- Password: username to be used for backup
- Level: basic | full
- ValidateCertificate: true | false

For each node, specify the name of the node and the URL:
- http://fqdn or https://fqdn
- http://ipaddress or https://ipaddress

Also specify the necessary settings:
- OutputBasePath: where to store backups
- FolderPerTarget: true | false


!!! **Note: settings are not taken into account yet** !!!

### Install
Create the user and necessary command policy on the ADC.

Run either
```citrixadc-backup install```

or

```citrixadc-backup install --config config.yaml```


For each target, you will be asked for admin credentials with the necessary permissions to perform the installation actions.

For example, you can use the nsroot account

### Backup
To start creating backups, issue one of the following commands:

```citrixadc-backup backup```

or

```citrixadc-backup backup --config config.yaml```


### Uninstall
Remove the user and command policy from the ADC.

Run either
```citrixadc-backup uninstall```

or

```citrixadc-backup uninstall --config config.yaml```


For each target, you will be asked for admin credentials with the necessary permissions to perform the installation actions.

For example, you can use the nsroot account