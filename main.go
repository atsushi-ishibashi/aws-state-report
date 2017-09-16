package main

import (
	"os"

	"github.com/atsushi-ishibashi/aws-state-report/cmd"
	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "awsconf",
			Usage: "~/.aws/credentialsから環境変数をセット(プロセスの間のみ)",
		},
		cli.StringFlag{
			Name:  "awsregion",
			Usage: "AWS_DEFAULT_REGIONにセット(プロセスの間のみ)",
			Value: "ap-northeast-1",
		},
	}

	networkCommand := cmd.NewNetworkCommand()
	iamCommand := cmd.NewIAMCommand()

	app.Commands = []cli.Command{
		networkCommand,
		iamCommand,
	}
	app.Run(os.Args)

	// pdf := gofpdf.New("P", "mm", "A4", "")
	// pdf.AddPage()
	// pdf.SetFont("Arial", "B", 16)
	// pdf.Cell(40, 50, "Hello World!")
	// pdf.Image("./images/ec2.png", 10, 10, 30, 0, false, "", 0, "")
	// pdf.Text(50, 20, "ec2.png")
	// fileStr := "./basic.pdf"
	// err := pdf.OutputFileAndClose(fileStr)
	// if err != nil {
	// 	fmt.Println(err)
	// }
}
