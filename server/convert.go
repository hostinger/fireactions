package server

import (
	"github.com/hostinger/fireactions"
	"github.com/hostinger/fireactions/helper/deepcopy"
	"github.com/hostinger/fireactions/helper/stringid"
)

func convertJobLabelConfigToRunner(config *JobLabelConfig, organisation string) *fireactions.Runner {
	runnerID := stringid.New()
	runner := &fireactions.Runner{
		ID:              runnerID,
		Organisation:    organisation,
		Name:            config.MustGetRunnerName(runnerID),
		NodeID:          nil,
		Status:          fireactions.RunnerStatus{State: fireactions.RunnerStatePending, Description: "Created"},
		Labels:          config.GetRunnerLabels(),
		Resources:       config.RunnerResources,
		ImagePullPolicy: config.RunnerImagePullPolicy,
		Image:           config.RunnerImage,
		Metadata:        deepcopy.Map(config.RunnerMetadata),
		Affinity:        config.RunnerAffinity,
	}

	return runner
}
