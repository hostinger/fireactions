package ghtokencache

import (
	"sync"
	"time"

	"github.com/google/go-github/v53/github"
)

// Cache is a GitHub registration token cache.
type Cache struct {
	registrationTokens map[string]*github.RegistrationToken
	removeTokens       map[string]*github.RemoveToken
	mu                 sync.RWMutex
}

// New returns a new GitHub registration token cache.
func New() *Cache {
	c := &Cache{
		registrationTokens: map[string]*github.RegistrationToken{},
		removeTokens:       map[string]*github.RemoveToken{},
		mu:                 sync.RWMutex{},
	}

	return c
}

// SetToken sets a GitHub registration token for the specified organisation.
func (c *Cache) SetRegistrationToken(org string, token *github.RegistrationToken) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.registrationTokens[org] = token
}

// SetRemoveToken sets a GitHub remove token for the specified organisation.
func (c *Cache) SetRemoveToken(org string, token *github.RemoveToken) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.removeTokens[org] = token
}

// GetToken returns a GitHub registration token for the specified organisation. If the token is close to expiring (< 1m), it
// will return nil. If no token is found, it will return nil.
func (c *Cache) GetRegistrationToken(org string) *string {
	c.mu.RLock()
	defer c.mu.RUnlock()

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
func (c *Cache) GetRemoveToken(org string) *string {
	c.mu.RLock()
	defer c.mu.RUnlock()

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
