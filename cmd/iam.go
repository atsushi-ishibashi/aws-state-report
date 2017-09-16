package cmd

import (
	"bytes"
	"fmt"
	"math"
	"net/url"

	"github.com/atsushi-ishibashi/aws-state-report/svc"
	"github.com/atsushi-ishibashi/aws-state-report/util"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/tealeg/xlsx"
	"github.com/urfave/cli"
)

func NewIAMCommand() cli.Command {
	return cli.Command{
		Name:  "iam",
		Usage: "export iam roles, users, policies, groups and relation among them.",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "src",
				Usage: "file name to export",
				Value: "iam",
			},
		},
		Action: func(c *cli.Context) error {
			if err := util.ConfigAWS(c); err != nil {
				return util.ErrorRed(err.Error())
			}
			mng, err := svc.NewManager()
			if err != nil {
				return util.ErrorRed(err.Error())
			}
			iam := &IAM{
				manager: mng,
				Errs:    make([]error, 0),
			}
			if err := iam.recursiveConstruct(); err != nil {
				return util.ErrorRed(err.Error())
			}
			iam.convertXlsx(c.String("src"))
			return nil
		},
	}
}

type IAM struct {
	Policies []*Policy
	Users    []*User
	Groups   []*Group
	Roles    []*Role
	manager  *svc.Manager
	Errs     []error
}

func (iam *IAM) recursiveConstruct() error {
	iam.constructPolicies().
		constructGroups().
		constructRoles().
		constructUsers()
	return iam.flattenErrs()
}

func (iam *IAM) constructPolicies() *IAM {
	result, err := iam.manager.FetchPolicies()
	if err != nil {
		return iam.stackError(err)
	}
	iam.Policies = parseListPoliciesOutputToPolicies(result)
	return iam
}

func (iam *IAM) constructGroups() *IAM {
	fgResult, err := iam.manager.FetchGroups()
	if err != nil {
		return iam.stackError(err)
	}
	groups := make([]*Group, 0)
	for _, v := range fgResult.Groups {
		output, err := iam.manager.FetchGroupPolicies(v.GroupName)
		if err != nil {
			iam.stackError(err)
			continue
		}
		pns := parseListGroupPoliciesOutput(output)
		moutput, err := iam.manager.FetchGroupManagedPolicies(v.GroupName)
		if err != nil {
			iam.stackError(err)
			continue
		}
		pns = append(pns, parseListAttachedGroupPoliciesOutput(moutput)...)
		g := &Group{
			Name:        *v.GroupName,
			PolicyNames: pns,
		}
		groups = append(groups, g)
	}
	iam.Groups = groups
	return iam
}

func (iam *IAM) constructUsers() *IAM {
	result, err := iam.manager.FetchUsers()
	if err != nil {
		return iam.stackError(err)
	}
	users := make([]*User, 0)
	for _, v := range result.Users {
		output, err := iam.manager.FetchUserPolicies(v.UserName)
		if err != nil {
			iam.stackError(err)
			continue
		}
		pns := parseListUserPoliciesOutput(output)
		moutput, err := iam.manager.FetchUserManagedPolicies(v.UserName)
		if err != nil {
			iam.stackError(err)
			continue
		}
		pns = append(pns, parseListAttachedUserPoliciesOutput(moutput)...)
		u := &User{Name: *v.UserName}
		u.PolicyNames = pns
		ugOutput, err := iam.manager.FetchUserGroups(v.UserName)
		if err != nil {
			iam.stackError(err)
			continue
		}
		u.GroupNames = parseListGroupsForUserOutput(ugOutput)
		users = append(users, u)
	}
	iam.Users = users
	return iam
}

func (iam *IAM) constructRoles() *IAM {
	result, err := iam.manager.FetchRoles()
	if err != nil {
		return iam.stackError(err)
	}
	roles := make([]*Role, 0)
	for _, v := range result.Roles {
		output, err := iam.manager.FetchRolePolicies(v.RoleName)
		if err != nil {
			iam.stackError(err)
			continue
		}
		pns := parseListRolePoliciesOutput(output)
		moutput, err := iam.manager.FetchRoleManagedPolicies(v.RoleName)
		if err != nil {
			iam.stackError(err)
			continue
		}
		pns = append(pns, parseListAttachedRolePoliciesOutput(moutput)...)
		role := &Role{
			Name:         *v.RoleName,
			AssumeEntity: *v.AssumeRolePolicyDocument,
		}
		role.PolicyNames = pns
		roles = append(roles, role)
	}
	iam.Roles = roles
	return iam
}

func (iam *IAM) stackError(err error) *IAM {
	iam.Errs = append(iam.Errs, err)
	return iam
}

func (iam *IAM) flattenErrs() error {
	if len(iam.Errs) == 0 {
		return nil
	}
	var errStr string
	for _, e := range iam.Errs {
		errStr = errStr + e.Error() + "\n"
	}
	return fmt.Errorf(errStr)
}

