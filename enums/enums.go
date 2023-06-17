package enums

type Environment string

const (
	DEVELOPMENT Environment = "DEVELOPMENT"
	PRODUCTION  Environment = "PRODUCTION"
)

func (e Environment) String() string {
	return string(e)
}
