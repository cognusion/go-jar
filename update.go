package jar

import (
	"github.com/inconshreveable/go-update"

	"github.com/cognusion/go-jar/aws"

	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
)

const (
	// ErrUpdateConfigS3NoEC2 is returned when the s3 updatepath is set, but ec2 is not.
	ErrUpdateConfigS3NoEC2 = Error("s3 updatepath set, but ec2 is false")

	// ErrUpdateConfigEmptyURL is returned when the updatepath is empty
	ErrUpdateConfigEmptyURL = Error("update url is empty, not updating")
)

var (
	// RestartSelf is a niladic that will trigger a graceful restart of this process
	RestartSelf func()
	// IntSelf is a niladic that will trigger an interrupt of this process
	IntSelf func()
	// KillSelf is a niladic that will trigger a graceful shutdown of this process
	KillSelf func()
)

func init() {
	// Link our Finishers
	Finishers["update"] = Update
	Finishers["restart"] = Restart

	// Ensure we only restart ourselves once
	// why 10? small buffer so multiple requests close together hit the channel
	// and don't block the caller
	signalChan := make(chan os.Signal, 10)

	// These functions are used by handlers to trigger signals to the running
	// process.
	RestartSelf = func() { signalChan <- syscall.SIGUSR2 }
	IntSelf = func() { signalChan <- syscall.SIGINT }
	KillSelf = func() { signalChan <- syscall.SIGKILL }

	// We mirror the signal interception of grace, so we can properly clean up
	ch := make(chan os.Signal, 10)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)

	go func() {
		for {
			select {
			case s := <-signalChan:
				DebugOut.Printf("%s signaled.", s.String())
				p, err := os.FindProcess(os.Getpid())
				if err != nil {
					ErrorOut.Printf("Error finding process '%d': %s\n", os.Getpid(), err)
				}
				p.Signal(s)
			case s := <-ch:
				switch s {
				case syscall.SIGINT, syscall.SIGTERM:
					StopFuncs.Call()
					signal.Stop(ch)
					return
				case syscall.SIGUSR2:
					StopFuncs.Call()
				}
			}
		}
	}()

	ConfigValidations["s3updatepath"] = func() error {
		if up := Conf.GetString(ConfigUpdatePath); up != "" && !Conf.GetBool(ConfigEC2) {
			// Update requires EC2, for now
			return ErrUpdateConfigS3NoEC2
		}
		return nil
	}
}

// Update signals the updater to update itself
func Update(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Updating...\n")
	err := handleUpdate(Conf.GetString(ConfigUpdatePath), Conf.GetBool(ConfigHotUpdate))
	if err != nil {
		fmt.Fprintf(w, "Error: %s\n", err)
	} else {
		fmt.Fprint(w, "Done\n")
	}
}

// Restart signals the server to restart itself
func Restart(w http.ResponseWriter, r *http.Request) {
	RestartSelf()
	fmt.Fprint(w, "Done\n")
}

// handleUpdate grabs a zip from S3, downloads it, unzips it, updates the binary, and if restart, signals
// ourself to to gracefully restart
func handleUpdate(url string, restart bool) error {
	if url == "" {
		ErrorOut.Printf("Update requested, but empty URL provided\n")
		return ErrUpdateConfigEmptyURL
	}

	// If the Ec2Session isn't initialized, we cannot grab it
	if Ec2Session == nil {
		return ErrNoSession
	}

	// Break the URL into parts
	bucket, bucketPath, filename := aws.S3urlToParts(url)

	baseName := path.Base(os.Args[0])

	// Set all the paths
	tmp := fmt.Sprintf("%s/%s-%d", Conf.GetString(ConfigTempFolder), baseName, os.Getpid())

	// Clear it out now, and later
	os.RemoveAll(tmp)
	defer os.RemoveAll(tmp)

	// Mkdir
	err := os.MkdirAll(tmp, 0700)
	if err != nil {
		ErrorOut.Printf("Error creating folder(s) '%s': %s\n", tmp, err)
		return err
	}

	tfile := fmt.Sprintf("%s/%s", tmp, filename)
	tzfolder := fmt.Sprintf("%s/%s", tmp, "jarupdate")

	// grab the file and save it
	_, err = Ec2Session.BucketToFile(bucket, bucketPath, tfile)
	if err != nil {
		ErrorOut.Printf("Error downloading '%s' from '%s' to '%s': %s\n", bucketPath, bucket, tfile, err)
		return err
	}

	// Unzip the file
	err = Unzip(tfile, tzfolder)
	if err != nil {
		ErrorOut.Printf("Error unzipping zip file '%s' to '%s': %s\n", tfile, tzfolder, err)
		return err
	}

	// Open the result
	f, err := os.Open(fmt.Sprintf("%s/%s", tzfolder, baseName))
	if err != nil {
		ErrorOut.Printf("Error opening temp file '%s': %s\n", fmt.Sprintf("%s/%s", tzfolder, baseName), err)
		return err
	}
	defer f.Close()

	// Update our binary
	err = update.Apply(f, update.Options{})
	if err != nil {
		if rerr := update.RollbackError(err); rerr != nil {
			ErrorOut.Printf("Failed to rollback from bad update: %v", rerr)
		}
		return err
	}

	// Restart ourself, maybe
	if restart {
		RestartSelf()
	}

	return nil
}

// Unzip takes a source zip, and a destination folder, and unzips source into dest,
// returning an error if appropriate
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
