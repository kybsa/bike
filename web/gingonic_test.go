package web

import (
	"testing"

	"github.com/kybsa/bike/config"
)

type EngineStruct struct {
	callRun bool
}

func (engineStruct *EngineStruct) Run(addr ...string) error {
	engineStruct.callRun = true
	return nil
}

func TestStart_GivenGinGonicComponent_WhenStart_ThenCallRun(t *testing.T) {
	// Given
	engineStruct := &EngineStruct{callRun: false}
	configComponent := config.SimpleConfig{
		MapConfig: map[string]string{"GinGonicComponent.Addr": "0.0.0.0:2020"},
	}

	ginGonicComponent := NewGinGonicComponent(engineStruct, &configComponent)
	// When
	ginGonicComponent.Start()

	// Then
	if !engineStruct.callRun {
		t.Error("Start must call Run")
	}
}
