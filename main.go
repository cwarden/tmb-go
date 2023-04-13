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
	epoch := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	datesWithEpoch := append([]time.Time{epoch}, dates...)
	sort.Slice(datesWithEpoch, func(i, j int) bool {
		return datesWithEpoch[i].After(datesWithEpoch[j])
	})

	var keepersRec func([]time.Time, []*KeepSpec) []time.Time
	keepersRec = func(dates []time.Time, specs []*KeepSpec) []time.Time {
		if len(dates) <= 1 {
			// Last remaining date is the epoch; discard it
			return []time.Time{}
		}
		if len(specs) == 0 {
			return []time.Time{}
		}

		spec := specs[0]
		if spec.howMany == 0 {
			// Use the next spec
			return keepersRec(dates, specs[1:])
		}

		first, second := dates[0], dates[1]
		if first.Sub(second) < spec.minDiff {
			// second is too close to first; it can be deleted
			return keepersRec(dates[0:1+copy(dates[1:], dates[2:])], specs)
		}

		spec.howMany--
		return append([]time.Time{first}, keepersRec(dates[1:], specs)...)
	}

	return keepersRec(datesWithEpoch, specs)
}

func toDelete(dates []time.Time, specs []*KeepSpec) []time.Time {
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
		{howMany: 12, minDiff: 30 * 24 * time.Hour}, // 12 months
		{howMany: 4, minDiff: 7 * 24 * time.Hour},   // 4 weeks
		{howMany: 7, minDiff: 18 * time.Hour},       // at least 18 hours between daily backups
	}
	sort.Slice(specs, func(i, j int) bool {
		return specs[i].minDiff < specs[j].minDiff
	})

	extraBackups := toDelete(validDates(dates), specs)

	for _, backup := range extraBackups {
		fmt.Println(backup.Format("2006-01-02T15:04:05"))
	}
}
