package svc

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type EC2Client struct {
	*ec2.EC2
}

func (c *EC2Client) FetchVpcs() (*ec2.DescribeVpcsOutput, error) {
	input := &ec2.DescribeVpcsInput{}
	return c.DescribeVpcs(input)
}

func (c *EC2Client) FetchRouteTablesWithVpc(vpcID string) (*ec2.DescribeRouteTablesOutput, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpcID)},
			},
		},
	}
	return c.DescribeRouteTables(input)
}

func (c *EC2Client) FetchSubnetsWithVpc(vpcID string) (*ec2.DescribeSubnetsOutput, error) {
	input := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpcID)},
			},
		},
	}
	return c.DescribeSubnets(input)
}
