package config

import "testing"

func Test_GivenConfig_WhenGet_ThenReturnExpectedValue(t *testing.T) {
	// Given
	simpleConfig := SimpleConfig{
		MapConfig: map[string]string{"key": "value"},
	}
	// When
	actual := simpleConfig.Get("key")
	// Then
	if actual != "value" {
		t.Error("Get must return value")
	}
}
