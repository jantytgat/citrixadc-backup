package controllers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/citrix/adc-nitro-go/service"
	"github.com/jantytgat/citrixadc-backup/data"
	"github.com/jantytgat/citrixadc-backup/models"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type BackupController struct{}
type BackupControllerLauncher interface {
	ExecuteRun()
}

func (b *BackupController) ExecuteBackup() {
	c, err := b.getBackupConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	err = b.createDirectory(c.Settings.OutputBasePath)
	if err != nil {
		log.Fatal("Access denied to ", c.Settings.OutputBasePath)
	}

	var wg sync.WaitGroup
	for _, t := range c.Targets {
		wg.Add(1)
		go b.runBackupCommands(t, c.Settings, &wg)
	}
	wg.Wait()
}

func (b *BackupController) getBackupConfiguration() (models.BackupConfiguration, error) {
	var config models.BackupConfiguration
	err := viper.Unmarshal(&config)

	return config, err
}

func (b *BackupController) runBackupCommands(t models.BackupTarget, s models.BackupSettings, wg *sync.WaitGroup) {
	var sharedController = SharedController{}
	var primaryNode models.BackupNode
	var err error
	var nitroClient = make(map[string]service.NitroClient, len(t.Nodes))

	nitroClient, err = b.createNitroClientsForNodes(t)
	if err != nil {
		log.Fatal("Error creating nitro clients")
		wg.Done()
		return
	}

	primaryNode, err = sharedController.GetPrimaryNode(nitroClient, t)
	if err != nil {
		wg.Done()
		return
	}

	timestamp := b.getTimestamp()
	err = b.createSystemBackup(nitroClient[primaryNode.Name], timestamp, t.Level)
	if err != nil {
		wg.Done()
		return
	}

	for _, n := range t.Nodes {
		var f string
		f, err = b.downloadSystemBackup(nitroClient[n.Name], timestamp+".tgz")
		if err != nil {
			fmt.Println(err)
			wg.Done()
			return
		}

		err = b.writeFileToDisk(b.generateFilename(timestamp, t.Name, n.Name), t.Name, f, s)
		if err != nil {
			fmt.Println(err)
			wg.Done()
			return
		}

		err = b.deleteSystemBackup(nitroClient[n.Name], timestamp+".tgz")
		if err != nil {
			wg.Done()
			return
		}
	}

	wg.Done()
}

func (b *BackupController) createNitroClientsForNodes(t models.BackupTarget) (map[string]service.NitroClient, error) {
	nitroClient := make(map[string]service.NitroClient, len(t.Nodes))
	var err error
	for _, n := range t.Nodes {
		client, err := service.NewNitroClientFromParams(
			service.NitroParams{
				Url:       n.Address,
				Username:  t.Username,
				Password:  t.Password,
				SslVerify: t.ValidateCertificate,
			})
		if err != nil {
			log.Fatal("Could not create client for target", t.Name, "node", n.Name)
		}
		nitroClient[n.Name] = *client
	}
	return nitroClient, err
}

func (b *BackupController) getTimestamp() string {
	t := time.Now()
	return fmt.Sprintf("%d%02d%02d_%02d%02d%02d",
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
	)
}

func (b *BackupController) createSystemBackup(c service.NitroClient, name string, level string) error {
	// Filename must have no extension
	name = strings.TrimSuffix(name, ".tgz")
	request := data.GetSystemBackupCreateData(name, level)

	err := c.ActOnResource(service.Systembackup.Type(), request, "create")
	return err
}

func (b *BackupController) downloadSystemBackup(c service.NitroClient, name string) (string, error) {
	var output string
	params := service.FindParams{
		ArgsMap:                  map[string]string{"fileLocation": url.PathEscape("/var/ns_sys_backup")},
		ResourceType:             "systemfile",
		ResourceName:             name,
		ResourceMissingErrorCode: 0,
	}

	response, err := c.FindResourceArrayWithParams(params)
	if err == nil {
		if response[0]["filecontent"] != "" {
			output = response[0]["filecontent"].(string)
		} else {
			output = ""
		}
	}
	return output, err
}

func (b *BackupController) createDirectory(path string) error {
	src, err := os.Stat(path)

	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	} else if src.Mode().IsRegular() {
		return os.ErrExist
	} else {
		return nil
	}
}

func (b *BackupController) writeFileToDisk(filename string, targetName string, data string, settings models.BackupSettings) error {
	var outputFile string

	if settings.FolderPerTarget {
		err := b.createDirectory(filepath.Join(settings.OutputBasePath, targetName))
		if err != nil {
			return err
		}
		outputFile = filepath.Join(settings.OutputBasePath, targetName, filename)
	} else {
		outputFile = filepath.Join(settings.OutputBasePath, filename)
	}

	fmt.Println("Writing to file", outputFile)
	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(data))
	buffer := bytes.Buffer{}

	_, err := buffer.ReadFrom(reader)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(outputFile, buffer.Bytes(), 0644)
	return err
}

func (b *BackupController) generateFilename(timestamp string, target string, node string) string {
	var output []string

	output = append(output, timestamp)
	output = append(output, target)
	output = append(output, node+".tgz")

	return strings.Join(output, "_")
}

func (b *BackupController) deleteSystemBackup(c service.NitroClient, name string) error {
	err := c.DeleteResource(service.Systembackup.Type(), name)
	return err
}
