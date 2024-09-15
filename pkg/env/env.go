package env

type Environment int

const (
	Development Environment = iota + 1
	Staging
	Production
)

// String implements the Stringer interface
func (e Environment) String() string {
	return [...]string{"dev", "stage", "prod"}[e]
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
