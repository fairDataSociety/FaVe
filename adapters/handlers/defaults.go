package handlers

var (
	DEFAULT_BEE_API  = "http://localhost:1633"
	DEFAULT_STAMP_ID = "0000000000000000000000000000000000000000000000000000000000000000"
)

func SetDefaults(config *HandlerConfig) {
	if config.BeeAPI == "" {
		config.BeeAPI = DEFAULT_BEE_API
	}

	if config.StampId == "" {
		config.StampId = DEFAULT_STAMP_ID
	}
}
