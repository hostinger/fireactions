package ramcapacity

import (
	"context"
	"testing"

	"github.com/hostinger/fireactions"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	type args struct {
		ctx    context.Context
		runner *fireactions.Runner
		node   *fireactions.Node
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "returns true when node has enough RAM capacity",
			args: args{
				ctx:    context.Background(),
				runner: &fireactions.Runner{Resources: fireactions.RunnerResources{MemoryBytes: 1}},
				node:   &fireactions.Node{RAM: fireactions.NodeResource{Capacity: 2, OvercommitRatio: 1.0}},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "returns false when node doesn't have enough RAM capacity",
			args: args{
				ctx:    context.Background(),
				runner: &fireactions.Runner{Resources: fireactions.RunnerResources{MemoryBytes: 2}},
				node:   &fireactions.Node{RAM: fireactions.NodeResource{Capacity: 1, OvercommitRatio: 1.0}},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "returns true when node has enough RAM capacity with overcommit",
			args: args{
				ctx:    context.Background(),
				runner: &fireactions.Runner{Resources: fireactions.RunnerResources{MemoryBytes: 2}},
				node:   &fireactions.Node{RAM: fireactions.NodeResource{Capacity: 1, OvercommitRatio: 2.0}},
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := New()
			got, err := f.Filter(tt.args.ctx, tt.args.runner, tt.args.node)
			if (err != nil) != tt.wantErr {
				t.Errorf("Filter.Filter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Filter.Filter() = %v, want %v", got, tt.want)
			}

			assert.Equal(t, "ram-capacity", f.Name())
		})
	}
}
