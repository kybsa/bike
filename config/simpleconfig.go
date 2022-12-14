package config

type SimpleConfig struct {
	MapConfig map[string]string
}

func (appConfig *SimpleConfig) Get(key string) string {
	return appConfig.MapConfig[key]
}
