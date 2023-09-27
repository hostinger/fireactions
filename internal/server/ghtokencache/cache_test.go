package ghtokencache

import (
	"testing"
	"time"

	"github.com/google/go-github/v53/github"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	c := New()

	assert.NotNil(t, c)
}

func TestSetToken(t *testing.T) {
	c := New()

	token := &github.RegistrationToken{
		Token:     github.String("token"),
		ExpiresAt: &github.Timestamp{Time: time.Now().Add(time.Hour)},
	}

	c.SetToken("org", token)

	assert.Equal(t, token, c.tokens["org"])
}

func TestGetToken(t *testing.T) {
	c := New()

	token := &github.RegistrationToken{
		Token:     github.String("token"),
		ExpiresAt: &github.Timestamp{Time: time.Now().Add(time.Hour)},
	}

	c.SetToken("org", token)

	assert.Equal(t, token.Token, c.GetToken("org"))
}

func TestGetTokenExpired(t *testing.T) {
	c := New()

	token := &github.RegistrationToken{
		Token:     github.String("token"),
		ExpiresAt: &github.Timestamp{Time: time.Now().Add(-time.Hour)},
	}

	c.SetToken("org", token)

	assert.Nil(t, c.GetToken("org"))
}

func TestGetTokenExpiresSoon(t *testing.T) {
	c := New()

	token := &github.RegistrationToken{
		Token:     github.String("token"),
		ExpiresAt: &github.Timestamp{Time: time.Now().Add(time.Minute)},
	}

	c.SetToken("org", token)

	assert.Nil(t, c.GetToken("org"))
}

func TestGetTokenNotFound(t *testing.T) {
	c := New()

	assert.Nil(t, c.GetToken("org"))
}
