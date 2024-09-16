package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

type Seconds int64

type KeepSpec struct {
	howMany int
	minDiff time.Duration
}

func (s *KeepSpec) String() string {
	return fmt.Sprintf("%+v\n", *s)
}

func validDates(dates []time.Time) []time.Time {
	valid := make([]time.Time, 0)
	for _, date := range dates {
		if !date.IsZero() {
			valid = append(valid, date)
		}
	}
	return valid
}

func keepers(dates []time.Time, specs []*KeepSpec) []time.Time {
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].After(dates[j])
	})

	var keepersRec func(time.Time, []time.Time, []*KeepSpec) []time.Time
	keepersRec = func(lastDate time.Time, dates []time.Time, specs []*KeepSpec) []time.Time {
		if len(dates) == 0 || len(specs) == 0 {
			return []time.Time{}
		}

		spec := specs[0]
		if spec.howMany == 0 {
			// Use the next spec
			return keepersRec(lastDate, dates, specs[1:])
		}

		first := dates[0]
		if !lastDate.IsZero() && len(dates) > 1 && lastDate.Sub(first) < spec.minDiff {
			// date is too close; it can be deleted unless it's the
			// last one
			return keepersRec(lastDate, dates[1:], specs)
		}

		spec.howMany--
		return append([]time.Time{first}, keepersRec(first, dates[1:], specs)...)
	}

	var s time.Time
	return keepersRec(s, dates, specs)
}

func toDelete(dates []time.Time, specs []*KeepSpec) []time.Time {
	sort.Slice(specs, func(i, j int) bool {
		return specs[i].minDiff < specs[j].minDiff
	})

	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	keep := keepers(dates, specs)
	toDel := make([]time.Time, 0)
	keepMap := make(map[string]bool)

	for _, k := range keep {
		keepMap[k.Format("2006-01-02T15:04:05")] = true
	}

	for _, date := range dates {
		if !keepMap[date.Format("2006-01-02T15:04:05")] {
			toDel = append(toDel, date)
		}
	}

	return toDel
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	var dates []time.Time

	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		date, err := time.Parse("2006-01-02T15:04:05", strings.TrimSpace(input))
		if err == nil {
			dates = append(dates, date)
		}
	}

	specs := []*KeepSpec{
		{howMany: 10, minDiff: 365.25 * 24 * time.Hour}, // 10 years
		{howMany: 12, minDiff: 30 * 24 * time.Hour},     // 12 months
		{howMany: 4, minDiff: 7 * 24 * time.Hour},       // 4 weeks
		{howMany: 7, minDiff: 18 * time.Hour},           // at least 18 hours between daily backups
	}
	extraBackups := toDelete(validDates(dates), specs)

	for _, backup := range extraBackups {
		fmt.Println(backup.Format("2006-01-02T15:04:05"))
	}
}
