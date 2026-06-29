package config

import (
	"strings"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type (
	HTTP struct {
		Addr               string
		CORSAllowedOrigins []string
	}
	Config struct {
		HTTP        HTTP
		PostgresDSN string
	}
)

func Load() (Config, error) {
	var k = koanf.New("")
	k.Load(file.Provider(".env"), dotenv.Parser())
	return Config{
		HTTP: HTTP{
			Addr:               k.MustString("HTTP_ADDR"),
			CORSAllowedOrigins: strings.Split(k.String("CORS"), ","),
		},
		PostgresDSN: k.MustString("GOOSE_DBSTRING"),
	}, nil
}
