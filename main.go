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
	sgCommand := cmd.NewSGCommand()

	app.Commands = []cli.Command{
		networkCommand,
		iamCommand,
		sgCommand,
	}
	app.Run(os.Args)
}
