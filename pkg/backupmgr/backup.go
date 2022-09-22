package backupmgr

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

type backupContainer interface {
	Upload(srcPath string) error
	DownloadNewest(dstPath string) (string, error)
}

func createBackup(dbPath string) (string, error) {
	dstPath := dbPath + "-" + time.Now().Format("2006-01-02T15:04:05-0700")
	cmd := exec.Command("sh", "-c", fmt.Sprintf("sqlite3 %s '.backup %s'", dbPath, dstPath))
	err := cmd.Run()

	if err != nil {
		return "", err
	}

	return dstPath, nil
}

func BackupAndShip(dbPath string, container backupContainer) error {
	path, err := createBackup(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create backup at %v: %w", path, err)
	}

	log.Printf("created backup at %s", path)

	err = container.Upload(path)
	if err != nil {
		return fmt.Errorf("failed to upload backup at %v: %w", path, err)
	}

	return nil
}

func RestoreFromLatest(dbPath string, container backupContainer) error {
	backupName, err := container.DownloadNewest(dbPath)
	if err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	log.Printf("restored db from backup: %v", backupName)

	return nil
}
