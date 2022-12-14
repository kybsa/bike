package web

import (
	"strings"

	"github.com/kybsa/bike/config"
)

type GinGonicComponent struct {
	Engine Engine
	Addr   []string
}

type Engine interface {
	Run(addr ...string) error
}

func (webComponent *GinGonicComponent) Start() error {
	return webComponent.Engine.Run(webComponent.Addr...)
}

func NewGinGonicComponent(engine Engine, configComponent config.ConfigComponent) *GinGonicComponent {
	addr := []string{"0.0.0.0:8080"}
	addrConfig, ok := configComponent.Get("GinGonicComponent.Addr")
	if ok {
		addr = strings.Split(addrConfig, ",")
	}
	return &GinGonicComponent{
		Engine: engine,
		Addr:   addr,
	}
}
