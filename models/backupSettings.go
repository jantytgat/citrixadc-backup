package models

type BackupSettings struct {
	OutputBasePath  string `yaml: outputbasepath`
	FolderPerTarget bool   `yaml: folderpertarget`
	Interval        int    `yaml: interval`
}
