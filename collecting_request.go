package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"github.com/google/uuid"
	"google.golang.org/api/drive/v3"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
)

type CollectingRequest struct {
	ID                      string
	RootCollectingRequestID string
	FolderID                string
	Folder                  *drive.File `json:"-"`
	PageToken               string
	CreatedTime             time.Time
}

func NewCollectingRequest(
	rootReqID string,
	f *drive.File,
	pageToken string,
) CollectingRequest {
	return CollectingRequest{
		uuid.NewString(),
		rootReqID,
		f.Id,
		f,
		pageToken,
		time.Now(),
	}
}

type CollectingRequestQueueClient struct {
	CloudTasksClient *cloudtasks.Client
	QueuePath        string
	HTTPTargetURL    string
}

func NewCollectingRequestQueueClient(
	cloudTasksClient *cloudtasks.Client,
	queuePath string,
	httpTargetURL string,
) *CollectingRequestQueueClient {
	return &CollectingRequestQueueClient{
		cloudTasksClient,
		queuePath,
		httpTargetURL,
	}
}

func (cli *CollectingRequestQueueClient) Push(ctx context.Context, req CollectingRequest) (*taskspb.Task, error) {
	if req.FolderID == "" {
		req.FolderID = req.Folder.Id
	}
	jsondata, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	taskreq := &taskspb.CreateTaskRequest{
		Parent: cli.QueuePath,
		Task: &taskspb.Task{
			// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#HttpRequest
			MessageType: &taskspb.Task_HttpRequest{
				HttpRequest: &taskspb.HttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Url:        cli.HTTPTargetURL,
					Body:       jsondata,
				},
			},
		},
	}

	createdTask, err := cli.CloudTasksClient.CreateTask(ctx, taskreq)
	if err != nil {
		return nil, fmt.Errorf("CollectingRequestQueueClient.Push: %v", err)
	}

	return createdTask, nil
}
