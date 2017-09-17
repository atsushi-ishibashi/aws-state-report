package svc

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
)

type Manager struct {
	*EC2Client
	*IAMClient
	*SGClient
}

func NewManager() (*Manager, error) {
	awsregion := os.Getenv("AWS_DEFAULT_REGION")
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	m := &Manager{}
	m.EC2Client = &EC2Client{EC2: ec2.New(sess, &aws.Config{Region: aws.String(awsregion)})}
	m.IAMClient = &IAMClient{IAM: iam.New(sess, &aws.Config{Region: aws.String(awsregion)})}
	m.SGClient = &SGClient{EC2: ec2.New(sess, &aws.Config{Region: aws.String(awsregion)})}
	return m, nil
}
