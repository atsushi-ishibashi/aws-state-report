package cmd

import (
	"fmt"
	"math"

	"github.com/atsushi-ishibashi/aws-state-report/svc"
	"github.com/atsushi-ishibashi/aws-state-report/util"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/jung-kurt/gofpdf"
	"github.com/tealeg/xlsx"
	"github.com/urfave/cli"
)

func NewNetworkCommand() cli.Command {
	return cli.Command{
		Name:  "network",
		Usage: "export vpcs, route tables and subnets information",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "src",
				Usage: "file name to export",
				Value: "network",
			},
			cli.BoolFlag{
				Name:  "pdf-mode",
				Usage: "output in pdf file.",
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
			ntw := &Network{
				manager: mng,
				Errs:    make([]error, 0),
			}
			if err := ntw.recursiveConstruct(); err != nil {
				return util.ErrorRed(err.Error())
			}
			if c.Bool("pdf-mode") {
				ntw.convertPdf()
			} else {
				ntw.convertXlsx(c.String("src"))
			}
			return nil
		},
	}
}

type Network struct {
	Vpcs    []*Vpc
	manager *svc.Manager
	Errs    []error
}

func (nt *Network) recursiveConstruct() error {
	nt.constructVpcs().
		constructRouteTables().
		constructSubnets().
		associateRouteTableSubnet()
	return nt.flattenErrs()
}

func (nt *Network) constructVpcs() *Network {
	result, err := nt.manager.FetchVpcs()
	if err != nil {
		return nt.stackError(err)
	}
	nt.Vpcs = parseDescribeVpcsOutputToVpcs(result)
	return nt
}

func (nt *Network) constructRouteTables() *Network {
	for _, vpc := range nt.Vpcs {
		if result, err := nt.manager.FetchRouteTablesWithVpc(vpc.ID); err != nil {
			nt.stackError(err)
		} else {
			vpc.RouteTables = parseDescribeRouteTablesOutputToRouteTables(result)
		}
	}
	return nt
}

func (nt *Network) constructSubnets() *Network {
	for _, vpc := range nt.Vpcs {
		if result, err := nt.manager.FetchSubnetsWithVpc(vpc.ID); err != nil {
			nt.stackError(err)
		} else {
			vpc.Subnets = parseDescribeSubnetsOutputToSubnets(result)
		}
	}
	return nt
}

func (nt *Network) associateRouteTableSubnet() *Network {
	for _, vpc := range nt.Vpcs {
		for _, sn := range vpc.Subnets {
			for _, rt := range vpc.RouteTables {
				for _, rtas := range rt.AssociationSubnets {
					if rtas == sn.ID {
						sn.AssociatedRouteTable = rt
					}
				}
			}
		}
	}
	return nt
}

func (nt *Network) convertXlsx(filename string) {
	file := xlsx.NewFile()
	for _, v := range nt.Vpcs {
		sheet, err := file.AddSheet(v.TagName)
		if err != nil {
			util.PrintlnRed(err.Error())
			continue
		}
		currentRow := 0
		headCell := sheet.Cell(currentRow, 0)
		headCell.Value = fmt.Sprintf("%s  %s", v.TagName, v.CidrBlock)
		headCell.Merge(3, 0)
		headCell.SetStyle(borderWithAlign("lrtb", true))
		currentRow++
		for _, rt := range v.RouteTables {
			rtCell := sheet.Cell(currentRow, 0)
			rtCell.Value = fmt.Sprintf("Route Table: %s", rt.TagName)
			rtCell.Merge(1, 0)
			rtCell.SetStyle(borderWithAlign("lrtb", true))
			snCell := sheet.Cell(currentRow, 2)
			snCell.Value = "Association Subnets"
			snCell.Merge(1, 0)
			snCell.SetStyle(borderWithAlign("lrtb", true))
			currentRow++
			var rtNo int
			for _, rtr := range rt.Routes {
				sheet.Cell(currentRow+rtNo, 0).Value = rtr.DestinationCidrBlock
				sheet.Cell(currentRow+rtNo, 0).SetStyle(borderWithAlign("l", false))
				sheet.Cell(currentRow+rtNo, 1).Value = rtr.Router
				sheet.Cell(currentRow+rtNo, 1).SetStyle(borderWithAlign("r", false))
				rtNo++
			}
			var snNo int
			for _, sn := range v.Subnets {
				if sn.AssociatedRouteTable == rt {
					sheet.Cell(currentRow+snNo, 2).Value = sn.TagName
					sheet.Cell(currentRow+snNo, 2).SetStyle(borderWithAlign("l", false))
					sheet.Cell(currentRow+snNo, 3).Value = sn.CidrBlock
					sheet.Cell(currentRow+snNo, 3).SetStyle(borderWithAlign("r", false))
					snNo++
				}
			}
			maxNo := int(math.Max(float64(rtNo), float64(snNo)))
			for i := 0; i < maxNo; i++ {
				sheet.Cell(currentRow+i, 0).SetStyle(borderWithAlign("l", false))
				sheet.Cell(currentRow+i, 3).SetStyle(borderWithAlign("r", false))
			}
			currentRow += maxNo
		}
		sheet.Cell(currentRow, 0).SetStyle(borderWithAlign("t", false))
		sheet.Cell(currentRow, 1).SetStyle(borderWithAlign("t", false))
		noaSnCell := sheet.Cell(currentRow, 2)
		noaSnCell.Value = "No Association Subnets"
		noaSnCell.Merge(1, 0)
		noaSnCell.SetStyle(borderWithAlign("lrtb", true))
		currentRow++
		for _, sn := range v.Subnets {
			if sn.AssociatedRouteTable == nil {
				sheet.Cell(currentRow, 2).Value = sn.TagName
				sheet.Cell(currentRow, 2).SetStyle(borderWithAlign("l", false))
				sheet.Cell(currentRow, 3).Value = sn.CidrBlock
				sheet.Cell(currentRow, 3).SetStyle(borderWithAlign("r", false))
				currentRow++
			}
		}
		sheet.Cell(currentRow, 2).SetStyle(borderWithAlign("t", false))
		sheet.Cell(currentRow, 3).SetStyle(borderWithAlign("t", false))
	}
	if err := file.Save(fmt.Sprintf("./%s.xlsx", filename)); err != nil {
		nt.stackError(err)
	}
}

