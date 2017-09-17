package cmd

import (
	"fmt"
	"math"
	"strings"

	"github.com/atsushi-ishibashi/aws-state-report/svc"
	"github.com/atsushi-ishibashi/aws-state-report/util"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/tealeg/xlsx"
	"github.com/urfave/cli"
)

func NewSGCommand() cli.Command {
	return cli.Command{
		Name:  "sg",
		Usage: "export security groups, network interfaces, instaces and relation among them.",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "src",
				Usage: "file name to export",
				Value: "sg",
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
			sg := &SG{
				manager: mng,
				Errs:    make([]error, 0),
			}
			if err := sg.recursiveConstruct(); err != nil {
				return util.ErrorRed(err.Error())
			}
			sg.convertXlsx(c.String("src"))
			return nil
		},
	}
}

type SG struct {
	SecurityGroups []*SecurityGroup
	manager        *svc.Manager
	Errs           []error
}

func (sg *SG) recursiveConstruct() error {
	sg.constructSecurityGroups().
		constructNetworkInterfaces()
	return sg.flattenErrs()
}

func (sg *SG) constructSecurityGroups() *SG {
	result, err := sg.manager.FetchSecurityGroups()
	if err != nil {
		return sg.stackError(err)
	}
	sg.SecurityGroups = parseDescribeSecurityGroupsOutput(result)
	return sg
}

func (sg *SG) constructNetworkInterfaces() *SG {
	gids := make([]*string, 0)
	for _, v := range sg.SecurityGroups {
		gids = append(gids, aws.String(v.ID))
	}
	result, err := sg.manager.FetchNetworkInterfaces(gids)
	if err != nil {
		return sg.stackError(err)
	}
	nis := parseDescribeNetworkInterfacesOutput(result)
	for _, v := range sg.SecurityGroups {
		for _, ni := range nis {
			if ni.InstanceID != "" {
				diResult, err := sg.manager.FetchEc2Instance(aws.String(ni.InstanceID))
				if err != nil {
					return sg.stackError(err)
				}
				ni.Ec2Instance = parseDescribeInstancesOutput(diResult)
			}
			for _, gid := range ni.GroupIds {
				if gid == v.ID {
					v.NetworkInterfaces = append(v.NetworkInterfaces, ni)
					break
				}
			}
		}
	}
	return sg
}

func (sg *SG) stackError(err error) *SG {
	sg.Errs = append(sg.Errs, err)
	return sg
}

func (sg *SG) flattenErrs() error {
	if len(sg.Errs) == 0 {
		return nil
	}
	var errStr string
	for _, e := range sg.Errs {
		errStr = errStr + e.Error() + "\n"
	}
	return fmt.Errorf(errStr)
}

func parseDescribeSecurityGroupsOutput(output *ec2.DescribeSecurityGroupsOutput) []*SecurityGroup {
	sgs := make([]*SecurityGroup, 0)
	for _, v := range output.SecurityGroups {
		sg := &SecurityGroup{
			ID:                *v.GroupId,
			GroupName:         *v.GroupName,
			TagName:           extractTagName(v.Tags),
			Description:       *v.Description,
			NetworkInterfaces: make([]*NetworkInterface, 0),
		}
		ingress := make([]*IpPermission, 0)
		for _, i := range v.IpPermissions {
			ip := &IpPermission{
				Protocol: *i.IpProtocol,
			}
			if i.FromPort != nil {
				ip.FromPort = *i.FromPort
			}
			if i.ToPort != nil {
				ip.ToPort = *i.ToPort
			}
			if i.IpRanges != nil {
				ranges := make([]string, 0)
				for _, r := range i.IpRanges {
					ranges = append(ranges, *r.CidrIp)
				}
				ip.Ranges = ranges
			}
			if i.UserIdGroupPairs != nil {
				gids := make([]string, 0)
				for _, r := range i.UserIdGroupPairs {
					gids = append(gids, *r.GroupId)
				}
				ip.GroupIds = gids
			}
			ingress = append(ingress, ip)
		}
		sg.Ingress = ingress
		egress := make([]*IpPermission, 0)
		for _, i := range v.IpPermissionsEgress {
			ip := &IpPermission{
				Protocol: *i.IpProtocol,
			}
			if i.FromPort != nil {
				ip.FromPort = *i.FromPort
			}
			if i.ToPort != nil {
				ip.ToPort = *i.ToPort
			}
			if i.IpRanges != nil {
				ranges := make([]string, 0)
				for _, r := range i.IpRanges {
					ranges = append(ranges, *r.CidrIp)
				}
				ip.Ranges = ranges
			}
			if i.UserIdGroupPairs != nil {
				gids := make([]string, 0)
				for _, r := range i.UserIdGroupPairs {
					gids = append(gids, *r.GroupId)
				}
				ip.GroupIds = gids
			}
			egress = append(egress, ip)
		}
		sg.Egress = egress
		sgs = append(sgs, sg)
	}
	return sgs
}

