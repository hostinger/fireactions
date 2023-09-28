package ghlabel

import "testing"

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("should return error if regexp does not match", func(t *testing.T) {
		_, err := New("foo2bar.foo2bar", WithDefaultFlavor("bar"))
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("should return Label with default flavor", func(t *testing.T) {
		l, err := New("group1", WithDefaultFlavor("1vcpu-2gb"))
		if err != nil {
			t.Fatalf("expected nil, got %s", err)
		}

		if l.Flavor != "1vcpu-2gb" {
			t.Errorf("expected flavor to be 1vcpu-2gb, got %s", l.Flavor)
		}
	})

	t.Run("should return Label with flavor", func(t *testing.T) {
		l, err := New("group1-1vcpu-2gb")
		if err != nil {
			t.Fatalf("expected nil, got %s", err)
		}

		if l.Flavor != "1vcpu-2gb" {
			t.Errorf("expected flavor to be 1vcpu-2gb, got %s", l.Flavor)
		}

		if l.Group != "group1" {
			t.Errorf("expected group to be group1, got %s", l.Group)
		}
	})
}
