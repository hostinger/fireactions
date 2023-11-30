package status

import (
	"context"
	"testing"

	"github.com/hostinger/fireactions"
	"github.com/stretchr/testify/assert"
)

func TestFiler(t *testing.T) {
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
			name: "returns true when node is ready",
			args: args{
				ctx:    context.Background(),
				runner: &fireactions.Runner{},
				node:   &fireactions.Node{Status: fireactions.NodeStatus{State: fireactions.NodeStatusReady}},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "returns false when node is not ready",
			args: args{
				ctx:    context.Background(),
				runner: &fireactions.Runner{},
				node:   &fireactions.Node{Status: fireactions.NodeStatus{State: fireactions.NodeStateNotReady}},
			},
			want:    false,
			wantErr: true,
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

			assert.Equal(t, f.Name(), "status")
		})
	}
}
