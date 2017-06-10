// Update protocol:
//
//   GET hk.heroku.com/hk/linux-amd64.json
//
//   200 ok
//   {
//       "Version": "2",
//       "Sha256": "..." // base64
//   }
//
// then
//
//   GET hkpatch.s3.amazonaws.com/hk/1/2/linux-amd64
//
//   200 ok
//   [bsdiff data]
//
// or
//
//   GET hkdist.s3.amazonaws.com/hk/2/linux-amd64.gz
//
//   200 ok
//   [gzipped executable data]
//
//
package selfupdate

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/kardianos/osext"
	"github.com/kr/binarydist"
	"gopkg.in/inconshreveable/go-update.v0"
)

const (
	upcktimePath = "cktime.json"
	plat         = runtime.GOOS + "-" + runtime.GOARCH
)

const devValidTime = 7 * 24 * time.Hour

var ErrHashMismatch = errors.New("new file hash mismatch after patch")
var up = update.New()
var defaultHTTPRequester = HTTPRequester{}

type updateTimeRecord struct {
	NextUpdate time.Time `json:"nextUpdate"`
	Message    string    `json:"lastError"`
}

// Updater is the configuration and runtime data for doing an update.
//
// Note that ApiURL, BinURL and DiffURL should have the same value if all files are available at the same location.
//
// Example:
//
//  updater := &selfupdate.Updater{
//  	CurrentVersion: version,
//  	ApiURL:         "http://updates.yourdomain.com/",
//  	BinURL:         "http://updates.yourdownmain.com/",
//  	DiffURL:        "http://updates.yourdomain.com/",
//  	Dir:            "update/",
//  	CmdName:        "myapp", // app name
//  }
//  if updater != nil {
//  	go updater.BackgroundRun()
//  }
type Updater struct {
	CurrentVersion string    // Currently running version.
	ApiURL         string    // Base URL for API requests (json files).
	CmdName        string    // Command name is appended to the ApiURL like http://apiurl/CmdName/. This represents one binary.
	BinURL         string    // Base URL for full binary downloads.
	DiffURL        string    // Base URL for diff downloads.
	Dir            string    // Directory to store selfupdate state.
	ForceCheck     bool      // Check for update regardless of cktime timestamp
	Requester      Requester //Optional parameter to override existing http request handler
	Info           struct {
		Version string
		Sha256  []byte
	}
}

func (u *Updater) getExecRelativeDir(dir string) string {
	filename, _ := osext.Executable()
	path := filepath.Join(filepath.Dir(filename), dir)
	return path
}

// BackgroundRun starts the update check and apply cycle.
func (u *Updater) BackgroundRun(update bool) error {
	if _, err := os.Stat(u.Dir); err != nil && os.IsNotExist(err) {
		os.MkdirAll(u.Dir, 0700)
	}
	if update {
		if err := up.CanUpdate(); err != nil {
			// fail
			return err
		}
		//self, err := osext.Executable()
		//if err != nil {
		// fail update, couldn't figure out path to self
		//return
		//}
		// TODO(bgentry): logger isn't on Windows. Replace w/ proper error reports.
		if err := u.update(); err != nil {
			return err
		}
	}
	return nil
}

func (u *Updater) WantUpdate() bool {
	path := filepath.Join(u.Dir, upcktimePath)
	if u.CurrentVersion == "dev" || (!u.ForceCheck && readTime(path).After(time.Now())) {
		//log.Println("not update")
		return false
	}
	if u.fetchInfo() != nil {
		return false
	}
	if u.Info.Version == u.CurrentVersion {
		return false
	}

	return true
}

func (u *Updater) WriteTimeWithError(err error) {
	path := filepath.Join(u.Dir, upcktimePath)
	msg := ""
	wait := 24*time.Hour + randDuration(24*time.Hour)
	if err != nil {
		wait = 1*time.Hour + randDuration(1*time.Hour)
		msg = fmt.Sprintf("binary update failed: %v", err)
	}
	writeTime(path, time.Now().Add(wait), msg)
}

