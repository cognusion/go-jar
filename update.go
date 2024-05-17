package jar

import (
	"github.com/cognusion/go-jar/aws"
	"github.com/inconshreveable/go-update"

	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

// Constants for configuration
const (
	ConfigHotUpdate  = ConfigKey("hotupdate")
	ConfigUpdatePath = ConfigKey("updatepath")
)

const (
	// ErrUpdateConfigS3NoAWS is returned when the s3 updatepath is set, but AWS is not.
	ErrUpdateConfigS3NoAWS = Error("s3 updatepath set, but AWS is not configured")

	// ErrUpdateConfigEmptyURL is returned when the updatepath is empty
	ErrUpdateConfigEmptyURL = Error("update url is empty, not updating")
)

func init() {
	// Link our Finishers
	Finishers["update"] = Update

	ConfigValidations[ConfigUpdatePath] = func() error {
		if up := Conf.GetString(ConfigUpdatePath); up != "" && AWSSession == nil {
			// Update requires AWS, for now
			return ErrUpdateConfigS3NoAWS
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

// handleUpdate grabs a zip from S3, downloads it, unzips it, updates the binary, and if restart, signals
// ourself to to gracefully restart
func handleUpdate(url string, restart bool) error {
	if url == "" {
		ErrorOut.Printf("Update requested, but empty URL provided\n")
		return ErrUpdateConfigEmptyURL
	}

	// If the AWSSession isn't initialized, we cannot grab it
	if AWSSession == nil {
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
	_, err = AWSSession.BucketToFile(bucket, bucketPath, tfile)
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
