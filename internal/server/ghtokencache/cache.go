package ghtokencache

import (
	"sync"
	"time"

	"github.com/google/go-github/v53/github"
)

// Cache is a GitHub registration token cache.
type Cache struct {
	tokens map[string]*github.RegistrationToken
	mu     sync.RWMutex
}

// New returns a new GitHub registration token cache.
func New() *Cache {
	c := &Cache{
		tokens: make(map[string]*github.RegistrationToken),
		mu:     sync.RWMutex{},
	}

	return c
}

// SetToken sets a GitHub registration token for the specified organisation.
func (c *Cache) SetToken(org string, token *github.RegistrationToken) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tokens[org] = token
}

// GetToken returns a GitHub registration token for the specified organisation. If the token is close to expiring (< 1m), it
// will return nil. If no token is found, it will return nil.
func (c *Cache) GetToken(org string) *string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	token, ok := c.tokens[org]
	if !ok {
		return nil
	}

	if time.Until(token.ExpiresAt.Time) < time.Minute { // Token is about to expire (< 1m), don't use it
		return nil
	}

	t := token.GetToken()
	return &t
}