func parseDescribeNetworkInterfacesOutput(output *ec2.DescribeNetworkInterfacesOutput) []*NetworkInterface {
	nis := make([]*NetworkInterface, 0)
	for _, v := range output.NetworkInterfaces {
		ni := &NetworkInterface{
			ID:          *v.NetworkInterfaceId,
			Description: *v.Description,
		}
		if v.Attachment.InstanceId != nil {
			ni.InstanceID = *v.Attachment.InstanceId
		}
		gids := make([]string, 0)
		for _, g := range v.Groups {
			gids = append(gids, *g.GroupId)
		}
		ni.GroupIds = gids
		nis = append(nis, ni)
	}
	return nis
}

func parseDescribeInstancesOutput(output *ec2.DescribeInstancesOutput) *Instance {
	res := output.Reservations[0].Instances[0]
	ins := &Instance{
		ID:               *res.InstanceId,
		TagName:          extractTagName(res.Tags),
		AvailabilityZone: *res.Placement.AvailabilityZone,
		PrivateIP:        *res.PrivateIpAddress,
		InstanceType:     *res.InstanceType,
	}
	if res.PublicIpAddress != nil {
		ins.PublicIP = *res.PublicIpAddress
	}
	if res.KeyName != nil {
		ins.KeyName = *res.KeyName
	}
	return ins
}

func (sg *SG) convertXlsx(filename string) {
	file := xlsx.NewFile()
	nis := make([]*NetworkInterface, 0)
	for _, v := range sg.SecurityGroups {
		nis = appendNIsWithoutDuplicate(nis, v.NetworkInterfaces)
	}
	ec2s := make([]*Instance, 0)
	for _, v := range nis {
		if v.Ec2Instance != nil {
			ec2s = append(ec2s, v.Ec2Instance)
		}
	}
	instanceLocation := make(map[string][2]int)
	sg.convertInstanceToXlsx(file, ec2s, &instanceLocation)
	networkInterfaceLocation := make(map[string][2]int)
	sg.convertNetworkInterfaceToXlsx(file, nis, instanceLocation, &networkInterfaceLocation)
	sg.convertSecurityGroupToXlsx(file, networkInterfaceLocation)
	if err := file.Save(fmt.Sprintf("./%s.xlsx", filename)); err != nil {
		sg.stackError(err)
	}
}

func (sg *SG) convertInstanceToXlsx(file *xlsx.File, ec2s []*Instance, locMap *map[string][2]int) {
	sheet, err := file.AddSheet("instance")
	if err != nil {
		util.PrintlnRed(err.Error())
	}
	m := make(map[string][2]int)
	currentRow := 0
	for _, v := range ec2s {
		m[v.ID] = [2]int{currentRow, 0}
		sheet.Cell(currentRow, 0).Merge(1, 0)
		sheet.Cell(currentRow, 0).Value = fmt.Sprintf("%s, tag: %s", v.ID, v.TagName)
		sheet.Cell(currentRow, 0).SetStyle(borderWithAlign("lrtb", true))
		currentRow++
		sheet.Cell(currentRow, 0).Value = "AvailabilityZone"
		sheet.Cell(currentRow, 0).SetStyle(borderWithAlign("lr", false))
		sheet.Cell(currentRow, 1).Value = v.AvailabilityZone
		sheet.Cell(currentRow, 1).SetStyle(borderWithAlign("lr", false))
		currentRow++
		sheet.Cell(currentRow, 0).Value = "Private IP"
		sheet.Cell(currentRow, 0).SetStyle(borderWithAlign("lr", false))
		sheet.Cell(currentRow, 1).Value = v.PrivateIP
		sheet.Cell(currentRow, 1).SetStyle(borderWithAlign("lr", false))
		currentRow++
		sheet.Cell(currentRow, 0).Value = "Public IP"
		sheet.Cell(currentRow, 0).SetStyle(borderWithAlign("lr", false))
		sheet.Cell(currentRow, 1).Value = v.PublicIP
		sheet.Cell(currentRow, 1).SetStyle(borderWithAlign("lr", false))
		currentRow++
		sheet.Cell(currentRow, 0).Value = "Instance Type"
		sheet.Cell(currentRow, 0).SetStyle(borderWithAlign("lr", false))
		sheet.Cell(currentRow, 1).Value = v.InstanceType
		sheet.Cell(currentRow, 1).SetStyle(borderWithAlign("lr", false))
		currentRow++
		sheet.Cell(currentRow, 0).Value = "Key Name"
		sheet.Cell(currentRow, 0).SetStyle(borderWithAlign("lrb", false))
		sheet.Cell(currentRow, 1).Value = v.KeyName
		sheet.Cell(currentRow, 1).SetStyle(borderWithAlign("lrb", false))
		currentRow += 2
	}
	*locMap = m
}

