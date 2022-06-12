# AWS

Enabling intelligent operations when running on Amazon Web Services EC2 instances, either stand-alone or via Elastic Beanstalk, is a core goal for JAR.

## Features

- Unlock AWS features using ``ec2: true``
  - Detects our own Instance information (AZ, machine type, etc.)
- Secured S3 file downloads (specifically for ``updatepath`` and ``hotupdate``)
- Detecting the AZ-locality of Pool members, and preferring local members if ``EC2Affinity: true``
- Load keys via config or environment, or use the instance IAM profile if nothing is provided
- Use S3Proxy to provide file uploads
- Use S3Pools to serve content from S3
