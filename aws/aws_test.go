package aws

import (
	. "github.com/smartystreets/goconvey/convey"

	"testing"
)

func TestS3urlToParts(t *testing.T) {

	Convey("When an S3 URL containing only a bucket is passed, the result is as expected", t, FailureContinues, func() {
		bucket, path, file := S3urlToParts("s3://mybucket/")

		So(bucket, ShouldEqual, "mybucket")
		So(path, ShouldEqual, "")
		So(file, ShouldEqual, "")
	})

	Convey("When an S3 URL containing only a bucket is passed, and there is no trailing slash, the result is as expected", t, FailureContinues, func() {
		bucket, path, file := S3urlToParts("s3://mybucket")

		So(bucket, ShouldEqual, "mybucket")
		So(path, ShouldEqual, "")
		So(file, ShouldEqual, "")
	})

	Convey("When an S3 URL containing only bucket and file is passed, the result is as expected", t, FailureContinues, func() {
		bucket, path, file := S3urlToParts("s3://mybucket/myfile.zip")

		So(bucket, ShouldEqual, "mybucket")
		So(path, ShouldEqual, "myfile.zip")
		So(file, ShouldEqual, "myfile.zip")
	})

	Convey("When an S3 URL containing a bucket, a path, and file is passed, the result is as expected", t, FailureContinues, func() {
		bucket, path, file := S3urlToParts("s3://mybucket/my/folders/myfile.zip")

		So(bucket, ShouldEqual, "mybucket")
		So(path, ShouldEqual, "my/folders/myfile.zip")
		So(file, ShouldEqual, "myfile.zip")
	})

	Convey("When an S3 URL containing a bucket, a path, and NO file is passed, the result is as expected", t, FailureContinues, func() {
		bucket, path, file := S3urlToParts("s3://mybucket/my/folders/")

		So(bucket, ShouldEqual, "mybucket")
		So(path, ShouldEqual, "my/folders/")
		So(file, ShouldEqual, "folders")
	})
}
