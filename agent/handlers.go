package agent

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/hostinger/fireactions/agent/runner"
)

var (
	ErrAlreadyRunning = errors.New("already running")
	ErrNotRunning     = errors.New("not running")
)

func (a *Agent) healthzHandlerFunc() gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "OK"})
	}

	return f
}

func (a *Agent) startHandlerFunc() gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		var req StartRequest
		err := ctx.BindJSON(&req)
		if err != nil {
			ctx.JSON(400, gin.H{"error": fmt.Sprintf("bad request: %s", err.Error())})
			return
		}

		runnerOpts := []runner.Opt{
			runner.WithDisableUpdate(req.DisableUpdate),
			runner.WithReplace(req.Replace),
			runner.WithEphemeral(req.Ephemeral),
		}

		err = a.StartGitHubRunner(ctx, req.Name, req.URL, req.Token, req.Labels, runnerOpts...)
		if err != nil && err != ErrAlreadyRunning {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"message": "OK"})
	}

	return f
}

func (a *Agent) stopHandlerFunc() gin.HandlerFunc {
	f := func(ctx *gin.Context) {
		var req StopRequest
		err := ctx.BindJSON(&req)
		if err != nil {
			ctx.JSON(400, gin.H{"error": fmt.Sprintf("bad request: %s", err.Error())})
			return
		}

		err = a.StopGitHubRunner(ctx, req.Token)
		if err != nil && err != ErrNotRunning {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(200, gin.H{"message": "OK"})
	}

	return f
}
