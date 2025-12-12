package domain

type PostgresSettings struct {
	User       string
	Password   string
	Host       string
	Port       string
	DBName     string
	SSlEnabled bool
}

func (p *PostgresSettings) GetUrl() string {
	result := "postgres://" + p.User + ":" + p.Password + "@" + p.Host + ":" + p.Port + "/" + p.DBName
	if p.SSlEnabled == false {
		result += "?sslmode=disable"
	}

	return result
}
