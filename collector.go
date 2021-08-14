package main

import (
	"context"
	"strings"

	"cloud.google.com/go/bigquery"
	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"google.golang.org/api/drive/v3"
)

type Collector struct {
	*DriveClient
	*CollectingRequestQueueClient
	*FileDataRepository
}

type CollectorConfig struct {
	Project            string
	FileDataRepository struct {
		Dataset string
		Table   string
	}
	CollectingRequestQueue struct {
		TaskQueuePath     string
		HTTPTaskTargetURL string
	}
}

func NewCollector(ctx context.Context, config CollectorConfig) (*Collector, error) {
	s, err := drive.NewService(ctx)
	if err != nil {
		return nil, err
	}
	driveClient := NewDriveClient(s)

	tasksClient, err := cloudtasks.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	queueCli := NewCollectingRequestQueueClient(
		tasksClient, config.CollectingRequestQueue.TaskQueuePath, config.CollectingRequestQueue.HTTPTaskTargetURL)

	bq, err := bigquery.NewClient(ctx, config.Project)
	if err != nil {
		return nil, err
	}
	table := bq.Dataset(config.FileDataRepository.Dataset).Table(config.FileDataRepository.Table)
	repo := NewFileDataRepository(driveClient, table)

	return &Collector{driveClient, queueCli, repo}, nil
}

func (c *Collector) Collect(ctx context.Context, req CollectingRequest) error {
	var err error
	fileList, err := c.DriveClient.GetFileList(ctx, req.Folder, req.PageToken, []string{
		"nextPageToken",
		"incompleteSearch",
		"files(" + getFileFileds() + ")",
	})

	if err != nil {
		return err
	}

	for _, f := range fileList.Files {
		if c.DriveClient.IsFolder(f) {
			_, err = c.CollectingRequestQueueClient.Push(
				ctx, NewCollectingRequest(req.RootCollectingRequestID, f, ""))
		}
		if err != nil {
			return err
		}

		data, err := ExtractFileData(ctx, c.DriveClient, f)
		if err != nil {
			return err
		}

		err = c.FileDataRepository.SaveNewRecord(ctx, req, f, data)
		if err != nil {
			return err
		}
	}

	if fileList.NextPageToken != "" {
		_, err = c.CollectingRequestQueueClient.Push(
			ctx, NewCollectingRequest(req.RootCollectingRequestID, req.Folder, fileList.NextPageToken))
	}

	return err
}

func getFileFileds() string {
	return strings.Join([]string{
		"id",
		"name",
		"mimeType",
		"description",
		"parents",
		"properties",
		"version",
		"thumbnailLink",
		"createdTime",
		"modifiedTime",
		"lastModifyingUser",
		"fullFileExtension",
		"size",
	}, ", ")
}
