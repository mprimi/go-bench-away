package reports

type ReportSpec struct {
	Title    string              `json:"title"`
	JobIds   []string            `json:"jobs"`
	Sections []ReportSectionSpec `json:"sections"`
}

type ReportSectionSpec struct {
	Title               string `json:"title"`
	Metric              string `json:"metric"`
	Type                string `json:"type"`
	BenchmarkFilterExpr string `json:"benchmark_filter"`
	ResultsTable        bool   `json:"results_table"`
	HiddenResultsTable  bool   `json:"hidden_results_table"`
}
