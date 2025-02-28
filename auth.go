package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
)

type octopusAuth struct {
	Token        string
	RefreshToken string
	// Unix timestamp when the refresh token will expire
	RefreshTokenExpiresAt int64
}

type ObtainKrakenTokenResponse struct {
	Token            string `json:"token"`
	RefreshToken     string `json:"refreshToken"`
	RefreshExpiresIn int    `json:"refreshExpiresIn"`
}

func obtainKrakenToken() (*ObtainKrakenTokenResponse, error) {
	apiKey := os.Getenv("OCTOPUS_API_KEY")

	q := QueryBody{
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

	responseBytes, err := Query(q)
	if err != nil {
		return nil, err
	}

	log.Println(string(responseBytes))

	response := struct {
		Data struct {
			ObtainKrakenToken *ObtainKrakenTokenResponse `json:"obtainKrakenToken"`
		} `json:"data"`
        Errors *[]struct {
            Message string `json:"message"`
        } `json:"errors"`
	}{}

	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, err
	}

	if response.Data.ObtainKrakenToken == nil {
        errorMsg := "Unknown error"
        if response.Errors != nil && len(*response.Errors) > 0 {
            errorMsg = (*response.Errors)[0].Message
        }

        return nil, errors.New(fmt.Sprintf("Failed to obtain Kraken token: %s", errorMsg))
	}

	return response.Data.ObtainKrakenToken, nil
}

func auth() error {
    tokenResult, err := obtainKrakenToken()

    if err != nil {
        log.Println(err)
    }

    log.Println(tokenResult)

	return nil
}
