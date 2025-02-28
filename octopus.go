package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// Encapsulates all methods for interacting with the Octopus API.
type Octopus struct {
	// The kraken authentication token to use on API requests. Valid for
	// one hour.
	Token string
	// Unix timestamp when the token will expire.
	TokenExpiresAt int64
	// The refresh token to use when the auth token expires. Valid for one week.
	RefreshToken string
	// Unix timestamp when the refresh token will expire.
	RefreshTokenExpiresAt int64
}

// Checks if we have a valid auth token that has not expired.
func (octo *Octopus) hasValidToken() bool {
	if octo == nil || octo.Token == "" {
		return false
	}
	return time.Now().Unix() < octo.TokenExpiresAt
}

// Checks if we have a valid refresh token that has not expired.
func (octo *Octopus) hasValidRefreshToken() bool {
	if octo == nil || octo.RefreshToken == "" {
		return false
	}
	return time.Now().Unix() < octo.RefreshTokenExpiresAt
}

// Sends an API request to obtain a kraken auth token. The input can be
// either an API key or a refresh token. Callers should use
// [Octopus.authWithApiKey] or [Octopus.authWithRefreshToken] rather than
// this method directly; they provide the necessary input arguments.
func (octo *Octopus) obtainKrakenToken(input any) error {
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
		return err
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
		return err
	}

	if response.Data.ObtainKrakenToken == nil {
		errorMsg := "Unknown error"
		if response.Errors != nil && len(*response.Errors) > 0 {
			errorMsg = (*response.Errors)[0].Message
		}

		return errors.New(fmt.Sprintf("Failed to obtain Kraken token: %s", errorMsg))
	}

	octo.Token = response.Data.ObtainKrakenToken.Token
	// Auth token is valid for one hour
	octo.TokenExpiresAt = time.Now().Add(time.Hour).Unix()
	octo.RefreshToken = response.Data.ObtainKrakenToken.RefreshToken
	octo.RefreshTokenExpiresAt = int64(response.Data.ObtainKrakenToken.RefreshExpiresIn)

	return nil
}

// Wraps around [Octopus.obtainKrakenToken] to authenticate with the user's
// API key. Expects the API key to be provided via the OCTOPUS_API_KEY
// environment variable.
func (octo *Octopus) authWithApiKey() error {
	apiKey := os.Getenv("OCTOPUS_API_KEY")

	if apiKey == "" {
		return errors.New("No API key available; OCTOPUS_API_KEY environment variable is not set")
	}

	return octo.obtainKrakenToken(struct {
		APIKey string
	}{
		APIKey: apiKey,
	})
}

// Wraps around [Octopus.obtainKrakenToken] to authenticate with the stored
// refresh token. It is an error to call this method with an invalid refresh
// token. Use [Octopus.hasValidRefreshToken] to check the validity of the
// token before calling this method.
func (octo *Octopus) authWithRefreshToken() error {
	if octo.RefreshToken == "" {
		return errors.New("No refresh token available")
	}

	return octo.obtainKrakenToken(struct {
		refreshToken string
	}{
		refreshToken: octo.RefreshToken,
	})
}

// Authenticates to the Octopus API, obtaining a Kraken token if necessary.
// This method should be called before making any API calls that require
// authentication.
func (octo *Octopus) auth() error {
	if octo.hasValidToken() {
		// We have a token; nothing to do here
		return nil
	}

	if octo.hasValidRefreshToken() {
		// Token has expired but refresh token is still valid
		// Authenticate with refresh token
		err := octo.authWithRefreshToken()
		if err != nil {
			return fmt.Errorf("Failed to get kraken token: %w", err)
		}

		return nil
	}

	// No valid token or refresh token
	// authenticate fresh
	err := octo.authWithApiKey()
	if err != nil {
		return fmt.Errorf("Failed to get kraken token: %w", err)
	}

	return nil
}
