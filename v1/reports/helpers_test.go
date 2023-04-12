package reports

import (
	"reflect"
	"testing"

	"github.com/mprimi/go-bench-away/v1/core"
)

func TestCreateLabels(t *testing.T) {

	j1 := &core.JobRecord{
		Id: "job1",
		Parameters: core.JobParameters{
			GitRef: "main",
		},
		SHA:       "05b8c30eed0a8c7d87d1b22c3f5c6ef77eece297",
		GoVersion: "go 1.19.3",
	}

	j2 := &core.JobRecord{
		Id: "job2",
		Parameters: core.JobParameters{
			GitRef: "399201d5760dfcc1f9193b7c8df14da8dc4a8a20",
		},
		SHA:       "399201d5760dfcc1f9193b7c8df14da8dc4a8a20",
		GoVersion: "go 1.19.3",
	}

	j3 := &core.JobRecord{
		Id: "job3",
		Parameters: core.JobParameters{
			GitRef: "main",
		},
		SHA:       "7b84175711fd460bee2c217174d8582332e82e1a",
		GoVersion: "go 1.19.3",
	}

	// Same as j3, with different go version
	j4 := &core.JobRecord{
		Id: "job4",
		Parameters: core.JobParameters{
			GitRef: "main",
		},
		SHA:       "7b84175711fd460bee2c217174d8582332e82e1a",
		GoVersion: "go 1.19.6",
	}

	// Same as j4 (except for job id)
	j5 := &core.JobRecord{
		Id: "job5",
		Parameters: core.JobParameters{
			GitRef: "main",
		},
		SHA:       "7b84175711fd460bee2c217174d8582332e82e1a",
		GoVersion: "go 1.19.6",
	}

	type TestCase struct {
		jobs           []*core.JobRecord
		expectedLabels []string
		description    string
	}

	testCases := []TestCase{
		{
			[]*core.JobRecord{j1},
			[]string{"main"},
			"One job with GitRef",
		},
		{
			[]*core.JobRecord{j2},
			[]string{"399201d"},
			"One job with SHA GitRef",
		},
		{
			[]*core.JobRecord{j1, j2},
			[]string{"main", "399201d"},
			"Two jobs with different GitRef",
		},
		{
			[]*core.JobRecord{j1, j2, j3},
			[]string{"main [05b8c30]", "399201d", "main [7b84175]"},
			"Three jobs two of which share GitRef",
		},
		{
			[]*core.JobRecord{j3, j4},
			[]string{"main [go 1.19.3]", "main [go 1.19.6]"},
			"Two jobs with same GitRef and SHA, but different go version",
		},
		{
			[]*core.JobRecord{j4, j5},
			[]string{"job4", "job5"},
			"Two identical jobs",
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.description,
			func(t *testing.T) {
				labels := createJobLabels(testCase.jobs)
				if !reflect.DeepEqual(labels, testCase.expectedLabels) {
					t.Fatalf("Expected: %v actual: %v", testCase.expectedLabels, labels)
				}
			},
		)
	}
}

func TestCreateLabelsPanic(t *testing.T) {

	job := &core.JobRecord{
		Id: "job1",
		Parameters: core.JobParameters{
			GitRef: "main",
		},
		SHA:       "05b8c30eed0a8c7d87d1b22c3f5c6ef77eece297",
		GoVersion: "go 1.19.3",
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("Expected panic")
		}
	}()

	createJobLabels([]*core.JobRecord{job, job})
}
