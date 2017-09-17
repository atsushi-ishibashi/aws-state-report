# aws-state-report
aws-state-report export infrastructure in aws to excel file.

## Usage
### network
```
$ aws-state-report network --help
NAME:
   aws-state-report network - export vpcs, route tables and subnets information

USAGE:
   aws-state-report network [arguments...]

Examples:
  $ aws-state-report --awsconf default network
```
### iam
```
$ aws-state-report iam --help
NAME:
  aws-state-report iam - export iam roles, users, policies, groups and relation among them.

USAGE:
  aws-state-report iam [command options] [arguments...]

OPTIONS:
  --src value  file name to export (default: "iam")

Examples:
  $ aws-state-report --awsconf default iam
```
### sg
```
$ aws-state-report sg --help
NAME:
  aws-state-report sg - export security groups, network interfaces, instaces and relation among them.

USAGE:
  aws-state-report sg [command options] [arguments...]

OPTIONS:
  --src value  file name to export (default: "sg")

Examples:
  $ aws-state-report --awsconf default sg
```
