package github

import (
	"sync"
	"time"

	"github.com/google/go-github/v53/github"
)

type tokenCache struct {
	registrationTokens map[string]*github.RegistrationToken
	removeTokens       map[string]*github.RemoveToken
	l                  sync.RWMutex
}

func newTokenCache() *tokenCache {
	c := &tokenCache{registrationTokens: map[string]*github.RegistrationToken{}, removeTokens: map[string]*github.RemoveToken{}, l: sync.RWMutex{}}
	return c
}

// SetToken sets a GitHub registration token for the specified organisation.
func (c *tokenCache) SetRegistrationToken(org string, token *github.RegistrationToken) {
	c.l.Lock()
	defer c.l.Unlock()

	c.registrationTokens[org] = token
}

// SetRemoveToken sets a GitHub remove token for the specified organisation.
func (c *tokenCache) SetRemoveToken(org string, token *github.RemoveToken) {
	c.l.Lock()
	defer c.l.Unlock()

	c.removeTokens[org] = token
}

// GetToken returns a GitHub registration token for the specified organisation. If the token is close to expiring (< 1m), it
// will return nil. If no token is found, it will return nil.
func (c *tokenCache) GetRegistrationToken(org string) *string {
	c.l.RLock()
	defer c.l.RUnlock()

	token, ok := c.registrationTokens[org]
	if !ok {
		return nil
	}

	if time.Until(token.ExpiresAt.Time) < time.Minute { // Token is about to expire (< 1m), don't use it
		return nil
	}

	t := token.GetToken()
	return &t
}

// GetRemoveToken returns a GitHub remove token for the specified organisation. If the token is close to expiring (< 1m), it
// will return nil. If no token is found, it will return nil.
func (c *tokenCache) GetRemoveToken(org string) *string {
	c.l.RLock()
	defer c.l.RUnlock()

	token, ok := c.removeTokens[org]
	if !ok {
		return nil
	}

	if time.Until(token.ExpiresAt.Time) < time.Minute { // Token is about to expire (< 1m), don't use it
		return nil
	}

	t := token.GetToken()
	return &t
}
