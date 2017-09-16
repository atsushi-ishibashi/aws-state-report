package cmd

type Policy struct {
	Name   string
	Detail string
}

type User struct {
	Name        string
	PolicyNames []string
	GroupNames  []string
}

type Group struct {
	Name        string
	PolicyNames []string
}

type Role struct {
	Name         string
	PolicyNames  []string
	AssumeEntity string
}
