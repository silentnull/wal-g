package walg

import "testing"

func TestLSNParse(t *testing.T) {
	lsn, err := ParseLsn("2/E5000028")
	if err != nil {
		t.Fatal(err)
	}
	if lsn != 0x2E5000028 {
		t.Fatal("LSN was not parsed correctly")
	}
}

func TestNextWALFileName(t *testing.T) {
	nextname, err := NextWALFileName("000000010000000000000051")
	if err != nil || nextname != "000000010000000000000052" {
		t.Fatal("TestNextWALFileName 000000010000000000000051 failed")
	}

	nextname, err = NextWALFileName("00000001000000000000005F")
	if err != nil || nextname != "000000010000000000000060" {
		t.Fatal("TestNextWALFileName 00000001000000000000005F failed")
	}

	nextname, err = NextWALFileName("0000000100000001000000FF")
	if err != nil || nextname != "000000010000000200000000" {
		t.Fatal("TestNextWALFileName 0000000100000001000000FF failed")
	}

	_, err = NextWALFileName("0000000100000001000001FF")
	if err == nil {
		t.Fatal("TestNextWALFileName 0000000100000001000001FF did not failed")
	}

	_, err = NextWALFileName("00000001000ZZ001000000FF")
	if err == nil {
		t.Fatal("TestNextWALFileName 00000001000ZZ001000001FF did not failed")
	}

	_, err = NextWALFileName("00000001000001000000FF")
	if err == nil {
		t.Fatal("TestNextWALFileName 00000001000001000001FF did not failed")
	}

	_, err = NextWALFileName("asdfasdf")
	if err == nil {
		t.Fatal("TestNextWALFileName asdfasdf did not failed")
	}
}

func TestPrefetchLocation(t *testing.T) {
	prefetchLocation, runningLocation, runningFile, fetchedFile := getPrefetchLocations("/var/pgdata/xlog/", "000000010000000000000051")
	if prefetchLocation != "/var/pgdata/xlog/.wal-g/prefetch" ||
		runningLocation != "/var/pgdata/xlog/.wal-g/prefetch/running" ||
		runningFile != "/var/pgdata/xlog/.wal-g/prefetch/running/000000010000000000000051" ||
		fetchedFile != "/var/pgdata/xlog/.wal-g/prefetch/000000010000000000000051" {
		t.Fatal("TestPrefetchLocation failed")
	}
}
