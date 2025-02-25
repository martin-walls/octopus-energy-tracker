package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

const octopusBaseUrl = "https://api.octopus.energy/v1/graphql/"

var httpClient = &http.Client{}

type Query struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

func query(q Query) ([]byte, error) {
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

type ObtainKrakenTokenResponse struct {
	Token            string `json:"token"`
	RefreshToken     string `json:"refreshToken"`
	RefreshExpiresIn int    `json:"refreshExpiresIn"`
}

func auth() error {
	apiKey := os.Getenv("OCTOPUS_API_KEY")

	q := Query{
		Query: `mutation ObtainKrakenToken($input: ObtainJSONWebTokenInput!) {
			obtainKrakenToken(input: $input) {
				token
				refreshToken
				refreshExpiresIn
			}
		}`,
		Variables: map[string]any{
			"input": struct {
				APIKey string
			}{
				APIKey: apiKey,
			},
		},
	}

	responseBytes, err := query(q)
	if err != nil {
		return err
	}

	log.Println(string(responseBytes))

	response := struct {
		Data struct {
			ObtainKrakenToken *ObtainKrakenTokenResponse `json:"obtainKrakenToken"`
		} `json:"data"`
	}{}

	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return err
	}

	if response.Data.ObtainKrakenToken == nil {
		log.Println("An error occurred :(")
	} else {
		log.Println(response.Data.ObtainKrakenToken.RefreshToken)
		log.Println(response.Data.ObtainKrakenToken.RefreshExpiresIn)
	}

	return nil
}
