package main

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

// Helper function to parse dates
func mustParseDate(t *testing.T, dateStr string) time.Time {
	date, err := time.Parse("2006-01-02T15:04:05", dateStr)
	if err != nil {
		t.Fatalf("failed to parse date: %s", dateStr)
	}
	return date
}

// Table-based test for the toDelete function
func TestToDelete(t *testing.T) {
	tests := []struct {
		name        string
		dates       []string
		specs       []*KeepSpec
		expectedDel []string
	}{
		{
			name: "Basic case with yearly, monthly, weekly, and daily backups",
			dates: []string{
				"2021-01-01T12:00:00",
				"2022-01-01T12:00:00",
				"2023-01-01T12:00:00",
				"2024-01-01T12:00:00",
				"2024-01-06T12:00:00",
				"2024-01-07T12:00:00",
				"2024-01-08T12:00:00",
				"2024-06-01T12:00:00",
				"2024-07-01T12:00:00",
				"2024-08-01T12:00:00",
			},
			specs: []*KeepSpec{
				{howMany: 3, minDiff: 365 * 24 * time.Hour}, // yearly
				{howMany: 2, minDiff: 30 * 24 * time.Hour},  // monthly
				{howMany: 2, minDiff: 7 * 24 * time.Hour},   // weekly
			},
			expectedDel: []string{
				"2024-01-01T12:00:00",
				"2024-01-06T12:00:00",
				"2024-01-07T12:00:00",
			},
		},
		{
			name: "All dates within minDiff, only keep latest",
			dates: []string{
				"2024-01-01T12:00:00",
				"2024-01-01T11:00:00",
				"2024-01-01T10:00:00",
				"2024-01-01T09:00:00",
			},
			specs: []*KeepSpec{
				{howMany: 1, minDiff: 12 * time.Hour}, // daily backups
			},
			expectedDel: []string{
				"2024-01-01T09:00:00",
				"2024-01-01T10:00:00",
				"2024-01-01T11:00:00",
			},
		},
		{
			name: "Keep all dates when they satisfy minDiff",
			dates: []string{
				"2024-01-01T12:00:00",
				"2023-01-01T12:00:00",
				"2022-01-01T12:00:00",
			},
			specs: []*KeepSpec{
				{howMany: 3, minDiff: 365 * 24 * time.Hour}, // yearly
			},
			expectedDel: []string{}, // No deletions, all satisfy minDiff
		},
		{
			name: "Keep oldest with 3 yearly",
			dates: []string{
				"2022-01-13T12:00:00",
				"2022-01-12T12:00:00",
				"2022-01-11T12:00:00",
				"2022-01-10T12:00:00",
				"2022-01-09T12:00:00",
				"2022-01-08T12:00:00",
				"2022-01-07T12:00:00",
				"2022-01-06T12:00:00",
				"2022-01-05T12:00:00",
				"2022-01-04T12:00:00",
				"2022-01-03T12:00:00",
				"2022-01-02T12:00:00",
				"2022-01-01T12:00:00", // Keep this one, even though it's not old enough, for future yearly backup
			},
			specs: []*KeepSpec{
				{howMany: 3, minDiff: 365 * 24 * time.Hour}, // yearly
				{howMany: 5, minDiff: 2 * 24 * time.Hour},   // bi-daily
			},
			expectedDel: []string{
				"2022-01-12T12:00:00",
				"2022-01-10T12:00:00",
				"2022-01-08T12:00:00",
				"2022-01-06T12:00:00",
				"2022-01-04T12:00:00",
				"2022-01-03T12:00:00",
				"2022-01-02T12:00:00",
			},
		},
		{
			name: "Keep oldest",
			dates: []string{
				"2022-01-13T12:00:00",
				"2022-01-12T12:00:00",
				"2022-01-11T12:00:00",
				"2022-01-10T12:00:00",
				"2022-01-09T12:00:00",
				"2022-01-08T12:00:00",
				"2022-01-07T12:00:00",
				"2022-01-06T12:00:00",
				"2022-01-05T12:00:00",
				"2022-01-04T12:00:00",
				"2022-01-03T12:00:00",
				"2022-01-02T12:00:00",
				"2022-01-01T12:00:00", // Keep this one, even though it's not old enough, for future yearly backup
			},
			specs: []*KeepSpec{
				{howMany: 1, minDiff: 365 * 24 * time.Hour}, // yearly
				{howMany: 5, minDiff: 2 * 24 * time.Hour},   // bi-daily
			},
			expectedDel: []string{
				"2022-01-12T12:00:00",
				"2022-01-10T12:00:00",
				"2022-01-08T12:00:00",
				"2022-01-06T12:00:00",
				"2022-01-04T12:00:00",
				"2022-01-03T12:00:00",
				"2022-01-02T12:00:00",
			},
		},
		{
			name: "Specs can be in any order",
			dates: []string{
				"2021-01-01T12:00:00",
				"2022-01-01T12:00:00",
				"2023-01-01T12:00:00",
				"2024-01-01T12:00:00",
				"2024-01-06T12:00:00",
				"2024-01-07T12:00:00",
				"2024-01-08T12:00:00",
				"2024-06-01T12:00:00",
				"2024-07-01T12:00:00",
				"2024-08-01T12:00:00",
			},
			specs: []*KeepSpec{
				{howMany: 2, minDiff: 30 * 24 * time.Hour},  // monthly
				{howMany: 3, minDiff: 365 * 24 * time.Hour}, // yearly
				{howMany: 2, minDiff: 7 * 24 * time.Hour},   // weekly
			},
			expectedDel: []string{
				"2024-01-01T12:00:00",
				"2024-01-06T12:00:00",
				"2024-01-07T12:00:00",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert string dates to time.Time
			dates := make([]time.Time, len(tt.dates))
			for i, dateStr := range tt.dates {
				dates[i] = mustParseDate(t, dateStr)
			}

			// Convert string expectedDel to time.Time
			expectedDel := make([]time.Time, len(tt.expectedDel))
			for i, dateStr := range tt.expectedDel {
				expectedDel[i] = mustParseDate(t, dateStr)
			}

			// Run the function
			actualDel := toDelete(dates, tt.specs)

			sort.Slice(expectedDel, func(i, j int) bool {
				return expectedDel[i].After(expectedDel[j])
			})

			// Compare the result with expected output
			if !reflect.DeepEqual(actualDel, expectedDel) {
				t.Errorf("\nactual:   %v\nexpected: %v", actualDel, expectedDel)
			}
		})
	}
}