func (nt *Network) convertPdf() {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 10)
	for _, v := range nt.Vpcs {
		pdf.CellFormat(0, 10, fmt.Sprintf("%s  %s", v.TagName, v.CidrBlock), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
		for _, rt := range v.RouteTables {
			pdf.CellFormat(95, 10, fmt.Sprintf("%s", rt.TagName), "1", 0, "C", false, 0, "")
			pdf.CellFormat(95, 10, "Association Subnets", "1", 0, "C", false, 0, "")
			pdf.Ln(-1)
			currentX, currentY := pdf.GetXY()
			var rtHeight float64
			for _, rtr := range rt.Routes {
				pdf.MoveTo(currentX, currentY+rtHeight)
				pdf.CellFormat(95, 10, fmt.Sprintf("%s %s", rtr.DestinationCidrBlock, rtr.Router), "RL", 0, "C", false, 0, "")
				rtHeight += 10.0
			}
			var snHeight float64
			for _, sn := range v.Subnets {
				if sn.AssociatedRouteTable == rt {
					pdf.MoveTo(currentX+95, currentY+snHeight)
					pdf.CellFormat(95, 10, fmt.Sprintf("%s %s", sn.TagName, sn.CidrBlock), "RL", 0, "C", false, 0, "")
					snHeight += 10.0
				}
			}
			maxHeight := math.Max(snHeight, rtHeight)
			pdf.MoveTo(currentX, currentY)
			pdf.CellFormat(0, maxHeight, "", "1", 0, "C", false, 0, "")
			pdf.Ln(-1)
		}
		pdf.CellFormat(0, 10, "No Association Subnets", "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
		currentX, currentY := pdf.GetXY()
		var noaSnHeight float64
		for _, sn := range v.Subnets {
			if sn.AssociatedRouteTable == nil {
				pdf.CellFormat(0, 10, fmt.Sprintf("%s %s", sn.TagName, sn.CidrBlock), "LR", 0, "C", false, 0, "")
				pdf.Ln(-1)
				noaSnHeight += 10
			}
		}
		pdf.MoveTo(currentX, currentY)
		pdf.CellFormat(0, noaSnHeight, "", "1", 0, "C", false, 0, "")
		pdf.AddPage()
	}
	if err := pdf.OutputFileAndClose("./network.pdf"); err != nil {
		nt.stackError(err)
	}
}

func (nt *Network) stackError(err error) *Network {
	nt.Errs = append(nt.Errs, err)
	return nt
}

func (nt *Network) flattenErrs() error {
	if len(nt.Errs) == 0 {
		return nil
	}
	var errStr string
	for _, e := range nt.Errs {
		errStr = errStr + e.Error() + "\n"
	}
	return fmt.Errorf(errStr)
}

func parseDescribeVpcsOutputToVpcs(output *ec2.DescribeVpcsOutput) []*Vpc {
	vs := make([]*Vpc, 0)
	for _, v := range output.Vpcs {
		vpc := &Vpc{
			ID:        *v.VpcId,
			TagName:   extractTagName(v.Tags),
			CidrBlock: *v.CidrBlock,
		}
		acbs := make([]string, 0)
		for _, cbs := range v.CidrBlockAssociationSet {
			acbs = append(acbs, *cbs.CidrBlock)
		}
		vpc.AssociatedCidrBlocks = acbs
		vs = append(vs, vpc)
	}
	return vs
}

func parseDescribeRouteTablesOutputToRouteTables(output *ec2.DescribeRouteTablesOutput) []*RouteTable {
	rts := make([]*RouteTable, 0)
	for _, v := range output.RouteTables {
		rt := &RouteTable{
			ID:      *v.RouteTableId,
			TagName: extractTagName(v.Tags),
		}
		rs := make([]*Route, 0)
		for _, r := range v.Routes {
			if r.DestinationCidrBlock == nil {
				continue
			}
			rr := &Route{
				DestinationCidrBlock: *r.DestinationCidrBlock,
			}
			var routerID string
			if r.GatewayId != nil {
				routerID = *r.GatewayId
			}
			if r.NatGatewayId != nil {
				routerID = *r.NatGatewayId
			}
			if r.VpcPeeringConnectionId != nil {
				routerID = *r.VpcPeeringConnectionId
			}
			rr.Router = routerID
			rs = append(rs, rr)
		}
		rt.Routes = rs
		asSubnets := make([]string, 0)
		for _, as := range v.Associations {
			if as.SubnetId != nil {
				asSubnets = append(asSubnets, *as.SubnetId)
			} else {
				asSubnets = append(asSubnets, "implicit")
			}
		}
		rt.AssociationSubnets = asSubnets
		rts = append(rts, rt)
	}
	return rts
}

func parseDescribeSubnetsOutputToSubnets(output *ec2.DescribeSubnetsOutput) []*Subnet {
	subnets := make([]*Subnet, 0)
	for _, v := range output.Subnets {
		sn := &Subnet{
			ID:        *v.SubnetId,
			TagName:   extractTagName(v.Tags),
			CidrBlock: *v.CidrBlock,
		}
		subnets = append(subnets, sn)
	}
	return subnets
}
