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
	// The name of the query, used for informative log outputs.
	name      string
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

func Query(q QueryBody, headers map[string]string) ([]byte, error) {
	body, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodPost, octopusBaseUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", "application/json")

	for k, v := range headers {
		request.Header.Add(k, v)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	log.Printf("Octopus query '%s' returned status %v", q.name, response.StatusCode)

	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return responseBytes, nil
}
