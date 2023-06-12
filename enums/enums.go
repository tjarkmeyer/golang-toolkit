package enums

type Environment string

const (
	DEVELOPMENT Environment = "DEVELOPMENT"
	PRODUCTION  Environment = "PRODUCTION"
)

func (env Environment) String() string {
	environments := [...]string{"DEVELOPMENT", "PRODUCTION"}

	value := string(env)

	for _, v := range environments {
		if v == value {
			return value
		}
	}

	return ""
}
