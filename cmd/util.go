package cmd

import "github.com/aws/aws-sdk-go/service/ec2"

func extractTagName(tags []*ec2.Tag) string {
	var name string
	for _, tg := range tags {
		if *tg.Key == "Name" {
			name = *tg.Value
		}
	}
	return name
}
