package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"gopkg.in/yaml.v2"
)

func main() {
	flag.Parse()
	configFilePath := flag.Arg(0)

	dat, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("main: %w", err)
		return
	}

	ctx := context.Background()
	config := CollectorConfig{}
	yaml.Unmarshal(dat, &config)
	collector, err := NewCollector(ctx, config)
	if err != nil {
		log.Fatalf("main: %w", err)
		return
	}

	handler := NewHTTPRequestHandler(config, collector)

	http.HandleFunc("/", handler.CollectFiles)
	http.ListenAndServe(":8000", nil)
}

type HTTPRequestHandler struct {
	CollectorConfig
	*Collector
}

func NewHTTPRequestHandler(config CollectorConfig, collector *Collector) *HTTPRequestHandler {
	return &HTTPRequestHandler{config, collector}
}

func (h *HTTPRequestHandler) CollectFiles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := CollectingRequest{}
	dat, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error HTTPRequestHandler.CollectFiles: %v", err)
		return
	}
	json.Unmarshal(dat, &req)
	f, err := h.Collector.DriveClient.GetFile(ctx, req.FolderID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error HTTPRequestHandler.CollectFiles: %v", err)
		return
	}
	req.Folder = f

	if err = h.Collector.Collect(ctx, req); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Error HTTPRequestHandler.CollectFiles: %v", err)
		return
	}

	fmt.Fprintf(w, "Success")
}
