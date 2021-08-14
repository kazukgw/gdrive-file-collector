package main

import (
	"context"
	"time"
	"fmt"

	cloudtasks "cloud.google.com/go/cloudtasks/apiv2"
	"github.com/google/uuid"
	"google.golang.org/api/drive/v3"
	taskspb "google.golang.org/genproto/googleapis/cloud/tasks/v2"
)

type CollectingRequest struct {
	ID                      string
	RootCollectingRequestID string
	Folder                  *drive.File
	PageToken               string
	CreatedTime             time.Time
}

func (req *CollectingRequest) UnmarshalJSON(b []byte) error {
	return nil
}

func (req CollectingRequest) MarshalJSON() ([]byte, error) {
	return []byte{}, nil
}

func NewCollectingRequest(
	rootReqID string,
	folder *drive.File,
	nextPageToken string,
) CollectingRequest {
	id := uuid.NewString()
	rootReqID_ := rootReqID
	if rootReqID_ == "" {
		rootReqID_ = id
	}
	return CollectingRequest{
		id,
		rootReqID_,
		folder,
		nextPageToken,
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
	taskreq := &taskspb.CreateTaskRequest{
		Parent: cli.QueuePath,
		Task: &taskspb.Task{
			// https://godoc.org/google.golang.org/genproto/googleapis/cloud/tasks/v2#HttpRequest
			MessageType: &taskspb.Task_HttpRequest{
				HttpRequest: &taskspb.HttpRequest{
					HttpMethod: taskspb.HttpMethod_POST,
					Url:        cli.HTTPTargetURL,
					Body: json.Marshal(req),
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
