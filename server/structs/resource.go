package structs

import "fmt"

type Resource struct {
	Allocated       int64
	Capacity        int64
	OvercommitRatio float64
}

func (r *Resource) Reserve(amount int64) {
	r.Allocated += amount
}

func (r *Resource) Release(amount int64) {
	r.Allocated -= amount
}

func (r *Resource) Available() int64 {
	return r.Capacity - r.Allocated
}

func (r *Resource) String() string {
	return fmt.Sprintf("%d/%d", r.Allocated, r.Capacity)
}

func (r *Resource) IsAvailable(amount int64) bool {
	return float64(r.Allocated+amount) <= float64(r.Capacity)*r.OvercommitRatio
}

func (r *Resource) IsFull() bool {
	return r.Allocated >= r.Capacity
}
