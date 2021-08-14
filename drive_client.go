package main

import (
	"context"
	"fmt"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
)

const (
	PAGE_SIZE              = 100
	FOLDER_MIMETYPE_STRING = "application/vnd.google-apps.folder"
)

type DriveClient struct {
	*drive.Service
	Paths map[string]string
}

func NewDriveClient(s *drive.Service) *DriveClient {
	return &DriveClient{s, map[string]string{}}
}

func (dc *DriveClient) IsFolder(file *drive.File) bool {
	return file.MimeType == FOLDER_MIMETYPE_STRING
}

func (dc *DriveClient) IsDoc(file *drive.File) bool {
	return file.MimeType == FOLDER_MIMETYPE_STRING
}

func (dc *DriveClient) GetFile(ctx context.Context, fileID string, fields []string) (*drive.File, error) {
	fs := drive.NewFilesService(dc.Service)
	call := fs.Get(fileID).Context(ctx)
	if len(fields) > 0 {
		_fields := []googleapi.Field{}
		for _, v := range fields {
			_fields = append(_fields, googleapi.Field(v))
		}
		call = call.Fields(_fields...)
	}
	return call.Do()
}

func (dc *DriveClient) GetFileList(ctx context.Context, folder *drive.File, pageToken string, fields []string) (*drive.FileList, error) {
	fs := drive.NewFilesService(dc.Service)
	fs.List().Context(ctx)
	call := fs.List().Context(ctx).PageSize(PAGE_SIZE).Q(fmt.Sprintf("'%s' in parents", folder.Id))
	if len(fields) > 0 {
		_fields := []googleapi.Field{}
		for _, v := range fields {
			_fields = append(_fields, googleapi.Field(v))
		}
		call = call.Fields(_fields...)
	}
	return call.Do()
}

func (dc *DriveClient) GetParents(ctx context.Context, file *drive.File) (*drive.File, error) {
	// file, err := dc.GetFile(ctx, file.Id, []string{})
	// if err != nil {
	// 	return nil, err
	// }
	if len(file.Parents) < 1 {
		return nil, nil
	}
	return dc.GetFile(ctx, file.Parents[0], []string{})
}

func (dc *DriveClient) GetFilePath(ctx context.Context, file *drive.File) (string, error) {
	var err error
	cnt := 0
	path := file.Name
	parents, err := dc.GetParents(ctx, file)
	if err != nil {
		return "", err
	}
	for {
		path = parents.Name + "/" + path
		parents, err = dc.GetParents(ctx, parents)
		if err != nil {
			return "", err
		}
		if parents == nil {
			break
		}
		cnt += 1
		if cnt > 200 {
			break
		}
	}
	return "/" + path, nil
}
