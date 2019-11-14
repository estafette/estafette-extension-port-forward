package main

import "fmt"

var defaultTimeout = 120
var defaultTillerless = true

type params struct {
	Credentials string `json:"credentials,omitempty" yaml:"credentials,omitempty"`
	Namespace   string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Service     string `json:"service,omitempty" yaml:"service,omitempty"`
	ServicePort string `json:"servicePort,omitempty" yaml:"servicePort,omitempty"`
	LocalPort   string `json:"localPort,omitempty" yaml:"localPort,omitempty"`
}

func (p *params) SetDefaults(releaseTargetName string) {

	if p.LocalPort == "" && p.ServicePort != "" {
		p.LocalPort = p.ServicePort
	}

	// default credentials to release name prefixed with gke if no override in stage params
	if p.Credentials == "" && releaseTargetName != "" {
		p.Credentials = fmt.Sprintf("gke-%v", releaseTargetName)
	}
}
