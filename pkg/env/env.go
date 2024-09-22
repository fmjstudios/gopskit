package env

import "strings"

type Environment int

const (
	Development Environment = iota + 1
	Staging
	Production
)

// String implements the Stringer interface
func (e Environment) String() string {
	return [...]string{"dev", "stage", "prod"}[e-1]
}

// Index makes the index of the Environment retrievable
func (e Environment) Index() int {
	return int(e)
}

// IsProduction is a utility function to quickly determine the environment
func (e Environment) IsProduction() bool {
	return e == Production
}

// IsStaging is a utility function to quickly determine the environment
func (e Environment) IsStaging() bool {
	return e == Staging
}

// IsDevelopment is a utility function to quickly determine the environment
func (e Environment) IsDevelopment() bool {
	return e == Development
}

func FromString(s string) Environment {
	switch strings.ToLower(s) {
	case "prod":
		return Production
	case "stage":
		return Staging
	default:
		return Development
	}
}
