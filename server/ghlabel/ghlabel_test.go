package ghlabel

import "testing"

func TestLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    string
		want *Label
	}{
		{
			name: "",
			s:    "",
			want: &Label{},
		},
		{
			name: "group1",
			s:    "group1",
			want: &Label{
				Group: "group1",
			},
		},
		{
			name: "group1-1vcpu-2gb",
			s:    "group1-1vcpu-2gb",
			want: &Label{
				Group:  "group1",
				Flavor: "1vcpu-2gb",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.s)

			if l.Group != tt.want.Group {
				t.Errorf("got %s; want %s", l.Group, tt.want.Group)
			}

			if l.Flavor != tt.want.Flavor {
				t.Errorf("got %s; want %s", l.Flavor, tt.want.Flavor)
			}

			if l.String() != tt.s {
				t.Errorf("got %s; want %s", l.String(), tt.s)
			}
		})
	}
}
