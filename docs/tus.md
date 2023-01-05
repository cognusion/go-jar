# TUS

The TUS finisher supports the [TUS](https://tus.io/) resumable upload protocol. Each **Path** needs a **tus.targeturi** set to a `file://` for local folder spooling or `s3://` for S3 spooling.

If you do "parallel" uploads > 1 on the client, there will be multiple "part files" left behind, in addition to the final file. It is recommended that your upload area be cleaned periodically of files old files. Yes, we could keep track of those parts, and after the final file is finished, delete the "part files" for you. We aren't.

Also, if you're using `file://` with multiple JAR instances, ensure you're also pinning sessions to the same instance using [consistent hashing](consistenthashing.md) or stickycookie.

## Roadmap

~~S3 backend with multipart upload support~~ Done

Post-upload event hooks trigger worker actions, which could notify, subrequest the file elsewhere, etc.

## Configuration

```yaml
-
    Path: /tus/
    Options:
      tus.targeturi: file:///tmp/tus/
      tus.appendfilename: true
    Finisher: tus
-
    Path: /tus2/
    Options:
      tus.targeturi: s3://my-s3-bucket
      tus.appendfilename: true
    Finisher: tus
```

### tus.targeturi: [file:// or s3:// URI for target]

Please note the `file://` URIs need an extra `/` for fully-qualified paths (e.g. `file:///tmp/tus/`).

Please note that only root-level S3 buckets are supported at this time (no "folders").

### tus.appendfilename: [true/false]

If `true` will append the original filename to the target filename, e.g. `hash-filename.ext`. **NOTE:** for S3 this is a COPY, DELETE and will incur additional charges. Also, for S3 this operation is **limited to files < 5GB in size**, as there is extra work required to "copy" files larger than 5GB within S3, and we're not doing that right now.