func (sg *SG) convertNetworkInterfaceToXlsx(file *xlsx.File, nis []*NetworkInterface, refIns map[string][2]int, locMap *map[string][2]int) {
	sheet, err := file.AddSheet("networkinterface")
	if err != nil {
		util.PrintlnRed(err.Error())
	}
	m := make(map[string][2]int)
	currentRow := 0
	for _, v := range nis {
		m[v.ID] = [2]int{currentRow, 0}
		sheet.Cell(currentRow, 0).Merge(1, 0)
		sheet.Cell(currentRow, 0).Value = v.ID
		sheet.Cell(currentRow, 0).SetStyle(borderWithAlign("lrtb", true))
		currentRow++
		sheet.Cell(currentRow, 0).Value = "Description"
		sheet.Cell(currentRow, 0).SetStyle(borderWithAlign("lr", false))
		sheet.Cell(currentRow, 1).Value = v.Description
		sheet.Cell(currentRow, 1).SetStyle(borderWithAlign("lr", false))
		currentRow++
		if v.InstanceID != "" {
			sheet.Cell(currentRow, 0).Value = "Instance"
			sheet.Cell(currentRow, 0).SetStyle(borderWithAlign("lr", false))
			if loc, ok := refIns[v.InstanceID]; ok {
				sheet.Cell(currentRow, 1).SetFormula(hyperlink("instance", loc[0], loc[1], v.InstanceID))
				sheet.Cell(currentRow, 1).SetStyle(borderWithAlign("lr", false))
			}
			currentRow++
		}
		sheet.Cell(currentRow, 0).SetStyle(borderWithAlign("t", false))
		sheet.Cell(currentRow, 1).SetStyle(borderWithAlign("t", false))
		currentRow++
	}
	*locMap = m
}

