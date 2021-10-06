package controllers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/citrix/adc-nitro-go/service"
	"github.com/jantytgat/citrixadc-backup/data"
	"github.com/jantytgat/citrixadc-backup/models"
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
	Run(s models.BackupConfiguration)
	runBackupCommands(t models.BackupTarget, s models.BackupSettings, wg *sync.WaitGroup)
	createSystemBackup(nitroClient service.NitroClient, name string, level string) error
	downloadSystemBackup(nitroClient service.NitroClient, name string) (string, error)
	deleteSystemBackup(nitroClient service.NitroClient, name string) error
	getTimestamp() string
	generateFilename(timestamp string, target string, node string) string
	createDirectory(path string) error
	writeFileToDisk(filename string, targetName string, data string, settings models.BackupSettings) error
}

func (c *BackupController) Run(s models.BackupConfiguration) {
	err := c.createDirectory(s.Settings.OutputBasePath)
	if err != nil {
		log.Fatal("Access denied to ", s.Settings.OutputBasePath)
	}

	var wg sync.WaitGroup
	for _, t := range s.Targets {
		wg.Add(1)
		go c.runBackupCommands(t, s.Settings, &wg)
	}
	wg.Wait()
}

func (c *BackupController) runBackupCommands(t models.BackupTarget, s models.BackupSettings, wg *sync.WaitGroup) {
	var primaryNode models.BackupNode
	var err error
	var nitroClient = make(map[string]service.NitroClient, len(t.Nodes))

	nitroClient, err = createNitroClientsForNodes(t)
	if err != nil {
		log.Fatal("Error creating nitro clients")
		wg.Done()
		return
	}

	primaryNode, err = getPrimaryNode(nitroClient, t)
	if err != nil {
		wg.Done()
		return
	}

	timestamp := c.getTimestamp()
	err = c.createSystemBackup(nitroClient[primaryNode.Name], timestamp, t.Level)
	if err != nil {
		wg.Done()
		return
	}

	for _, n := range t.Nodes {
		var f string
		f, err = c.downloadSystemBackup(nitroClient[n.Name], timestamp+".tgz")
		if err != nil {
			fmt.Println(err)
			wg.Done()
			return
		}

		err = c.writeFileToDisk(c.generateFilename(timestamp, t.Name, n.Name), t.Name, f, s)
		if err != nil {
			fmt.Println(err)
			wg.Done()
			return
		}

		err = c.deleteSystemBackup(nitroClient[n.Name], timestamp+".tgz")
		if err != nil {
			wg.Done()
			return
		}
	}

	wg.Done()
}


func (c *BackupController) createSystemBackup(nitroClient service.NitroClient, name string, level string) error {
	// Filename must have no extension
	name = strings.TrimSuffix(name, ".tgz")
	request := data.GetSystemBackupCreateData(name, level)

	err := nitroClient.ActOnResource(service.Systembackup.Type(), request, "create")
	return err
}

func (c *BackupController) downloadSystemBackup(nitroClient service.NitroClient, name string) (string, error) {
	var output string
	params := service.FindParams{
		ArgsMap:                  map[string]string{"fileLocation": url.PathEscape("/var/ns_sys_backup")},
		ResourceType:             "systemfile",
		ResourceName:             name,
		ResourceMissingErrorCode: 0,
	}

	response, err := nitroClient.FindResourceArrayWithParams(params)
	if err == nil {
		if response[0]["filecontent"] != "" {
			output = response[0]["filecontent"].(string)
		} else {
			output = ""
		}
	}
	return output, err
}

func (c *BackupController) deleteSystemBackup(nitroClient service.NitroClient, name string) error {
	err := nitroClient.DeleteResource(service.Systembackup.Type(), name)
	return err
}

func (c *BackupController) getTimestamp() string {
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

func (c *BackupController) generateFilename(timestamp string, target string, node string) string {
	var output []string

	output = append(output, timestamp)
	output = append(output, target)
	output = append(output, node+".tgz")

	return strings.Join(output, "_")
}

func (c *BackupController) createDirectory(path string) error {
	src, err := os.Stat(path)

	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	} else if src.Mode().IsRegular() {
		return os.ErrExist
	} else {
		return nil
	}
}

func (c *BackupController) writeFileToDisk(filename string, targetName string, data string, settings models.BackupSettings) error {
	var outputFile string

	if settings.FolderPerTarget {
		err := c.createDirectory(filepath.Join(settings.OutputBasePath, targetName))
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


