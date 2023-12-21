package utils

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func MoveFileToDirectory(fullFilePathToMove string, newDirectory string) error {
	// Get the full path of the csvFile
	fullPath, err := filepath.Abs(fullFilePathToMove)
	if err != nil {
		return err
	}

	// Get the full path of the new directory
	newDirFullPath, err := filepath.Abs(newDirectory)
	if err != nil {
		return err
	}

	// Create the new directory if it doesn't exist
	if _, err := os.Stat(newDirFullPath); os.IsNotExist(err) {
		if err := os.MkdirAll(newDirFullPath, 0755); err != nil {
			return err
		}
	}

	// Move the file to the new directory
	newFilePath := filepath.Join(newDirFullPath, filepath.Base(fullPath))
	if err := os.Rename(fullPath, newFilePath); err != nil {
		return err
	}

	log.Infof("Moved %s to %s", fullPath, newFilePath)
	return nil
}
