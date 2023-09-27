package containerd

import (
	"context"
	"fmt"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/leases"
)

func leaseName(owner string) string {
	return fmt.Sprintf("fireactions/%s", owner)
}

// newContextWithOwnerLease creates a new context with a lease for the given owner. If a lease already exists for the
// owner, it will be reused.
func newContextWithOwnerLease(ctx context.Context, client *containerd.Client, owner string) (context.Context, error) {
	leaseName := leaseName(owner)

	leasesService := client.LeasesService()
	existingLeases, err := leasesService.List(ctx, fmt.Sprintf("id==%s", leaseName))
	if err != nil {
		return nil, err
	}

	for _, lease := range existingLeases {
		if lease.ID != leaseName {
			continue
		}

		return leases.WithLease(ctx, lease.ID), nil
	}

	lease, err := leasesService.Create(ctx, leases.WithID(leaseName))
	if err != nil {
		return nil, err
	}

	return leases.WithLease(ctx, lease.ID), nil
}

func deleteLease(ctx context.Context, client *containerd.Client, owner string) error {
	leaseName := leaseName(owner)

	leasesService := client.LeasesService()
	existingLeases, err := leasesService.List(ctx, fmt.Sprintf("id==%s", leaseName))
	if err != nil {
		return err
	}

	for _, lease := range existingLeases {
		if lease.ID != leaseName {
			continue
		}

		return leasesService.Delete(ctx, lease)
	}

	return nil
}
