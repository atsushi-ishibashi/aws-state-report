package svc

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

type IAMClient struct {
	*iam.IAM
}

func (c *IAMClient) FetchRoles() (*iam.ListRolesOutput, error) {
	input := &iam.ListRolesInput{}
	return c.ListRoles(input)
}

func (c *IAMClient) FetchRolePolicies(name *string) (*iam.ListRolePoliciesOutput, error) {
	input := &iam.ListRolePoliciesInput{
		RoleName: name,
	}
	return c.ListRolePolicies(input)
}

func (c *IAMClient) FetchRoleManagedPolicies(name *string) (*iam.ListAttachedRolePoliciesOutput, error) {
	input := &iam.ListAttachedRolePoliciesInput{
		RoleName: name,
	}
	return c.ListAttachedRolePolicies(input)
}

func (c *IAMClient) FetchGroups() (*iam.ListGroupsOutput, error) {
	input := &iam.ListGroupsInput{}
	return c.ListGroups(input)
}

func (c *IAMClient) FetchGroupPolicies(name *string) (*iam.ListGroupPoliciesOutput, error) {
	input := &iam.ListGroupPoliciesInput{
		GroupName: name,
	}
	return c.ListGroupPolicies(input)
}

func (c *IAMClient) FetchGroupManagedPolicies(name *string) (*iam.ListAttachedGroupPoliciesOutput, error) {
	input := &iam.ListAttachedGroupPoliciesInput{
		GroupName: name,
	}
	return c.ListAttachedGroupPolicies(input)
}

func (c *IAMClient) FetchUsers() (*iam.ListUsersOutput, error) {
	input := &iam.ListUsersInput{}
	return c.ListUsers(input)
}

func (c *IAMClient) FetchUserPolicies(name *string) (*iam.ListUserPoliciesOutput, error) {
	input := &iam.ListUserPoliciesInput{
		UserName: name,
	}
	return c.ListUserPolicies(input)
}

func (c *IAMClient) FetchUserManagedPolicies(name *string) (*iam.ListAttachedUserPoliciesOutput, error) {
	input := &iam.ListAttachedUserPoliciesInput{
		UserName: name,
	}
	return c.ListAttachedUserPolicies(input)
}

func (c *IAMClient) FetchUserGroups(name *string) (*iam.ListGroupsForUserOutput, error) {
	input := &iam.ListGroupsForUserInput{
		UserName: name,
	}
	return c.ListGroupsForUser(input)
}

func (c *IAMClient) FetchPolicies() (*iam.ListPoliciesOutput, error) {
	input := &iam.ListPoliciesInput{
		OnlyAttached: aws.Bool(true),
	}
	result, err := c.ListPolicies(input)
	if err != nil {
		return nil, err
	}
	for _, v := range result.Policies {
		gpvResult, err := c.fetchPolicyVersion(v.Arn, v.DefaultVersionId)
		if err != nil {
			return nil, err
		}
		v.Description = gpvResult.PolicyVersion.Document
	}
	return result, nil
}

func (c *IAMClient) fetchPolicyVersion(arn, version *string) (*iam.GetPolicyVersionOutput, error) {
	input := &iam.GetPolicyVersionInput{
		PolicyArn: arn,
		VersionId: version,
	}
	return c.GetPolicyVersion(input)
}
