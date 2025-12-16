package domain

// PostgresSettings contains configuration parameters for PostgreSQL database connection.
type PostgresSettings struct {
	User       string
	Password   string
	Host       string
	Port       string
	DBName     string
	SSlEnabled bool
}

// GetUrl constructs and returns a PostgreSQL connection URL string.
// The URL format is: postgres://user:password@host:port/dbname
// If SSlEnabled is false, "?sslmode=disable" is appended to the URL.
// This method does not return any errors; it assumes all fields are properly set.
func (p *PostgresSettings) GetUrl() string {
	result := "postgres://" + p.User + ":" + p.Password + "@" + p.Host + ":" + p.Port + "/" + p.DBName
	if p.SSlEnabled == false {
		result += "?sslmode=disable"
	}

	return result
}
