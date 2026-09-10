package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cristianemek/go-verifactu/record"
)

type TenantConfig struct {
	Nombre string `json:"nombre"`
	Token  string `json:"token"`
}
type Config struct {
	Listen          string                    `json:"listen"`
	Data            string                    `json:"data"`
	Entorno         string                    `json:"entorno"`
	Certificado     string                    `json:"certificado"`
	TipoCertificado string                    `json:"tipo_certificado"`
	RemisionCada    string                    `json:"remision_cada"`
	Sistema         record.SistemaInformatico `json:"sistema"`
	Tenants         map[string]TenantConfig   `json:"tenants"`
}

func cargarConfig(path string) (*Config, error) {
	config, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config

	err = json.Unmarshal(config, &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Listen == "" {
		return nil, fmt.Errorf("missing listen address in config")
	}

	if cfg.Data == "" {
		return nil, fmt.Errorf("missing data path in config")
	}

	if cfg.Entorno == "" {
		return nil, fmt.Errorf("missing entorno in config")
	}

	if cfg.Certificado == "" {
		return nil, fmt.Errorf("missing certificado in config")
	}

	if len(cfg.Tenants) == 0 {
		return nil, fmt.Errorf("missing tenants in config")
	}

	for nif, t := range cfg.Tenants {
		if t.Token == "" {
			return nil, fmt.Errorf("missing token for tenant %s", nif)
		}
	}

	return &cfg, nil
}
