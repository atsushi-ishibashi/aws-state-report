package cmd

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/tealeg/xlsx"
)

func hyperlink(sheet string, row, col int, name string) string {
	a, b := col/26, col%26
	colBytes := make([]byte, 0)
	if a != 0 {
		colBytes = append(colBytes, byte(64+a))
	}
	colBytes = append(colBytes, byte(65+b))
	return fmt.Sprintf(`HYPERLINK("#%s!%s%d","%s")`, sheet, string(colBytes), row+1, name)
}

func extractTagName(tags []*ec2.Tag) string {
	var name string
	for _, tg := range tags {
		if *tg.Key == "Name" {
			name = *tg.Value
		}
	}
	return name
}

func borderWithAlign(lrtb string, isAlign bool) *xlsx.Style {
	b := xlsx.Border{}
	btype := "thin"
	bcolor := "FF000000"
	if strings.Index(lrtb, "l") != -1 {
		b.Left, b.LeftColor = btype, bcolor
	}
	if strings.Index(lrtb, "r") != -1 {
		b.Right, b.RightColor = btype, bcolor
	}
	if strings.Index(lrtb, "t") != -1 {
		b.Top, b.TopColor = btype, bcolor
	}
	if strings.Index(lrtb, "b") != -1 {
		b.Bottom, b.BottomColor = btype, bcolor
	}
	st := &xlsx.Style{
		Border:      b,
		ApplyBorder: true,
	}
	if isAlign {
		st.Alignment = xlsx.Alignment{Horizontal: "center"}
	}
	return st
}
