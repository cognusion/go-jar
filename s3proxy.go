package jar

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/cognusion/go-timings"

	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

// Constants for configuration key strings
const (
	ConfigS3StreamProxyName              = ConfigKey("s3proxy.name")
	ConfigS3StreamProxyBucket            = ConfigKey("s3proxy.bucket")
	ConfigS3StreamProxyPrefix            = ConfigKey("s3proxy.prefix")
	ConfigS3StreamProxyRedirectURL       = ConfigKey("s3proxy.redirecturl")
	ConfigS3StreamProxyFormNameField     = ConfigKey("s3proxy.namefield")
	ConfigS3StreamProxyFormEmailField    = ConfigKey("s3proxy.emailfield")
	ConfigS3StreamProxyFormToField       = ConfigKey("s3proxy.tofield")
	ConfigS3StreamProxyFormFileField     = ConfigKey("s3proxy.filefield")
	ConfigS3StreamProxyBadFileExtensions = ConfigKey("s3proxy.badfileexts")
	ConfigS3StreamProxyWrapSuccess       = ConfigKey("s3proxy.wrapsuccess")
)

const (
	// ErrS3ProxyConfigNoEC2 is returned when the s3proxy is used, but ec2 is not.
	ErrS3ProxyConfigNoEC2 = Error("s3proxy used, but ec2 is false")
)

var (
	bofcACL = "bucket-owner-full-control"
	charMap map[string]string
)

func init() {
	// Set up the static finishers
	Finishers["s3proxy"] = S3StreamProxyFinisher

	charMap = map[string]string{
		"&":  "",
		"@":  "",
		":":  "",
		",":  "",
		"/":  "",
		"$":  "",
		"=":  "",
		"+":  "",
		"?":  "",
		";":  "",
		"^":  "",
		"`":  "",
		"'":  "",
		"\"": "",
		">":  "",
		"<":  "",
		"{":  "",
		"}":  "",
		"[":  "",
		"]":  "",
		"#":  "",
		"%":  "",
		"~":  "",
		"|":  "",
		"!":  "",
	}
}

// S3StreamProxyFinisher is a finisher that streams a POSTd file to an S3 bucket
func S3StreamProxyFinisher(w http.ResponseWriter, r *http.Request) {
	var (
		pathOptions PathOptions
		badExts     []string
	)

	if popts := r.Context().Value(PathOptionsKey); popts != nil {
		pathOptions = popts.(PathOptions)
	}

	badExts = pathOptions.GetStringSlice(ConfigS3StreamProxyBadFileExtensions)

	fname := fmt.Sprintf("S3StreamProxy %s", pathOptions.GetString(ConfigS3StreamProxyName))
	defer timings.Track(fname, timings.Now(), TimingOut)

	mr, err := r.MultipartReader()
	if err != nil {
		// Form isn't valid?
		ErrorOut.Println(ErrRequestError{r, fmt.Sprintf("Error in S3StreamProxy getting MultipartReader: '%v'", err)})
		http.Error(w, ErrRequestError{r, "There was an error submitting data"}.Error(), http.StatusInternalServerError)
		return
	}

	var (
		name  string
		email string
		to    string
		from  string

		basefn string
	)

	svc := s3manager.NewUploader(Ec2Session.AWS)

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			// We're done
			break
		}
		if err != nil {
			ErrorOut.Println(ErrRequestError{r, fmt.Sprintf("Error in S3StreamProxy reading multipart body: '%v'", err)})
			http.Error(w, ErrRequestError{r, "There was an error reading data"}.Error(), http.StatusInternalServerError)
			return
		}

		switch p.FormName() {
		case pathOptions.GetString(ConfigS3StreamProxyFormNameField):
			name = ReaderToString(p)
			from = fmt.Sprintf("%s <%s>", name, email)
		case pathOptions.GetString(ConfigS3StreamProxyFormEmailField):
			email = ReaderToString(p)
			from = fmt.Sprintf("%s <%s>", name, email)
		case pathOptions.GetString(ConfigS3StreamProxyFormToField):
			to = ReaderToString(p)
		case pathOptions.GetString(ConfigS3StreamProxyFormFileField):
			fn := p.FileName()
			basefn = sanitizeFilename(filepath.Base(fn))

			// See if it's worth grabbing
			if isBadFileMaybe(fn, badExts) {
				p.Close()
				ErrorOut.Println(ErrRequestError{r, fmt.Sprintf("S3StreamProxy is rejecting file '%s' from %s to %s", fn, from, to)})
				http.Error(w, ErrRequestError{r, "The submitted file has been rejected"}.Error(), http.StatusBadRequest)
				return
			}

			if pathOptions.GetString(ConfigS3StreamProxyPrefix) != "" {
				// We have a prefix
				basefn = fmt.Sprintf("%s%s", pathOptions.GetString(ConfigS3StreamProxyPrefix), basefn)
			}

			bucket := pathOptions.GetString(ConfigS3StreamProxyBucket)
			DebugOut.Println(ErrRequestError{r, fmt.Sprintf("S3StreamProxy: Upload '%s' to Bucket '%s' Key '%s'", fn, bucket, basefn)})

			// Upload the file to S3.
			_, err := svc.Upload(&s3manager.UploadInput{
				ACL:    &bofcACL,
				Bucket: aws.String(bucket),
				Key:    aws.String(basefn),
				Body:   p,
			})

			if err != nil {
				p.Close()
				ErrorOut.Println(ErrRequestError{r, fmt.Sprintf("Error in S3StreamProxy uploading file '%s' from %s: %v", basefn, from, err)})
				http.Error(w, ErrRequestError{r, "There was an error uploading data"}.Error(), http.StatusInternalServerError)
				return
			}
			// Win
		}
		p.Close()
	}

	// Success!!

	if pathOptions.GetString(ConfigS3StreamProxyRedirectURL) != "" {
		// After we're done, redirect them elsewhere
		http.Redirect(w, r, pathOptions.GetString(ConfigS3StreamProxyRedirectURL), http.StatusMovedPermanently)
		return
	}

	w.WriteHeader(http.StatusOK)
	if pathOptions.GetBool(ConfigS3StreamProxyWrapSuccess) && ErrorTemplate != nil {
		// Pretty success page requested
		te := TemplateError{
			ErrorCode:    "GENERIC",
			ErrorMessage: "Upload successful",
		}

		err := ErrorTemplate.Execute(w, te)
		if err != nil {
			ErrorOut.Println(ErrRequestError{r, fmt.Sprintf("error executing template: %s", err)})
		}
	} else {
		// No ErrorTemplate
		w.Write(ErrRequestError{r, "Upload successful"}.Bytes())
	}
}

// isBadFileMaybe determintes if a file is "bad" based on its "extension". Awful, I know.
func isBadFileMaybe(filename string, badExts []string) (isBad bool) {
	isBad = false // Default good

	if len(badExts) == 0 {
		return
	}

	ext := strings.ToLower(filepath.Ext(filename))
	for _, e := range badExts {
		if strings.ToLower(strings.TrimSpace(e)) == ext {
			isBad = true
			break
		}
	}

	DebugOut.Printf("isBadFileMaybe '%s' '%v' = %t\n", filename, badExts, isBad)
	return
}

// sanitizeFilename replaces S3 reserved characters with an equivalent
func sanitizeFilename(filename string) string {
	for k, v := range charMap {
		filename = strings.ReplaceAll(filename, k, v)
	}
	return filename
}
