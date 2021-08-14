package main

import (
	"context"

	"google.golang.org/api/drive/v3"
)

const (
	FOLDER_MIMETYPE_STRING = "application/vnd.google-apps.folder"
)

type DriveClient struct {
	*drive.Service
	Paths map[string]string
}

func NewDriveClient(s *drive.Service) *DriveClient {
	return &DriveClient{s, map[string]string{}}
}

func (du *DriveClient) IsFolder(file *drive.File) bool {
	return file.MimeType == FOLDER_MIMETYPE_STRING
}

func (du *DriveClient) IsDoc(file *drive.File) bool {
	return file.MimeType == FOLDER_MIMETYPE_STRING
}

func (du *DriveClient) GetFilePath(ctx context.Context, file *drive.File) (string, error) {
	return "", nil
}

func (du *DriveClient) GetParents(ctx context.Context, file *drive.File) (*drive.File, error) {
	return nil, nil
}

func (du *DriveClient) GetFileList(ctx context.Context, folder *drive.File, pageToken string) (*drive.FileList, error) {
	return nil, nil
}