func parseListPoliciesOutputToPolicies(output *iam.ListPoliciesOutput) []*Policy {
	pls := make([]*Policy, 0)
	for _, v := range output.Policies {
		p := &Policy{
			Name:   *v.PolicyName,
			Detail: *v.Description,
		}
		pls = append(pls, p)
	}
	return pls
}

func parseListUserPoliciesOutput(output *iam.ListUserPoliciesOutput) []string {
	pns := make([]string, len(output.PolicyNames))
	for _, v := range output.PolicyNames {
		pns = append(pns, *v)
	}
	return pns
}

func parseListAttachedUserPoliciesOutput(output *iam.ListAttachedUserPoliciesOutput) []string {
	pns := make([]string, len(output.AttachedPolicies))
	for _, v := range output.AttachedPolicies {
		pns = append(pns, *v.PolicyName)
	}
	return pns
}

func parseListGroupsForUserOutput(output *iam.ListGroupsForUserOutput) []string {
	gs := make([]string, len(output.Groups))
	for _, v := range output.Groups {
		gs = append(gs, *v.GroupName)
	}
	return gs
}

func parseListGroupPoliciesOutput(output *iam.ListGroupPoliciesOutput) []string {
	pns := make([]string, len(output.PolicyNames))
	for _, v := range output.PolicyNames {
		pns = append(pns, *v)
	}
	return pns
}

func parseListAttachedGroupPoliciesOutput(output *iam.ListAttachedGroupPoliciesOutput) []string {
	pns := make([]string, len(output.AttachedPolicies))
	for _, v := range output.AttachedPolicies {
		pns = append(pns, *v.PolicyName)
	}
	return pns
}

func parseListRolePoliciesOutput(output *iam.ListRolePoliciesOutput) []string {
	pns := make([]string, len(output.PolicyNames))
	for _, v := range output.PolicyNames {
		pns = append(pns, *v)
	}
	return pns
}

func parseListAttachedRolePoliciesOutput(output *iam.ListAttachedRolePoliciesOutput) []string {
	pns := make([]string, len(output.AttachedPolicies))
	for _, v := range output.AttachedPolicies {
		pns = append(pns, *v.PolicyName)
	}
	return pns
}

