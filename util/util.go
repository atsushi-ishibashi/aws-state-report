package util

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/urfave/cli"
)

const (
	accessKeyID     = "AWS_ACCESS_KEY_ID"
	secretAccessKey = "AWS_SECRET_ACCESS_KEY"
	defaultRegion   = "AWS_DEFAULT_REGION"
)

func ConfigAWS(c *cli.Context) error {
	region := c.GlobalString("awsregion")
	os.Setenv(defaultRegion, region)
	name := c.GlobalString("awsconf")
	if name == "" {
		return nil
	}
	cred := credentials.NewSharedCredentials("", name)
	credValue, err := cred.Get()
	if err != nil {
		return err
	}
	PrintlnGreen(fmt.Sprintf("AWS Profile Name: %s, Region: %s", name, region))
	os.Setenv(accessKeyID, credValue.AccessKeyID)
	os.Setenv(secretAccessKey, credValue.SecretAccessKey)
	return nil
}

//PrintlnGreen Println in Green
func PrintlnGreen(s string) {
	fmt.Printf("\x1b[32m%s\x1b[0m\n", s)
}

//PrintlnRed Println in Red
func PrintlnRed(s string) {
	fmt.Printf("\x1b[31m%s\x1b[0m\n", s)
}

//PrintlnYellow Println in Yellow
func PrintlnYellow(s string) {
	fmt.Printf("\x1b[33m%s\x1b[0m\n", s)
}

//ErrorlnRed Error in Red
func ErrorRed(s string) error {
	return fmt.Errorf("\x1b[31m%s\x1b[0m", s)
}

//SprintGreen Sprintf in Green
func SprintGreen(s string) string {
	return fmt.Sprintf("\x1b[32m%s\x1b[0m", s)
}

//SprintRed Sprintf in Red
func SprintRed(s string) string {
	return fmt.Sprintf("\x1b[31m%s\x1b[0m", s)
}

//SprintYellow Sprintf in Yellow
func SprintYellow(s string) string {
	return fmt.Sprintf("\x1b[33m%s\x1b[0m", s)
}
