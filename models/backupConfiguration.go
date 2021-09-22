package models

type BackupConfiguration struct {
	Targets         []BackupTarget `yaml: targets`
	OutputBasePath  string         `yaml: outputbasepath`
	FolderPerTarget bool           `yaml: folderpertarget`
	TimestampMode   string         `yaml: timestampmode`
}
