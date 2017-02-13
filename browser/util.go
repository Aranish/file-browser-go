package browser;

import (
	"sync";
	"os";
	"path/filepath";
	"gopkg.in/vmihailenco/msgpack.v2";
)

// Check to make sure msgpack output does not get
// interleaved

var outLock sync.Mutex;
func WriteOut(encoder *msgpack.Encoder , res interface{}){
	outLock.Lock();
	encoder.Encode(res);
	outLock.Unlock();
}

// Locking paths to make sure there's no interference

var isLocked map[string]bool = make(map[string]bool);

func LockPath (path string) {
	isLocked[path] = true;
}

func UnlockPath (path string) {
	isLocked[path] = false;
}

func IsLocked (path string) bool{
	if isLocked[path] {
		return true;
	}
	dir, f := filepath.Split(path);
	for f != "" {
		if isLocked[dir] {
			return true;
		}
		dir, f = filepath.Split(dir);
	}
	return false;
}

// On exit make browser wait until every copy operation
// is complete
var op sync.WaitGroup;

func OpAdd () {
	op.Add(1);
}

func OpDone () {
	op.Done();
}

// Utility functions
func WaitForOperationsToComplete() {
	op.Wait();
}

func IsDir (dir string) bool {
	file, err := os.Open(dir);
	defer file.Close();
	if err != nil {
		return false;
	}
	finfo, err := file.Stat();
	if err != nil {
		return false;
	}
	return finfo.IsDir();
}

func ValidateDirPath (dir *string) bool {
	*dir = filepath.Clean(*dir);
	if !filepath.IsAbs(*dir) {
		return false;
	}
	return true;
}
