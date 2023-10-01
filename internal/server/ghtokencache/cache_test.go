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

func TestSetRegistrationToken(t *testing.T) {
	c := New()

	token := &github.RegistrationToken{
		Token:     github.String("token"),
		ExpiresAt: &github.Timestamp{Time: time.Now().Add(time.Hour)},
	}

	c.SetRegistrationToken("org", token)

	assert.Equal(t, token, c.registrationTokens["org"])
}

func TestSetRemoveToken(t *testing.T) {
	c := New()

	token := &github.RemoveToken{
		Token: github.String("token"),
	}

	c.SetRemoveToken("org", token)

	assert.Equal(t, token, c.removeTokens["org"])
}

func TestGetRegistrationToken(t *testing.T) {
	c := New()

	token := &github.RegistrationToken{
		Token:     github.String("token"),
		ExpiresAt: &github.Timestamp{Time: time.Now().Add(time.Hour)},
	}

	c.SetRegistrationToken("org", token)

	assert.Equal(t, token.Token, c.GetRegistrationToken("org"))
}

func TestGetRegistrationToken_Expired(t *testing.T) {
	c := New()

	token := &github.RegistrationToken{
		Token:     github.String("token"),
		ExpiresAt: &github.Timestamp{Time: time.Now().Add(-time.Hour)},
	}

	c.SetRegistrationToken("org", token)

	assert.Nil(t, c.GetRegistrationToken("org"))
}

func TestGetRegistrationToken_ExpiresIn1m(t *testing.T) {
	c := New()

	token := &github.RegistrationToken{
		Token:     github.String("token"),
		ExpiresAt: &github.Timestamp{Time: time.Now().Add(time.Minute)},
	}

	c.SetRegistrationToken("org", token)

	assert.Nil(t, c.GetRegistrationToken("org"))
}

func TestGetRegistrationToken_NotFound(t *testing.T) {
	c := New()

	assert.Nil(t, c.GetRegistrationToken("org"))
}

func TestGetRemoveToken(t *testing.T) {
	c := New()

	token := &github.RemoveToken{
		Token:     github.String("token"),
		ExpiresAt: &github.Timestamp{Time: time.Now().Add(time.Hour)},
	}

	c.SetRemoveToken("org", token)

	assert.Equal(t, token.Token, c.GetRemoveToken("org"))
}

func TestGetRemoveToken_Expired(t *testing.T) {
	c := New()

	token := &github.RemoveToken{
		Token:     github.String("token"),
		ExpiresAt: &github.Timestamp{Time: time.Now().Add(-time.Hour)},
	}

	c.SetRemoveToken("org", token)

	assert.Nil(t, c.GetRemoveToken("org"))
}

func TestGetRemoveToken_NotFound(t *testing.T) {
	c := New()

	assert.Nil(t, c.GetRemoveToken("org"))
}

func TestGetRemoveToken_ExpiresIn1m(t *testing.T) {
	c := New()

	token := &github.RemoveToken{
		Token:     github.String("token"),
		ExpiresAt: &github.Timestamp{Time: time.Now().Add(time.Minute)},
	}

	c.SetRemoveToken("org", token)

	assert.Nil(t, c.GetRemoveToken("org"))
}
