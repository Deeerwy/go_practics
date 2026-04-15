package config

type Config struct {
	Addr     string
	CertFile string
	KeyFile  string
	DSN      string
}

func New() Config {
	return Config{
		Addr:     ":8443",
		CertFile: "certs/server.crt",
		KeyFile:  "certs/server.key",
		DSN:      "postgres://postgres:postgres@localhost:5432/study_security?sslmode=disable",
	}
}
