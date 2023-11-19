package planner

import (
	"context"
	"fmt"
)

type Plan interface {
	Name() string

	Create(ctx context.Context) ([]Procedure, error)
}

type Procedure interface {
	Name() string

	Do(ctx context.Context) ([]Procedure, error)
}

type Planner interface {
	Execute(ctx context.Context, plan Plan) error
}

func NewPlanner() *plannerImpl {
	return &plannerImpl{}
}

type plannerImpl struct{}

func (e *plannerImpl) Execute(ctx context.Context, plan Plan) error {
	steps, err := plan.Create(ctx)
	if err != nil {
		return err
	}

	if len(steps) == 0 {
		return nil
	}

	if err := e.executeSteps(ctx, steps); err != nil {
		return fmt.Errorf("failed to execute plan %s: %w", plan.Name(), err)
	}

	return nil
}

func (e *plannerImpl) executeSteps(ctx context.Context, steps []Procedure) error {
	var innerSteps []Procedure
	for _, step := range steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var err error
		innerSteps, err = step.Do(ctx)
		if err != nil {
			return fmt.Errorf("step %s: %w", step.Name(), err)
		}

		if len(innerSteps) < 1 {
			continue
		}

		return e.executeSteps(ctx, innerSteps)
	}

	return nil
}
