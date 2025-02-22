package config

import (
	"flag"
	"lod2/internal/utils"
)

var Config struct {
	Http struct {
		Host string
		Port int
	}

	// configuration directory used for relatively long-term persistent configuration. read-only.
	ConfigPath string

	// dynamic data directory for things such as DBs and files. read-write.
	DataPath string
}

func Init() {
	flag.StringVar(&Config.Http.Host, "host", "localhost", "host to listen on")
	flag.IntVar(&Config.Http.Port, "port", 10800, "port to listen on")

	flag.StringVar(&Config.ConfigPath, "config", "~/.config/lod2/", "path to configuration directory")
	flag.StringVar(&Config.DataPath, "data", "~/.local/share/lod2/", "path to data directory")

	Config.ConfigPath = utils.ExpandHomePath(Config.ConfigPath)
	Config.DataPath = utils.ExpandHomePath(Config.DataPath)

	flag.Parse()
}
