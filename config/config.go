// Package config contains config features
package config

type ConfigComponent interface {
	Get(key string) (string, bool)
}