func (u *Updater) update() error {
	path, err := osext.Executable()
	if err != nil {
		return err
	}
	old, err := os.Open(path)
	if err != nil {
		return err
	}
	defer old.Close()

	/*
		err = u.fetchInfo()
		if err != nil {
			return err
		}
		if u.Info.Version == u.CurrentVersion {
			return nil
		}
	*/
	bin, err := u.fetchAndVerifyPatch(old)
	if err != nil {
		if err == ErrHashMismatch {
			//log.Println("update: hash mismatch from patched binary")
		} else {
			if u.DiffURL != "" {
				//log.Println("update: patching binary,", err)
			}
		}

		bin, err = u.fetchAndVerifyFullBin()
		if err != nil {
			if err == ErrHashMismatch {
				//log.Println("update: hash mismatch from full binary")
			} else {
				//log.Println("update: fetching full binary,", err)
			}
			return err
		}
	}

	// close the old binary before installing because on windows
	// it can't be renamed if a handle to the file is still open
	old.Close()

	err, errRecover := up.FromStream(bytes.NewBuffer(bin))
	if errRecover != nil {
		return fmt.Errorf("update and recovery errors: %q %q", err, errRecover)
	}
	if err != nil {
		return err
	}
	return nil
}

func (u *Updater) fetchInfo() error {
	r, err := u.fetch(u.ApiURL + u.CmdName + "/" + plat + ".json")
	if err != nil {
		return err
	}
	defer r.Close()
	err = json.NewDecoder(r).Decode(&u.Info)
	if err != nil {
		return err
	}
	if len(u.Info.Sha256) != sha256.Size {
		return errors.New("bad cmd hash in info")
	}
	return nil
}

func (u *Updater) fetchAndVerifyPatch(old io.Reader) ([]byte, error) {
	bin, err := u.fetchAndApplyPatch(old)
	if err != nil {
		return nil, err
	}
	if !verifySha(bin, u.Info.Sha256) {
		return nil, ErrHashMismatch
	}
	return bin, nil
}

func (u *Updater) fetchAndApplyPatch(old io.Reader) ([]byte, error) {
	r, err := u.fetch(u.DiffURL + u.CmdName + "/" + u.CurrentVersion + "/" + u.Info.Version + "/" + plat)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var buf bytes.Buffer
	err = binarydist.Patch(old, &buf, r)
	return buf.Bytes(), err
}

func (u *Updater) fetchAndVerifyFullBin() ([]byte, error) {
	bin, err := u.fetchBin()
	if err != nil {
		return nil, err
	}
	verified := verifySha(bin, u.Info.Sha256)
	if !verified {
		return nil, ErrHashMismatch
	}
	return bin, nil
}

func (u *Updater) fetchBin() ([]byte, error) {
	r, err := u.fetch(u.BinURL + u.CmdName + "/" + u.Info.Version + "/" + plat + ".gz")
	if err != nil {
		return nil, err
	}
	defer r.Close()
	buf := new(bytes.Buffer)
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(buf, gz); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// returns a random duration in [0,n).
func randDuration(n time.Duration) time.Duration {
	return time.Duration(rand.Int63n(int64(n)))
}

func (u *Updater) fetch(url string) (io.ReadCloser, error) {
	if u.Requester == nil {
		return defaultHTTPRequester.Fetch(url)
	}

	readCloser, err := u.Requester.Fetch(url)
	if err != nil {
		return nil, err
	}

	if readCloser == nil {
		return nil, fmt.Errorf("Fetch was expected to return non-nil ReadCloser")
	}

	return readCloser, nil
}

func readTime(path string) time.Time {
	p, err := os.Open(path)
	if os.IsNotExist(err) {
		return time.Time{}
	}
	if err != nil {
		//log.Println(err)
		return time.Now().Add(1000 * time.Hour)
	}
	var update updateTimeRecord
	if err = json.NewDecoder(p).Decode(&update); err != nil {
		return time.Now().Add(1000 * time.Hour)
	}
	return update.NextUpdate
}

func verifySha(bin []byte, sha []byte) bool {
	h := sha256.New()
	h.Write(bin)
	return bytes.Equal(h.Sum(nil), sha)
}

func writeTime(path string, t time.Time, msg string) bool {
	data, err := json.MarshalIndent(updateTimeRecord{t, msg}, "", "\t")
	if err != nil {
		return false
	}
	err = ioutil.WriteFile(path, data, 0600)
	if err != nil {
		//log.Println(err)
		return false
	}
	return true
}
