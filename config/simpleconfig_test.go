package config

import "testing"

func TestGet_GivenConfig_WhenGet_ThenReturnExpectedValue(t *testing.T) {
	// Given
	simpleConfig := SimpleConfig{
		MapConfig: map[string]string{"key": "value"},
	}
	// When
	actual, ok := simpleConfig.Get("key")
	// Then
	if actual != "value" || !ok {
		t.Error("Get must return value")
	}
}
