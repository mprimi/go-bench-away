package web

import (
	"testing"
)

func TestJobResourcesRegexp(t *testing.T) {
	var expectNoMatchCases = []string{
		"",
		"/",
		"/job",
		"/job//log",
		"/job/_2fb41f25-7e17-4383-9e08-8ab115152db2/log",    // Non-hexadecimal char in JobId UUID
		"/job/2fb41f2-7e17-4383-9e08-8ab115152db2/log",      // Short 1 char in 1st group
		"/job/2fb41f25-717-4383-9e08-8ab115152db2/log",      // Short 1 char in middle group
		"/job/2fb41f25-7e17-483-9e08-8ab115152db2/log",      // Short 1 char in middle group
		"/job/2fb41f25-7e17-4383-908-8ab115152db2/log",      // Short 1 char in middle group
		"/job/2fb41f25-7e17-4383-9e08-ab115152db2/log",      // Short 1 char in last group
		"/job/2fb41f257e1743839e088ab115152db2/log",         // Missing dashes
		"/job/2fb41f25-7e17-4383-9e08-8ab115152db2/blah",    // Invalid resource type
		"/job/2fb41f25-7e17-4383-9e08-8ab115152db2/log/foo", // Extra path component
	}

	for _, s := range expectNoMatchCases {
		if jobResourceRegexp.MatchString(s) {
			t.Errorf("Should not have matched, but did: '%s'", s)
		}
	}

	var expectMatchCases = []struct {
		input            string
		expectedJobId    string
		expectedResource string
	}{
		{
			input:            "/job/2fb41f25-7e17-4383-9e08-8ab115152db2/log",
			expectedJobId:    "2fb41f25-7e17-4383-9e08-8ab115152db2",
			expectedResource: "log",
		},
		{
			input:            "/job/2fb41f25-7e17-4383-9e08-8ab115152db2/log/",
			expectedJobId:    "2fb41f25-7e17-4383-9e08-8ab115152db2",
			expectedResource: "log",
		},
		{
			input:            "/job/2fb41f25-7e17-4383-9e08-8ab115152db2/script/",
			expectedJobId:    "2fb41f25-7e17-4383-9e08-8ab115152db2",
			expectedResource: "script",
		},
	}

	for _, tc := range expectMatchCases {
		if !jobResourceRegexp.MatchString(tc.input) {
			t.Errorf("Should have matched, but didn't: '%s'", tc.input)
		}
	}
}
