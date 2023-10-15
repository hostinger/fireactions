package bbolt

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"go.etcd.io/bbolt"
)

/*
Store is a bbolt implementation of the Store interface using BoltDB.

Current BoltDB schema:
|-- settings
|   |-- default-flavor -> models.Flavor
|   |-- default-group  -> models.Group
|-- nodes
|   |-- <ID>           -> models.Node
|-- jobs
|   |-- <ID>           -> models.Job
|-- runners
|   |-- <ID>           -> models.Runner
|-- groups
|   |-- <name>         -> models.Group
|-- flavors
|   |-- <name>         -> models.Flavor
*/
type Store struct {
	db *bbolt.DB

	// Metrics
	resourceCount *prometheus.GaugeVec
	scrapeErrors  prometheus.Counter
}

// New creates a new bbolt Store.
func New(path string) (*Store, error) {
	db, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	s := &Store{
		db: db,
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		buckets := []string{"settings", "nodes", "jobs", "runners", "groups", "flavors", "images"}
		for _, bucket := range buckets {
			_, err := tx.CreateBucketIfNotExists([]byte(bucket))
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	s.resourceCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name:      "resources_total",
		Namespace: "fireactions",
		Subsystem: "store",
		Help:      "Number of resources in the store (nodes, jobs, runners, groups, flavors)",
	}, []string{"resource"})

	s.scrapeErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name:      "scrape_errors_total",
		Namespace: "fireactions",
		Subsystem: "store",
		Help:      "Number of errors while scraping the store",
	})

	return s, nil
}

// Close closes the Store.
func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Collect(ch chan<- prometheus.Metric) {
	count, err := s.GetNodesCount(context.Background())
	if err != nil {
		s.scrapeErrors.Inc()
	} else {
		s.resourceCount.WithLabelValues("nodes").Set(float64(count))
	}

	count, err = s.GetJobsCount(context.Background())
	if err != nil {
		s.scrapeErrors.Inc()
	} else {
		s.resourceCount.WithLabelValues("jobs").Set(float64(count))
	}

	count, err = s.GetRunnersCount(context.Background())
	if err != nil {
		s.scrapeErrors.Inc()
	} else {
		s.resourceCount.WithLabelValues("runners").Set(float64(count))
	}

	count, err = s.GetGroupsCount(context.Background())
	if err != nil {
		s.scrapeErrors.Inc()
	} else {
		s.resourceCount.WithLabelValues("groups").Set(float64(count))
	}

	count, err = s.GetFlavorsCount(context.Background())
	if err != nil {
		s.scrapeErrors.Inc()
	} else {
		s.resourceCount.WithLabelValues("flavors").Set(float64(count))
	}

	s.resourceCount.Collect(ch)
	s.scrapeErrors.Collect(ch)
}

func (s *Store) Describe(ch chan<- *prometheus.Desc) {
	s.resourceCount.Describe(ch)
	s.scrapeErrors.Describe(ch)
}
