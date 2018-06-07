package walg

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// BgUploader represents the state of concurrent WAL upload
type BgUploader struct {
	// pg_[wals|xlog]
	dir string

	// count of running gorutines
	parallelWorkers int32

	// usually defined by WALG_DOWNLOAD_CONCURRENCY
	maxParallelWorkers int32

	// waitgroup to handle Stop gracefully
	running sync.WaitGroup

	// every file is attempted only once
	started map[string]interface{}

	// uploading structure
	tu *TarUploader

	// to control amount of work done in one cycle of archive_comand
	totalUploaded int32

	mutex sync.Mutex

	pre    *Prefix
	verify bool
}

// Start up checking what's inside archive_status
func (u *BgUploader) Start(walFilePath string, maxParallelWorkers int32, tu *TarUploader, pre *Prefix, verify bool) {
	if maxParallelWorkers < 1 {
		return // Nothing to start
	}
	// prepare state
	u.tu = tu
	u.maxParallelWorkers = maxParallelWorkers
	u.dir = filepath.Dir(walFilePath)
	u.started = make(map[string]interface{})
	u.started[filepath.Base(walFilePath)+readySuffix] = walFilePath
	u.pre = pre
	u.verify = verify

	// This goroutine will spawn new if necessary
	go scanOnce(u)
}

// Stop pipeline
func (u *BgUploader) Stop() {
	for atomic.LoadInt32(&u.parallelWorkers) != 0 {
		time.Sleep(50 * time.Millisecond)
	} // Wait until noone works

	u.mutex.Lock()
	defer u.mutex.Unlock()
	atomic.StoreInt32(&u.maxParallelWorkers, 0) // stop new jobs
	u.running.Wait()                            // wait again for those how jumped to the closing door
}

var readySuffix = ".ready"
var archiveStatus = "archive_status"
var done = ".done"

func scanOnce(u *BgUploader) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	files, err := ioutil.ReadDir(filepath.Join(u.dir, archiveStatus))
	if err != nil {
		log.Print("Error of parallel upload: ", err)
		return
	}

	for _, f := range files {
		if haveNoSlots(u) {
			break
		}
		name := f.Name()
		if !strings.HasSuffix(name, readySuffix) {
			continue
		}
		if _, ok := u.started[name]; ok {
			continue
		}
		u.started[name] = name

		if shouldKeepScanning(u) {
			u.running.Add(1)
			atomic.AddInt32(&u.parallelWorkers, 1)
			go u.Upload(f)
		}
	}
}

func shouldKeepScanning(u *BgUploader) bool {
	return atomic.LoadInt32(&u.maxParallelWorkers) > 0 && atomic.LoadInt32(&u.totalUploaded) < 1024
}

func haveNoSlots(u *BgUploader) bool {
	return atomic.LoadInt32(&u.parallelWorkers) >= atomic.LoadInt32(&u.maxParallelWorkers)
}

// Upload one WAL file
func (u *BgUploader) Upload(info os.FileInfo) {
	walfilename := strings.TrimSuffix(info.Name(), readySuffix)
	UploadWALFile(u.tu.Clone(), filepath.Join(u.dir, walfilename), u.pre, u.verify)

	ready := filepath.Join(u.dir, archiveStatus, info.Name())
	done := filepath.Join(u.dir, archiveStatus, walfilename+done)
	err := os.Rename(ready, done)
	if err != nil {
		log.Print("Error renaming .ready to .done: ", err)
	}

	atomic.AddInt32(&u.totalUploaded, 1)

	scanOnce(u)
	atomic.AddInt32(&u.parallelWorkers, -1)

	u.running.Done()
}
