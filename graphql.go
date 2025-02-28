package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

const octopusBaseUrl = "https://api.octopus.energy/v1/graphql/"

var httpClient = &http.Client{}

type QueryBody struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

func Query(q QueryBody) ([]byte, error) {
	body, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}

	log.Println(string(body))

	request, err := http.NewRequest(http.MethodPost, octopusBaseUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json")

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	log.Printf("Status %v", response.StatusCode)

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return responseBytes, nil
}