func (iam *IAM) convertXlsx(filename string) {
	file := xlsx.NewFile()

	//policy
	policySheet, err := file.AddSheet("policy")
	if err != nil {
		util.PrintlnRed(err.Error())
	}
	policyLocation := make(map[string][2]int)
	currentPolicyRow := 0
	for _, v := range iam.Policies {
		policyLocation[v.Name] = [2]int{currentPolicyRow, 0}
		policySheet.Cell(currentPolicyRow, 0).Value = v.Name
		policySheet.Cell(currentPolicyRow, 0).SetStyle(borderWithAlign("lrtb", false))
		currentPolicyRow++
		ss, _ := url.QueryUnescape(v.Detail)
		sb := bytes.Replace([]byte(ss), []byte{32}, []byte{}, -1)
		sb = bytes.Replace(sb, []byte{10}, []byte{}, -1)
		sb = bytes.Replace(sb, []byte{123}, []byte{123, 10}, -1)
		sb = bytes.Replace(sb, []byte{91}, []byte{91, 10}, -1)
		sb = bytes.Replace(sb, []byte{44}, []byte{44, 10}, -1)
		sb = bytes.Replace(sb, []byte{125}, []byte{10, 125}, -1)
		sb = bytes.Replace(sb, []byte{93}, []byte{10, 93}, -1)
		policySheet.Cell(currentPolicyRow, 0).SetValue(sb)
		policySheet.Cell(currentPolicyRow, 0).SetStyle(borderWithAlign("lrtb", false))
		currentPolicyRow++
		currentPolicyRow++
	}

	//group
	groupSheet, err := file.AddSheet("group")
	if err != nil {
		util.PrintlnRed(err.Error())
	}
	groupLocation := make(map[string][2]int)
	currentGroupRow := 0
	for _, v := range iam.Groups {
		groupLocation[v.Name] = [2]int{currentGroupRow, 0}
		groupSheet.Cell(currentGroupRow, 0).Value = v.Name
		groupSheet.Cell(currentGroupRow, 0).SetStyle(borderWithAlign("lrtb", false))
		currentGroupRow++
		for _, up := range v.PolicyNames {
			loc, ok := policyLocation[up]
			if !ok {
				continue
			}
			groupSheet.Cell(currentGroupRow, 0).SetFormula(hyperlink("policy", loc[0], loc[1], up))
			groupSheet.Cell(currentGroupRow, 0).SetStyle(borderWithAlign("lr", false))
			currentGroupRow++
		}
		groupSheet.Cell(currentGroupRow, 0).SetStyle(borderWithAlign("t", false))
		currentGroupRow++
	}

	//user
	userSheet, err := file.AddSheet("user")
	if err != nil {
		util.PrintlnRed(err.Error())
	}
	currentUserRow := 0
	for _, v := range iam.Users {
		userSheet.Cell(currentUserRow, 0).Value = v.Name
		userSheet.Cell(currentUserRow, 0).SetStyle(borderWithAlign("lrtb", true))
		userSheet.Cell(currentUserRow, 0).Merge(1, 0)
		currentUserRow++
		userSheet.Cell(currentUserRow, 0).Value = "Groups"
		userSheet.Cell(currentUserRow, 0).SetStyle(borderWithAlign("lrtb", true))
		userSheet.Cell(currentUserRow, 1).Value = "Policies"
		userSheet.Cell(currentUserRow, 1).SetStyle(borderWithAlign("lrtb", true))
		currentUserRow++
		ugNo := 0
		for _, gn := range v.GroupNames {
			loc, ok := groupLocation[gn]
			if !ok {
				continue
			}
			userSheet.Cell(currentUserRow+ugNo, 0).SetFormula(hyperlink("group", loc[0], loc[1], gn))
			userSheet.Cell(currentUserRow+ugNo, 0).SetStyle(borderWithAlign("lr", false))
			ugNo++
		}
		upnNo := 0
		for _, up := range v.PolicyNames {
			loc, ok := policyLocation[up]
			if !ok {
				continue
			}
			userSheet.Cell(currentUserRow+upnNo, 1).SetFormula(hyperlink("policy", loc[0], loc[1], up))
			userSheet.Cell(currentUserRow+upnNo, 1).SetStyle(borderWithAlign("lr", false))
			upnNo++
		}
		maxNo := int(math.Max(float64(ugNo), float64(upnNo)))
		for i := 0; i < maxNo; i++ {
			userSheet.Cell(currentUserRow+i, 0).SetStyle(borderWithAlign("lr", false))
			userSheet.Cell(currentUserRow+i, 1).SetStyle(borderWithAlign("lr", false))
		}
		currentUserRow += maxNo
		userSheet.Cell(currentUserRow, 0).SetStyle(borderWithAlign("t", false))
		userSheet.Cell(currentUserRow, 1).SetStyle(borderWithAlign("t", false))
		currentUserRow++
	}

	//role
	roleSheet, err := file.AddSheet("role")
	if err != nil {
		util.PrintlnRed(err.Error())
	}
	currentRoleRow := 0
	for _, v := range iam.Roles {
		roleSheet.Cell(currentRoleRow, 0).Value = v.Name
		roleSheet.Cell(currentRoleRow, 0).SetStyle(borderWithAlign("lrtb", true))
		roleSheet.Cell(currentRoleRow, 0).Merge(1, 0)
		currentRoleRow++
		roleSheet.Cell(currentRoleRow, 0).Value = "Assume Entity"
		roleSheet.Cell(currentRoleRow, 0).SetStyle(borderWithAlign("lrtb", true))
		roleSheet.Cell(currentRoleRow, 1).Value = "Policies"
		roleSheet.Cell(currentRoleRow, 1).SetStyle(borderWithAlign("lrtb", true))
		currentRoleRow++
		ss, _ := url.QueryUnescape(v.AssumeEntity)
		sb := bytes.Replace([]byte(ss), []byte{32}, []byte{}, -1)
		sb = bytes.Replace(sb, []byte{10}, []byte{}, -1)
		sb = bytes.Replace(sb, []byte{123}, []byte{123, 10}, -1)
		sb = bytes.Replace(sb, []byte{91}, []byte{91, 10}, -1)
		sb = bytes.Replace(sb, []byte{44}, []byte{44, 10}, -1)
		sb = bytes.Replace(sb, []byte{125}, []byte{10, 125}, -1)
		sb = bytes.Replace(sb, []byte{93}, []byte{10, 93}, -1)
		roleSheet.Cell(currentRoleRow, 0).SetValue(sb)
		roleSheet.Cell(currentRoleRow, 0).SetStyle(borderWithAlign("lrtb", true))
		pnNo := 0
		for _, up := range v.PolicyNames {
			loc, ok := policyLocation[up]
			if !ok {
				continue
			}
			roleSheet.Cell(currentRoleRow+pnNo, 1).SetFormula(hyperlink("policy", loc[0], loc[1], up))
			roleSheet.Cell(currentRoleRow+pnNo, 1).SetStyle(borderWithAlign("lr", false))
			pnNo++
		}
		maxNo := int(math.Max(float64(1), float64(pnNo)))
		for i := 0; i < maxNo; i++ {
			roleSheet.Cell(currentRoleRow+i, 0).SetStyle(borderWithAlign("lr", false))
			roleSheet.Cell(currentRoleRow+i, 1).SetStyle(borderWithAlign("lr", false))
		}
		currentRoleRow += maxNo
		roleSheet.Cell(currentRoleRow, 0).SetStyle(borderWithAlign("t", false))
		roleSheet.Cell(currentRoleRow, 1).SetStyle(borderWithAlign("t", false))
		currentRoleRow++
	}
	if err := file.Save(fmt.Sprintf("./%s.xlsx", filename)); err != nil {
		iam.stackError(err)
	}
}