func (sg *SG) convertSecurityGroupToXlsx(file *xlsx.File, refNi map[string][2]int) {
	sheet, err := file.AddSheet("security-group")
	if err != nil {
		util.PrintlnRed(err.Error())
	}
	currentRow := 0
	for _, v := range sg.SecurityGroups {
		sheet.Cell(currentRow, 0).Merge(5, 0)
		sheet.Cell(currentRow, 0).Value = fmt.Sprintf("%s %s, tag: %s", v.ID, v.GroupName, v.TagName)
		sheet.Cell(currentRow, 0).SetStyle(borderWithAlign("lrtb", true))
		currentRow++
		sheet.Cell(currentRow, 0).Value = "Description"
		sheet.Cell(currentRow, 0).SetStyle(borderWithAlign("lrtb", false))
		sheet.Cell(currentRow, 1).Merge(4, 0)
		sheet.Cell(currentRow, 1).Value = v.Description
		sheet.Cell(currentRow, 1).SetStyle(borderWithAlign("lrtb", false))
		currentRow++
		sheet.Cell(currentRow, 0).Merge(2, 0)
		sheet.Cell(currentRow, 0).Value = "Ingress Rules"
		sheet.Cell(currentRow, 0).SetStyle(borderWithAlign("lrtb", true))
		sheet.Cell(currentRow, 3).Merge(2, 0)
		sheet.Cell(currentRow, 3).Value = "Egress Rules"
		sheet.Cell(currentRow, 3).SetStyle(borderWithAlign("lrtb", true))
		currentRow++
		sheet.Cell(currentRow, 0).Value = "Protocol"
		sheet.Cell(currentRow, 0).SetStyle(borderWithAlign("lrtb", true))
		sheet.Cell(currentRow, 1).Value = "Port"
		sheet.Cell(currentRow, 1).SetStyle(borderWithAlign("lrtb", true))
		sheet.Cell(currentRow, 2).Value = "Target"
		sheet.Cell(currentRow, 2).SetStyle(borderWithAlign("lrtb", true))
		sheet.Cell(currentRow, 3).Value = "Protocol"
		sheet.Cell(currentRow, 3).SetStyle(borderWithAlign("lrtb", true))
		sheet.Cell(currentRow, 4).Value = "Port"
		sheet.Cell(currentRow, 4).SetStyle(borderWithAlign("lrtb", true))
		sheet.Cell(currentRow, 5).Value = "Target"
		sheet.Cell(currentRow, 5).SetStyle(borderWithAlign("lrtb", true))
		currentRow++
		iRow := 0
		for _, i := range v.Ingress {
			sheet.Cell(currentRow+iRow, 0).Value = i.Protocol
			sheet.Cell(currentRow+iRow, 0).SetStyle(borderWithAlign("lr", false))
			sheet.Cell(currentRow+iRow, 1).Value = fmt.Sprintf("%d - %d", i.FromPort, i.ToPort)
			sheet.Cell(currentRow+iRow, 1).SetStyle(borderWithAlign("lr", false))
			if len(i.GroupIds) > 0 {
				sheet.Cell(currentRow+iRow, 2).Value = strings.Join(i.GroupIds, ", ")
			} else {
				sheet.Cell(currentRow+iRow, 2).Value = strings.Join(i.Ranges, ", ")
			}
			sheet.Cell(currentRow+iRow, 2).SetStyle(borderWithAlign("lr", false))
			iRow++
		}
		eRow := 0
		for _, e := range v.Egress {
			sheet.Cell(currentRow+eRow, 3).Value = e.Protocol
			sheet.Cell(currentRow+eRow, 3).SetStyle(borderWithAlign("lr", false))
			sheet.Cell(currentRow+eRow, 4).Value = fmt.Sprintf("%d - %d", e.FromPort, e.ToPort)
			sheet.Cell(currentRow+eRow, 4).SetStyle(borderWithAlign("lr", false))
			if len(e.GroupIds) > 0 {
				sheet.Cell(currentRow+eRow, 5).Value = strings.Join(e.GroupIds, ", ")
			} else {
				sheet.Cell(currentRow+eRow, 5).Value = strings.Join(e.Ranges, ", ")
			}
			sheet.Cell(currentRow+eRow, 5).SetStyle(borderWithAlign("lr", false))
			eRow++
		}
		maxNo := int(math.Max(float64(iRow), float64(eRow)))
		currentRow += maxNo
		sheet.Cell(currentRow, 0).Merge(5, 0)
		sheet.Cell(currentRow, 0).Value = "Network Interface"
		sheet.Cell(currentRow, 0).SetStyle(borderWithAlign("lrtb", true))
		currentRow++
		for i, ni := range v.NetworkInterfaces {
			row, col := i/7, i%7
			if loc, ok := refNi[ni.ID]; ok {
				sheet.Cell(currentRow+row, col).SetFormula(hyperlink("networkinterface", loc[0], loc[1], ni.ID))
			}
			if col == 0 {
				sheet.Cell(currentRow+row, col).SetStyle(borderWithAlign("l", false))
			}
			if col == 6 {
				sheet.Cell(currentRow+row, col).SetStyle(borderWithAlign("r", false))
			}
		}
		addRow := len(v.NetworkInterfaces)/7 + 1
		currentRow += addRow
		sheet.Cell(currentRow, 0).SetStyle(borderWithAlign("t", false))
		sheet.Cell(currentRow, 1).SetStyle(borderWithAlign("t", false))
		sheet.Cell(currentRow, 2).SetStyle(borderWithAlign("t", false))
		sheet.Cell(currentRow, 3).SetStyle(borderWithAlign("t", false))
		sheet.Cell(currentRow, 4).SetStyle(borderWithAlign("t", false))
		sheet.Cell(currentRow, 5).SetStyle(borderWithAlign("t", false))
		currentRow++
	}
}
