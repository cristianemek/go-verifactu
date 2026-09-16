package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/cristianemek/go-verifactu/record"
)

type TenantConfig struct {
	Nombre          string `json:"nombre"`
	Token           string `json:"token"`
	Certificado     string `json:"certificado"`
	TipoCertificado string `json:"tipo_certificado"`
}
type Config struct {
	Listen       string                    `json:"listen"`
	Data         string                    `json:"data"`
	Entorno      string                    `json:"entorno"`
	RemisionCada string                    `json:"remision_cada"`
	Sistema      record.SistemaInformatico `json:"sistema"`
	Tenants      map[string]TenantConfig   `json:"tenants"`
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

	if cfg.RemisionCada == "" {
		cfg.RemisionCada = "60s"
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

	if len(cfg.Tenants) == 0 {
		return nil, fmt.Errorf("missing tenants in config")
	}

	tenants := make(map[string]TenantConfig, len(cfg.Tenants))

	for nif, t := range cfg.Tenants {
		if t.Token == "" {
			return nil, fmt.Errorf("missing token for tenant %s", nif)
		}

		if t.Certificado == "" {
			return nil, fmt.Errorf("missing certificado for tenant %s", nif)
		}

		if t.TipoCertificado == "" {
			return nil, fmt.Errorf("missing tipo_certificado for tenant %s", nif)
		}

		nif = strings.ToUpper(nif)

		if _, exists := tenants[nif]; exists {
			return nil, fmt.Errorf("duplicate tenant NIF: %s", nif)
		}

		tenants[nif] = t

	}

	cfg.Tenants = tenants

	return &cfg, nil
}
