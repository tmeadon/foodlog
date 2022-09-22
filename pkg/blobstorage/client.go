package blobstorage

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type BackupContainer struct {
	container *azblob.ContainerClient
}

func NewBackupContainer(sasToken string) (*BackupContainer, error) {
	c, err := azblob.NewContainerClientWithNoCredential(sasToken, nil)
	if err != nil {
		return &BackupContainer{}, err
	}

	container := BackupContainer{container: c}
	return &container, nil
}

type cloudBlob struct {
	Name         string
	LastModified time.Time
}

func (c *BackupContainer) Upload(srcPath string) error {
	// check src exists
	if _, err := os.Stat(srcPath); err != nil {
		return err
	}

	// open the src file
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// create a blob client
	blobClient, err := c.container.NewBlockBlobClient(path.Base(srcPath))
	if err != nil {
		return err
	}

	// upload the file
	_, err = blobClient.UploadFile(context.Background(), file, azblob.UploadOption{})
	if err != nil {
		return err
	}

	log.Printf("Uploaded %s to %s\n", srcPath, strings.Split(c.container.URL(), "?")[0])
	return nil
}

func (c *BackupContainer) DownloadNewest(dstFilePath string) (string, error) {
	blobs := c.listAllBlobs()
	if len(blobs) == 0 {
		return "", fmt.Errorf("no backups found in blob container")
	}

	newestBlob := findNewestBlob(blobs)
	err := c.downloadBlob(newestBlob.Name, dstFilePath)
	return newestBlob.Name, err
}

func (c *BackupContainer) listAllBlobs() []cloudBlob {
	allBlobs := make([]cloudBlob, 0)

	pager := c.container.ListBlobsFlat(&azblob.ContainerListBlobsFlatOptions{})

	for pager.NextPage(context.Background()) {
		pageItems := pager.PageResponse().ListBlobsFlatSegmentResponse.Segment.BlobItems
		for _, p := range pageItems {
			allBlobs = append(allBlobs, cloudBlob{*p.Name, *p.Properties.LastModified})
		}
	}

	return allBlobs
}

func findNewestBlob(blobs []cloudBlob) cloudBlob {
	newestBlob := blobs[0]
	for _, b := range blobs {
		if b.LastModified.After(newestBlob.LastModified) {
			newestBlob = b
		}
	}
	return newestBlob
}

func (c *BackupContainer) downloadBlob(name string, dstFilePath string) error {
	// check if file already exists
	if _, err := os.Stat(dstFilePath); err == nil {
		return fmt.Errorf("file already exists at path %s", dstFilePath)
	}

	// create the file
	file, err := os.Create(dstFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// create a blob client and download the file
	blobClient, err := c.container.NewBlockBlobClient(name)
	if err != nil {
		return err
	}

	err = blobClient.DownloadToFile(context.Background(), 0, 0, file, azblob.DownloadOptions{})
	return err
}
