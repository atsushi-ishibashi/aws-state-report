package cmd

type Vpc struct {
	ID                   string
	TagName              string
	CidrBlock            string
	AssociatedCidrBlocks []string
	RouteTables          []*RouteTable
	Subnets              []*Subnet
}

type RouteTable struct {
	ID                 string
	TagName            string
	Routes             []*Route
	AssociationSubnets []string //subnet-id
}

type Route struct {
	DestinationCidrBlock string
	Router               string
}

type Subnet struct {
	ID                   string
	TagName              string
	CidrBlock            string
	AssociatedRouteTable *RouteTable
}
