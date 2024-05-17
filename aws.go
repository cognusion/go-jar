package jar

import (
	"fmt"

	"github.com/cognusion/go-jar/aws"
)

// Constants for configuration
const (
	ConfigEC2              = ConfigKey("ec2")
	ConfigKeysAwsRegion    = ConfigKey("keys.aws.region")
	ConfigKeysAwsAccessKey = ConfigKey("keys.aws.access")
	ConfigKeysAwsSecretKey = ConfigKey("keys.aws.secret")
)

// Constants for errors
const (
	// ErrNoSession is called when an AWS feature is called, but there is no initialized AWS session
	ErrNoSession = Error("there is no initialized AWS session")
)

var (
	// AWSSession is an aws.Session for use in various places
	AWSSession *aws.Session
)

func init() {
	Bootstrappers["aws"] = awsInit // hook

	// 	v.SetDefault(ConfigEC2, false)                              // Enable AWS EC2-specific features
}

// awsInit is a Bootstrapper to load AWS-specific stuff early in the startup process
func awsInit() error {
	// If we're going to use AWS/EC2 features, we need to turn this on early
	if Conf.GetBool(ConfigEC2) || Conf.GetString(ConfigKeysAwsAccessKey) != "" {
		aws.DebugOut = DebugOut
		aws.TimingOut = TimingOut

		var (
			awsRegion    = Conf.GetString(ConfigKeysAwsRegion)
			awsAccessKey = Conf.GetString(ConfigKeysAwsAccessKey)
			awsSecretKey = Conf.GetString(ConfigKeysAwsSecretKey)
			ec2          = Conf.GetBool(ConfigEC2)
			err          error
		)

		DebugOut.Printf("AWS Setup: Region: %s AccessKey: %s SecretKey: hahaha EC2: %t\n", awsRegion, awsAccessKey, ec2)
		AWSSession, err = aws.NewSession(awsRegion, awsAccessKey, awsSecretKey, ec2)
		if err != nil {
			return fmt.Errorf("error intializing AWS session: '%w'", err)
		}
	}
	return nil
}
