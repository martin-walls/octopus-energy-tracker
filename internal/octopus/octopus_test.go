package octopus

import (
	"testing"
	"time"
)

type hasValidTokenTest struct {
	octo               *Octopus
	expectedTokenValid bool
}

var hasValidTokenTests = []hasValidTokenTest{
	{
		&Octopus{
			Token:          "authtoken",
			TokenExpiresAt: time.Now().Add(time.Minute).Unix(),
		},
		true,
	},
	{
		&Octopus{
			Token:          "authtoken",
			TokenExpiresAt: time.Now().Add(-2 * time.Minute).Unix(),
		},
		false,
	},
	{
		&Octopus{
			Token: "",
		},
		false,
	},
	{
		&Octopus{},
		false,
	},
	{
		nil,
		false,
	},
}

func TestHasValidToken(t *testing.T) {
	for _, test := range hasValidTokenTests {
		tokenValid := test.octo.hasValidToken()
		if tokenValid != test.expectedTokenValid {
			t.Errorf("hasValidToken() = %t, want %t", tokenValid, test.expectedTokenValid)
		}
	}
}

var hasValidRefreshTokenTests = []hasValidTokenTest{
	{
		&Octopus{
			RefreshToken:          "refreshtoken",
			RefreshTokenExpiresAt: time.Now().Add(time.Hour).Unix(),
		},
		true,
	},
	{
		&Octopus{
			RefreshToken:          "refreshtoken",
			RefreshTokenExpiresAt: time.Now().Add(-2 * time.Minute).Unix(),
		},
		false,
	},
	{
		&Octopus{
			RefreshToken: "",
		},
		false,
	},
	{
		&Octopus{},
		false,
	},
	{
		nil,
		false,
	},
}

func TestHasValidRefreshToken(t *testing.T) {
	for _, test := range hasValidRefreshTokenTests {
		tokenValid := test.octo.hasValidRefreshToken()
		if tokenValid != test.expectedTokenValid {
			t.Errorf("hasValidRefreshToken() = %t, want %t", tokenValid, test.expectedTokenValid)
		}
	}
}
