package svc

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type SGClient struct {
	*ec2.EC2
}

func (c *SGClient) FetchSecurityGroups() (*ec2.DescribeSecurityGroupsOutput, error) {
	input := &ec2.DescribeSecurityGroupsInput{}
	return c.DescribeSecurityGroups(input)
}

func (c *SGClient) FetchNetworkInterfaces(gids []*string) (*ec2.DescribeNetworkInterfacesOutput, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("group-id"),
				Values: gids,
			},
		},
	}
	return c.DescribeNetworkInterfaces(input)
}

func (c *SGClient) FetchEc2Instance(iid *string) (*ec2.DescribeInstancesOutput, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{iid},
	}
	return c.DescribeInstances(input)
}
