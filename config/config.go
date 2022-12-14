package config

type ConfigComponent interface {
	Get(key string) string
}
