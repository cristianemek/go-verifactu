package main

import (
	"net/http"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/aeat"
)

type servidor struct {
	engine  *verifactu.Engine
	cliente *aeat.Client
	tenants map[string]TenantConfig
}

func (s *servidor) healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
