package handlers

import (
	"os"
)

func FromEnv(config *HandlerConfig) {
	if enabled(os.Getenv("VERBOSE")) {
		config.Verbose = true
	}

	if v := os.Getenv("BEE_API"); v != "" {
		config.BeeAPI = v
	}

	if v := os.Getenv("RPC_API"); v != "" {
		config.RPCEndpoint = v
	}

	if v := os.Getenv("STAMP_ID"); v != "" {
		config.StampId = v
	}

	if v := os.Getenv("GLOVE_LEVELDB_URL"); v != "" {
		config.GloveLevelDBUrl = v
	}
	if v := os.Getenv("USER"); v != "" {
		config.Username = v
	}
	if v := os.Getenv("PASSWORD"); v != "" {
		config.Password = v
	}
	if v := os.Getenv("POD"); v != "" {
		config.Pod = v
	}
}

func enabled(value string) bool {
	if value == "" {
		return false
	}

	if value == "on" ||
		value == "enabled" ||
		value == "1" ||
		value == "true" {
		return true
	}

	return false
}
