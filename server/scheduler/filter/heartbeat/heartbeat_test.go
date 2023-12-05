package heartbeat

import (
	"context"
	"testing"
	"time"

	"github.com/hostinger/fireactions"
	"gotest.tools/assert"
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
			name: "returns true when last reconcile is less than reconcile interval",
			args: args{
				ctx:    context.Background(),
				runner: &fireactions.Runner{},
				node:   &fireactions.Node{LastReconcileAt: time.Now(), ReconcileInterval: 1 * time.Second},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "returns false when last reconcile is more than reconcile interval",
			args: args{
				ctx:    context.Background(),
				runner: &fireactions.Runner{},
				node:   &fireactions.Node{LastReconcileAt: time.Now().Add(-2 * time.Second), ReconcileInterval: 1 * time.Second},
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

			assert.Equal(t, f.Name(), "heartbeat")
		})
	}
}
