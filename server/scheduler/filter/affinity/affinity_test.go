package affinity

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
			name: "returns true when affinity is nil",
			args: args{
				ctx:    context.Background(),
				runner: &fireactions.Runner{},
				node:   &fireactions.Node{},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "returns true when node matches affinity expression",
			args: args{
				ctx: context.Background(),
				runner: &fireactions.Runner{
					Affinity: []*fireactions.RunnerAffinityExpression{
						{
							Key:      "region",
							Operator: "In",
							Values:   []string{"europe"},
						},
					},
				},
				node: &fireactions.Node{
					Labels: map[string]string{
						"region": "europe",
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "returns false when node doesn't match affinity expression with NotIn operation",
			args: args{
				ctx: context.Background(),
				runner: &fireactions.Runner{
					Affinity: []*fireactions.RunnerAffinityExpression{
						{
							Key:      "region",
							Operator: "NotIn",
							Values:   []string{"europe"},
						},
					},
				},
				node: &fireactions.Node{
					Labels: map[string]string{
						"region": "europe",
					},
				},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "returns false when node doesn't match affinity expression with In operation",
			args: args{
				ctx: context.Background(),
				runner: &fireactions.Runner{
					Affinity: []*fireactions.RunnerAffinityExpression{
						{
							Key:      "region",
							Operator: "In",
							Values:   []string{"europe"},
						},
					},
				},
				node: &fireactions.Node{
					Labels: map[string]string{
						"region": "asia",
					},
				},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "returns false when node doesn't match affinity expression because of no label",
			args: args{
				ctx: context.Background(),
				runner: &fireactions.Runner{
					Affinity: []*fireactions.RunnerAffinityExpression{
						{
							Key:      "region",
							Operator: "NotIn",
							Values:   []string{"europe"},
						},
					},
				},
				node: &fireactions.Node{
					Labels: map[string]string{},
				},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "returns false when node doesn't match affinity expression because of unsupported operation",
			args: args{
				ctx: context.Background(),
				runner: &fireactions.Runner{
					Affinity: []*fireactions.RunnerAffinityExpression{
						{
							Key:      "region",
							Operator: "Equals",
							Values:   []string{"europe"},
						},
					},
				},
				node: &fireactions.Node{
					Labels: map[string]string{"region": "europe"},
				},
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

			assert.Equal(t, "affinity", f.Name())
		})
	}
}
