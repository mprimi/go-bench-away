package reports

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestReportSpec_LoadFile(t *testing.T) {

	const TestConfigsDir = "testconfig"

	var emptyReportCfg ReportConfig
	emptyReportCfg.AddSections(JobsTable())

	resetChartId()
	var validReportCfg1 ReportConfig
	validReportCfg1.AddSections(JobsTable())
	validReportCfg1.Title = "Trend, Bars, Delta, no filters"
	validReportCfg1.AddSections(
		TrendChart("Time/Op Trend", TimeOp, ""),
		ResultsTable(TimeOp, "", true),
		HorizontalBarChart("Speed measurements", Speed, ""),
		ResultsTable(Speed, "", true),
		HorizontalDeltaChart("Speed delta", Speed, ""),
		ResultsDeltaTable(Speed, "", true),
	)

	resetChartId()
	var validReportCfg2 ReportConfig
	validReportCfg2.AddSections(JobsTable())
	validReportCfg2.Title = "Trend, Bars, Delta, with filters"
	validReportCfg2.AddSections(
		TrendChart("Time/Op Trend", TimeOp, "foo.*"),
		ResultsTable(TimeOp, "foo.*", true),
		TrendChart("Op/s Trend", OpsPerSec, "foo.*"),
		ResultsTable(OpsPerSec, "foo.*", true),
		TrendChart("Msg/s Trend", MsgPerSec, "foo.*"),
		ResultsTable(MsgPerSec, "foo.*", true),
		HorizontalBarChart("Speed measurements", Speed, "bar.*"),
		ResultsTable(Speed, "bar.*", true),
		HorizontalDeltaChart("Speed delta", Speed, "baz.*"),
		ResultsDeltaTable(Speed, "baz.*", true),
	)

	testCases := []struct {
		specPath          string
		expectedReportCfg *ReportConfig
	}{
		{
			"report_spec_empty_1.json",
			&emptyReportCfg,
		},
		{
			"report_spec_empty_2.json",
			&emptyReportCfg,
		},
		{
			"report_spec_valid_1.json",
			&validReportCfg1,
		},
		{
			"report_spec_valid_2.json",
			&validReportCfg2,
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.specPath,
			func(t *testing.T) {
				resetChartId()

				specPath := filepath.Join(TestConfigsDir, testCase.specPath)

				var spec ReportSpec
				err := spec.LoadFile(specPath)
				if err != nil {
					t.Fatal(err)
				}

				reportCfg := &ReportConfig{}
				err = spec.ConfigureReport(reportCfg)
				if err != nil {
					t.Fatal(err)
				}

				if len(reportCfg.sections) != len(testCase.expectedReportCfg.sections) {
					t.Fatalf(
						"Sections mismatch, expected: %d, actual: %d",
						testCase.expectedReportCfg.sections,
						reportCfg.sections,
					)
				}

				for i, section := range testCase.expectedReportCfg.sections {
					if !reflect.DeepEqual(section, reportCfg.sections[i]) {
						t.Fatalf("sections[%d] mismatch\nExp:%+v\nAct:%+v", i, section, reportCfg.sections[i])
					}
				}

				if !reflect.DeepEqual(reportCfg, testCase.expectedReportCfg) {
					t.Fatalf(
						"Report configuration mismatch\nExpected: %+v\nActual:   %+v",
						testCase.expectedReportCfg,
						reportCfg,
					)
				}
			},
		)
	}
}

func TestReportSpec_LoadError(t *testing.T) {
	const TestConfigsDir = "testconfig"

	testCases := []struct {
		specPath      string
		expectedError string
	}{
		{
			"report_spec_invalid_0.json",
			"no such file or directory",
		},
		{
			"report_spec_invalid_1.json",
			"invalid character",
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.specPath,
			func(t *testing.T) {
				specPath := filepath.Join(TestConfigsDir, testCase.specPath)
				var spec ReportSpec
				err := spec.LoadFile(specPath)
				if err == nil {
					t.Fatalf("Expecting an error when loading %s", specPath)
				}
				errorString := fmt.Sprintf("%s", err)
				if !strings.Contains(errorString, testCase.expectedError) {
					t.Fatalf("Expected error: \"%s\", actual: \"%s\"", testCase.expectedError, errorString)
				}
			},
		)
	}
}

func TestReportSpec_ConfigureError(t *testing.T) {
	const TestConfigsDir = "testconfig"

	testCases := []struct {
		specPath      string
		expectedError string
	}{
		{
			"report_spec_invalid_2.json",
			"unknown metric",
		},
		{
			"report_spec_invalid_3.json",
			"unknown section",
		},
	}

	for _, testCase := range testCases {
		t.Run(
			testCase.specPath,
			func(t *testing.T) {
				specPath := filepath.Join(TestConfigsDir, testCase.specPath)
				var spec ReportSpec
				err := spec.LoadFile(specPath)
				if err != nil {
					t.Fatalf("Unexpected load error: %s", err)
				}

				var cfg ReportConfig
				err = spec.ConfigureReport(&cfg)
				if err == nil {
					t.Fatalf("Expecting an error when loading %s", specPath)
				}
				errorString := fmt.Sprintf("%s", err)
				if !strings.Contains(errorString, testCase.expectedError) {
					t.Fatalf("Expected error: \"%s\", actual: \"%s\"", testCase.expectedError, errorString)
				}
			},
		)
	}
}
