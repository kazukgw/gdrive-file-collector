package main

import (
	"context"

	"google.golang.org/api/drive/v3"
)

type FileHandlers []FileHandler

type FileHandler func(*drive.File) error

type Walker struct {
	*DriveClient
	FileHandlers
}

func NewWalker(
	driveClient *DriveClient,
	handlers FileHandlers,
) *Walker {
	return &Walker{
		driveClient,
		handlers,
	}
}

func (w *Walker) Walk(ctx context.Context, folder *drive.File, pageToken string) (string, error) {
	fileList, err := w.DriveClient.GetFileList(ctx, folder, pageToken)
	if err != nil {
		return "", err
	}

	for _, f := range fileList.Files {
		if err := w.handleFile(ctx, f); err != nil {
			return "", err
		}
	}

	return fileList.NextPageToken, nil
}

func (w *Walker) handleFile(ctx context.Context, f *drive.File) error {
	for _, v := range w.FileHandlers {
		if err := v(f); err != nil {
			return err
		}
	}
	return nil
}
