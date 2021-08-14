package main

import (
	"context"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/drive/v3"
)

type FileData struct {
	CollectedTime       time.Time
	CollectingRequestID string
	Path                string
	ParentsID           string
	ParentsName         string
	ID                  string
	Name                string
	MIMEType            string
	Data                string
}

type FileDataRepository struct {
	*DriveClient
	*bigquery.Table
	*bigquery.Inserter
}

func NewFileDataRepository(
	driveClient *DriveClient,
	table *bigquery.Table,
) *FileDataRepository {
	return &FileDataRepository{driveClient, table, table.Inserter()}
}

func (repo *FileDataRepository) SaveNewRecord(
	ctx context.Context,
	req CollectingRequest,
	file *drive.File,
	data string,
) error {
	path, err := repo.DriveClient.GetFilePath(ctx, file)
	if err != nil {
		return err
	}
	parents, err := repo.DriveClient.GetParents(ctx, file)
	if err != nil {
		return err
	}

	fd := FileData{
		time.Now(),
		req.ID,
		path,
		parents.Id,
		parents.Name,
		file.Id,
		file.Name,
		file.MimeType,
		data,
	}

	return repo.Inserter.Put(ctx, fd)
}
