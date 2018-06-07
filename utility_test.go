package walg_test

import (
	"github.com/wal-g/wal-g"
	"sort"
	"testing"
	"time"
)

var times = []struct {
	input walg.BackupTime
}{
	{walg.BackupTime{
		Name:        "second",
		Time:        time.Date(2017, 2, 2, 30, 48, 39, 651387233, time.UTC),
		WalFileName: "",
	}},
	{walg.BackupTime{
		Name:        "fourth",
		Time:        time.Date(2009, 2, 27, 20, 8, 33, 651387235, time.UTC),
		WalFileName: "",
	}},
	{walg.BackupTime{
		Name:        "fifth",
		Time:        time.Date(2008, 11, 20, 16, 34, 58, 651387232, time.UTC),
		WalFileName: "",
	}},
	{walg.BackupTime{
		Name:        "first",
		Time:        time.Date(2020, 11, 31, 20, 3, 58, 651387237, time.UTC),
		WalFileName: "",
	}},
	{walg.BackupTime{
		Name:        "third",
		Time:        time.Date(2009, 3, 13, 4, 2, 42, 651387234, time.UTC),
		WalFileName: "",
	}},
}

func TestSortLatestTime(t *testing.T) {
	correct := [5]string{"first", "second", "third", "fourth", "fifth"}
	sortTimes := make([]walg.BackupTime, 5)

	for i, val := range times {
		sortTimes[i] = val.input
	}

	sort.Sort(walg.TimeSlice(sortTimes))

	for i, val := range sortTimes {
		if val.Name != correct[i] {
			t.Errorf("utility: Sort times expected %v as %s but got %s instead", val.Time, correct[i], val.Name)
		}
	}
}
