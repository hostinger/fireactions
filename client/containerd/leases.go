package containerd

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/leases"
)

// CreateLease creates a new Lease if it does not already exist. If it does
// exist, it returns the existing Lease.
func CreateLease(ctx context.Context, client Client, leaseID string, opts ...leases.Opt) (*leases.Lease, error) {
	leasesService := client.LeasesService()
	existingLeases, err := leasesService.List(ctx, fmt.Sprintf("id==%s", leaseID))
	if err != nil {
		return nil, err
	}

	for _, lease := range existingLeases {
		if lease.ID != leaseID {
			continue
		}

		return &lease, nil
	}

	opts = append(opts, leases.WithID(leaseID))
	lease, err := leasesService.Create(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return &lease, nil
}

// DeleteLease deletes a Lease if it exists.
func DeleteLease(ctx context.Context, client Client, leaseID string) error {
	leasesService := client.LeasesService()
	existingLeases, err := leasesService.List(ctx, fmt.Sprintf("id==%s", leaseID))
	if err != nil {
		return err
	}

	for _, lease := range existingLeases {
		if lease.ID != leaseID {
			continue
		}

		return leasesService.Delete(ctx, lease)
	}

	return nil
}

// NewContextWithLease creates a new context with a Lease. The Lease is created
// if it does not already exist.
func NewContextWithLease(ctx context.Context, client Client, leaseID string, opts ...leases.Opt) (context.Context, error) {
	lease, err := CreateLease(ctx, client, leaseID, opts...)
	if err != nil {
		return nil, err
	}

	return leases.WithLease(ctx, lease.ID), nil
}
