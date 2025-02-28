package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// The token to use to authenticate with the Octopus API.
type krakenToken struct {
	Token string
	// Unix timestamp when the token will expire
	TokenExpiresAt int64
	RefreshToken   string
	// Unix timestamp when the refresh token will expire
	RefreshTokenExpiresAt int64
}

func (token *krakenToken) hasTokenExpired() bool {
	if token == nil {
		return true
	}
	return time.Now().Unix() < token.TokenExpiresAt
}

func (token *krakenToken) hasRefreshTokenExpired() bool {
	if token == nil {
		return true
	}
	return time.Now().Unix() < token.RefreshTokenExpiresAt
}

func obtainKrakenToken(input any) (*krakenToken, error) {
	q := QueryBody{
		Query: `mutation ObtainKrakenToken($input: ObtainJSONWebTokenInput!) {
			obtainKrakenToken(input: $input) {
				token
				refreshToken
				refreshExpiresIn
			}
		}`,
		Variables: map[string]any{
			"input": input,
		},
	}

	responseBytes, err := Query(q)
	if err != nil {
		return nil, err
	}

	response := struct {
		Data struct {
			ObtainKrakenToken *struct {
				Token            string `json:"token"`
				RefreshToken     string `json:"refreshToken"`
				RefreshExpiresIn int    `json:"refreshExpiresIn"`
			} `json:"obtainKrakenToken"`
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

	result := krakenToken{
		Token:                 response.Data.ObtainKrakenToken.Token,
		TokenExpiresAt:        time.Now().Add(time.Hour).Unix(),
		RefreshToken:          response.Data.ObtainKrakenToken.RefreshToken,
		RefreshTokenExpiresAt: int64(response.Data.ObtainKrakenToken.RefreshExpiresIn),
	}

	return &result, nil
}

func authWithApiKey() (*krakenToken, error) {
	apiKey := os.Getenv("OCTOPUS_API_KEY")

	return obtainKrakenToken(struct {
		APIKey string
	}{
		APIKey: apiKey,
	})
}

func authWithRefreshToken(token string) (*krakenToken, error) {
	return obtainKrakenToken(struct {
		refreshToken string
	}{
		refreshToken: token,
	})
}

var storedToken *krakenToken

func auth() error {
	if !storedToken.hasTokenExpired() {
		// We have a token; nothing to do here
		return nil
	}

	if !storedToken.hasRefreshTokenExpired() {
		// Token has expired but refresh token is still valid
		// Authenticate with refresh token
		tokenResult, err := authWithRefreshToken(storedToken.RefreshToken)
		if err != nil {
			return fmt.Errorf("Failed to get kraken token: %w", err)
		}

		storedToken = tokenResult
		return nil
	}

	// No valid token or refresh token
	// authenticate fresh
	tokenResult, err := authWithApiKey()
	if err != nil {
		return fmt.Errorf("Failed to get kraken token: %w", err)
	}

	storedToken = tokenResult
	return nil
}
