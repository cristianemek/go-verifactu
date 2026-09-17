package memory

import (
	"testing"

	"github.com/cristianemek/go-verifactu"
	"github.com/cristianemek/go-verifactu/store/storetest"
)

func TestConformidad(t *testing.T) {
	storetest.Conformidad(t, func(t *testing.T) verifactu.Store {
		return New()
	})
}
