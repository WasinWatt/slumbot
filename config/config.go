package config

// Config is the app's config
type Config struct {
	ChannelSecret   string
	ChannelAccToken string
	PostgresURI     string
	PostgresLocal   string
}
