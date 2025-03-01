package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
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

	responseBytes, err := Query(q, nil)
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

// Make a query to the Octopus API, ensuring we are authenticated first.
func (octo *Octopus) query(q QueryBody) ([]byte, error) {
	err := octo.auth()
	if err != nil {
		return nil, fmt.Errorf("Failed to make Octopus request: %w", err)
	}

	// log.Println(octo.Token)

	headers := map[string]string{
		"Authorization": octo.Token,
	}

	return Query(q, headers)
}

func (octo *Octopus) smartMeterId() (string, error) {
	accountNumber := os.Getenv("OCTOPUS_ACCOUNT_NUMBER")

	if accountNumber == "" {
		return "", errors.New("No account number available; OCTOPUS_ACCOUNT_NUMBER environment variable is not set")
	}

	q := QueryBody{
		Query: `query Account($accountNumber: String!) {
			account(accountNumber: $accountNumber) {
				electricityAgreements(active: true) {
					meterPoint {
						meters(includeInactive: false) {
							smartImportElectricityMeter {
								deviceId
							}
						}
					}
				}
			}
		}`,
		Variables: map[string]any{
			"accountNumber": accountNumber,
		},
	}

	responseBytes, err := octo.query(q)
	if err != nil {
		return "", fmt.Errorf("Get smart meter ID: %w", err)
	}

	response := struct {
		Data struct {
			Account *struct {
				ElectricityAgreements []struct {
					MeterPoint struct {
						Meters []struct {
							SmartImportElectricityMeter struct {
								DeviceId string `json:"deviceId"`
							} `json:"smartImportElectricityMeter"`
						} `json:"meters"`
					} `json:"meterPoint"`
				} `json:"electricityAgreements"`
			} `json:"account"`
		} `json:"data"`
		Errors *[]struct {
			Message string `json:"message"`
		} `json:"errors"`
	}{}

	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return "", fmt.Errorf("Deserialise accounts response: %w", err)
	}

	if response.Errors != nil {
		errorMsg := (*response.Errors)[0].Message
		return "", errors.New(fmt.Sprintf("Failed to obtain account data: %s", errorMsg))
	}

	electricityAgreements := response.Data.Account.ElectricityAgreements
	if len(electricityAgreements) == 0 {
		return "", errors.New("No electricity agreements found")
	}

	meters := electricityAgreements[0].MeterPoint.Meters
	if len(meters) == 0 {
		return "", errors.New("No electricity meters found")
	}

	return meters[0].SmartImportElectricityMeter.DeviceId, nil
}

type ConsumptionReading struct {
	// The point in time that this reading was made.
	Timestamp time.Time
	// The total energy consumption of the meter, in Wh.
	TotalConsumption int
	// The energy consumption since the last reading, in Wh.
	ConsumptionDelta int
	// The current demand at the given timestamp, in W.
	Demand int
}

func (octo *Octopus) LiveConsumption() (*ConsumptionReading, error) {
	deviceId, err := octo.smartMeterId()
	if err != nil {
		return nil, fmt.Errorf("Get live consumption: %w", err)
	}

	q := QueryBody{
		Query: `query SmartMeterTelemetry(
			$deviceId: String!
			$start: DateTime!
			$end: DateTime!
		) {
			smartMeterTelemetry(
				deviceId: $deviceId
				grouping: TEN_SECONDS
				start: $start
				end: $end
			) {
				readAt
				consumption
				consumptionDelta
				demand
			}
		}`,
		Variables: map[string]any{
			"deviceId": deviceId,
			"start":    time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
			"end":      time.Now().Format(time.RFC3339),
		},
	}

	responseBytes, err := octo.query(q)
	if err != nil {
		return nil, fmt.Errorf("Get live consumption: %w", err)
	}

	response := struct {
		Data struct {
			SmartMeterTelemetry *[]struct {
				ReadAt time.Time `json:"readAt"`
				// String containing a float that is always to the nearest integer
				Consumption string `json:"consumption"`
				// String containing a float that is always to the nearest integer
				ConsumptionDelta string `json:"consumptionDelta"`
				// String containing a float that is always to the nearest integer
				Demand string `json:"demand"`
			} `json:"smartMeterTelemetry"`
		} `json:"data"`
		Errors *[]struct {
			Message string `json:"message"`
		} `json:"errors"`
	}{}

	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return nil, fmt.Errorf("Deserialise live consumption: %w", err)
	}

	if response.Errors != nil {
		errorMsg := (*response.Errors)[0].Message
		return nil, errors.New(fmt.Sprintf("Failed to obtain live consumption: %s", errorMsg))
	}

	readings := *response.Data.SmartMeterTelemetry
	if len(readings) == 0 {
		return nil, errors.New("No electricity meter readings found")
	}

	latestReading := readings[len(readings)-1]

	consumption, err := strconv.ParseFloat(latestReading.Consumption, 64)
	if err != nil {
		return nil, fmt.Errorf("Deserialise live consumption: %w", err)
	}
	consumptionDelta, err := strconv.ParseFloat(latestReading.ConsumptionDelta, 64)
	if err != nil {
		return nil, fmt.Errorf("Deserialise live consumption: %w", err)
	}
	demand, err := strconv.ParseFloat(latestReading.Demand, 64)
	if err != nil {
		return nil, fmt.Errorf("Deserialise live consumption: %w", err)
	}

	return &ConsumptionReading{
		Timestamp:        latestReading.ReadAt,
		TotalConsumption: int(consumption),
		ConsumptionDelta: int(consumptionDelta),
		Demand:           int(demand),
	}, nil
}
