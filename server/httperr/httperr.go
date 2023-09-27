package httperr

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type errorMap struct {
	fromError    error
	toStatusCode int
	toMessage    string
}

func Map(err error) *errorMap {
	return &errorMap{fromError: err}
}

func (em *errorMap) To(statusCode int, message string) *errorMap {
	em.toStatusCode = statusCode
	em.toMessage = message
	return em
}

func HandlerFunc(errMaps ...*errorMap) gin.HandlerFunc {
	f := func(c *gin.Context) {
		c.Next()

		lastErr := c.Errors.Last()
		if lastErr == nil {
			return
		}

		for _, errMap := range errMaps {
			if !errors.Is(lastErr.Err, errMap.fromError) {
				continue
			}

			c.AbortWithStatusJSON(errMap.toStatusCode, gin.H{"error": errMap.toMessage})
			return
		}

		c.AbortWithStatusJSON(500, gin.H{"error": lastErr.Error()})
	}

	return f
}
