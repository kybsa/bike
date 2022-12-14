package config

type SimpleConfig struct {
	MapConfig map[string]string
}

func (appConfig *SimpleConfig) Get(key string) (string, bool) {
	value, ok := appConfig.MapConfig[key]
	return value, ok
}
