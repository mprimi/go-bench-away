package reports

import (
	"github.com/mprimi/go-bench-away/v1/core"
)

type jobsTableSection struct {
	baseSection
	Jobs []*core.JobRecord
}

func (s *jobsTableSection) fillData(dt *dataTableImpl) error {
	s.Jobs = dt.jobs
	return nil
}

func JobsTable() SectionConfig {
	return &jobsTableSection{
		baseSection: baseSection{
			Type:  "jobs_table",
			Title: "Jobs",
		},
	}
}
