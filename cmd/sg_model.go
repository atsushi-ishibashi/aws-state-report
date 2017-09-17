package cmd

type SecurityGroup struct {
	ID                string
	GroupName         string
	TagName           string
	Description       string
	Ingress           []*IpPermission
	Egress            []*IpPermission
	NetworkInterfaces []*NetworkInterface
}

type IpPermission struct {
	Protocol string
	FromPort int64
	ToPort   int64
	Ranges   []string
	GroupIds []string
}

type NetworkInterface struct {
	ID          string
	Description string
	InstanceID  string
	Ec2Instance *Instance
	GroupIds    []string
}

type Instance struct {
	ID               string
	AvailabilityZone string
	PrivateIP        string
	PublicIP         string
	InstanceType     string
	KeyName          string
	TagName          string
}

func appendNIsWithoutDuplicate(slices, elements []*NetworkInterface) []*NetworkInterface {
	a := append(slices, elements...)
	encountered := map[*NetworkInterface]bool{}
	result := []*NetworkInterface{}
	for v := range a {
		if encountered[a[v]] == true {
		} else {
			encountered[a[v]] = true
			result = append(result, a[v])
		}
	}
	return result
}
