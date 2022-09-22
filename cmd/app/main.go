package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tmeadon/foodlog/pkg/backupmgr"
	"github.com/tmeadon/foodlog/pkg/blobstorage"
	"github.com/tmeadon/foodlog/pkg/webapp"
)

var dbPath string = "db/sqlite/foodlog.db"
var backupContainer *blobstorage.BackupContainer

func main() {
	nobackups := flag.Bool("nobackups", false, "disables backups")
	flag.Parse()

	done := make(chan bool)

	if !*nobackups {
		getBackupContainer()
		restoreDbIfNeeded()
		sigs := registerSigListeners()
		go backupRoutine(sigs, done)
	}

	go func() {
		s := webapp.NewServer(dbPath, []byte("secret"))
		s.Start()
	}()

	<-done
	log.Printf("exiting")
}

func restoreDbIfNeeded() {
	if restoreNeeded() {
		log.Print("database file missing, restoring from backup")
		err := backupmgr.RestoreFromLatest(dbPath, backupContainer)

		if err != nil {
			log.Fatalf("failed to restore backup: %v", err)
		}
	}
}

func restoreNeeded() bool {
	if _, err := os.Stat(dbPath); err == nil {
		return false
	}

	return true
}

func registerSigListeners() chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	return ch
}

func getBackupContainer() {
	sas, ok := os.LookupEnv("FOODLOG_BACKUP_SAS")
	if !ok {
		log.Fatal("FOODLOG_BACKUP_SAS environment variable not set")
	}

	container, err := blobstorage.NewBackupContainer(sas)
	if err != nil {
		log.Fatalf("unable to connect to backup container: %v", err)
	}

	backupContainer = container
}

func backupRoutine(sig chan os.Signal, done chan bool) {
	for {
		select {
		case <-time.After(8 * time.Hour):
			log.Printf("backup timer elapsed")
			backup()
		case <-sig:
			log.Printf("backup signal received")
			backup()
			done <- true
		}
	}
}

func backup() {
	err := backupmgr.BackupAndShip(dbPath, backupContainer)

	if err != nil {
		log.Printf("failed to execute database backup: %v", err)
	} else {
		log.Printf("backup completed")
	}
}
