package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/cmd/fireactions-agent/metadata"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.
		New(os.Stdout).Level(zerolog.DebugLevel).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	})

	logger.Info().Msgf("starting Fireactions Agent")

	metadataClient, err := metadata.NewClient(nil)
	if err != nil {
		logger.Fatal().Err(err).Msgf("error creating Firecracker MMDS client")
	}
	defer metadataClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	metadata, _, err := metadataClient.GetMetadata(ctx, "/", nil)
	if err != nil {
		logger.Fatal().Err(err).Msgf("error getting metadata from Firecracker MMDS")
	}

	fireactionsServerURL, ok := metadata["fireactions-server-url"].(string)
	if !ok {
		logger.Fatal().Msgf("error getting 'fireactions-server-url' from Firecracker MMDS")
	}

	fireactionsRunnerID, ok := metadata["fireactions-runner-id"].(string)
	if !ok {
		logger.Fatal().Msgf("error getting 'fireactions-runner-id' from Firecracker MMDS")
	}

	fireactionsClient := fireactions.NewClient(nil, fireactions.WithEndpoint(fireactionsServerURL))
	runner, _, err := fireactionsClient.Runners().Get(context.Background(), fireactionsRunnerID)
	if err != nil {
		logger.Fatal().Err(err).Msgf("error getting Fireactions Runner")
	}

	err = startGitHubRunner(&logger, context.Background(), fireactionsClient, runner)
	if err != nil {
		logger.Fatal().Err(err).Msgf("error starting GitHub runner")
	}

	logger.Info().Msgf("stopping Fireactions Agent")
}

func startGitHubRunner(logger *zerolog.Logger, ctx context.Context, client *fireactions.Client, runner *fireactions.Runner) error {
	cwd, _ := os.Getwd()
	err := configureGitHubRunner(logger, ctx, client, runner, cwd)
	if err != nil {
		return fmt.Errorf("error configuring GitHub runner: %w", err)
	}

	logger.Info().Msgf("starting GitHub runner")

	runsvcShPath := fmt.Sprintf("%s/run.sh", cwd)
	if _, err := os.Stat(runsvcShPath); os.IsNotExist(err) {
		return fmt.Errorf("%s/run.sh does not exist", cwd)
	}

	runsvcCmd := exec.CommandContext(ctx, runsvcShPath)
	runsvcCmd.Stdout = os.Stdout
	runsvcCmd.Stderr = os.Stderr

	err = runsvcCmd.Start()
	if err != nil {
		return fmt.Errorf("error starting run.sh: %w", err)
	}

	_, err = client.
		Runners().
		SetStatus(ctx, runner.ID, fireactions.RunnerSetStatusRequest{Phase: fireactions.RunnerPhaseRunning})
	if err != nil {
		logger.Error().Err(err).Msgf("error setting Fireactions Runner status to %s", fireactions.RunnerPhaseRunning)
	}

	err = runsvcCmd.Wait()
	if err != nil {
		_, err = client.
			Runners().
			SetStatus(ctx, runner.ID, fireactions.RunnerSetStatusRequest{Phase: fireactions.RunnerPhaseFailed})
		if err != nil {
			logger.Error().Err(err).Msgf("error setting Fireactions Runner status to %s", fireactions.RunnerPhaseFailed)
		}

		return fmt.Errorf("run.sh exited with error: %w", err)
	}

	_, err = client.
		Runners().
		Complete(ctx, runner.ID)
	if err != nil {
		logger.Error().Err(err).Msgf("error completing Fireactions Runner")
	}

	logger.Info().Msgf("GitHub runner stopped")
	return nil
}

func configureGitHubRunner(logger *zerolog.Logger, ctx context.Context, client *fireactions.Client, runner *fireactions.Runner, runnerPath string) error {
	_, err := os.Stat(fmt.Sprintf("%s/.runner", runnerPath))
	if err == nil {
		logger.Info().Msgf("GitHub runner already configured. Skipping configuration. To reconfigure, delete the .runner file.")
		return nil
	}

	logger.Info().Msgf("configuring GitHub runner")

	logger.Info().Msgf("retrieving GitHub runner registration token")
	token, _, err := client.
		Runners().
		GetRegistrationToken(ctx, runner.ID)
	if err != nil {
		return fmt.Errorf("error getting GitHub runner registration token: %w", err)
	}

	configShPath := fmt.Sprintf("%s/config.sh", runnerPath)
	if _, err := os.Stat(configShPath); os.IsNotExist(err) {
		return fmt.Errorf("%s/config.sh does not exist", runnerPath)
	}

	configArgs := []string{"--url", fmt.Sprintf("https://github.com/%s", runner.Organisation), "--name", runner.Name,
		"--labels", strings.Join(runner.Labels, ","), "--token", token.Token, "--ephemeral", "--replace", "--disableupdate", "--unattended", "--no-default-labels", "--replace"}

	logger.Info().Msgf("starting config.sh with args: %s", strings.Join(configArgs, " "))
	configCmd := exec.CommandContext(ctx, configShPath, configArgs...)
	configCmd.Stdout = os.Stdout
	configCmd.Stderr = os.Stderr

	err = configCmd.Run()
	if err != nil {
		return fmt.Errorf("error running config.sh: %w", err)
	}

	logger.Info().Msgf("GitHub runner configured")
	return nil
}
