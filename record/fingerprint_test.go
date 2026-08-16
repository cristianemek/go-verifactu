package record

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	testDataDir = "../testdata/vectors/huella"
)

type vector struct {
	Descripcion string       `json:"descripcion"`
	Tipo        string       `json:"tipo"`
	Campos      vectorCampos `json:"campos"`
	Cadena      string       `json:"cadena"`
	Huella      string       `json:"huella"`
}

type vectorCampos struct {
	IDEmisorFactura        string `json:"IDEmisorFactura"`
	NumSerieFactura        string `json:"NumSerieFactura"`
	FechaExpedicionFactura string `json:"FechaExpedicionFactura"`
	TipoFactura            string `json:"TipoFactura"`
	CuotaTotal             string `json:"CuotaTotal"`
	ImporteTotal           string `json:"ImporteTotal"`

	IDEmisorFacturaAnulada        string `json:"IDEmisorFacturaAnulada"`
	NumSerieFacturaAnulada        string `json:"NumSerieFacturaAnulada"`
	FechaExpedicionFacturaAnulada string `json:"FechaExpedicionFacturaAnulada"`

	FechaHoraHusoGenRegistro string `json:"FechaHoraHusoGenRegistro"`
	PreviousHash             string `json:"Huella"`
}

func TestFingerPrintVectors(t *testing.T) {
	entries, err := os.ReadDir(testDataDir)
	if err != nil {
		t.Fatalf("Error reading test data directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			aeatExampleFilesPath := filepath.Join(testDataDir, entry.Name())

			v := readAndUnmarshalVector(t, aeatExampleFilesPath)

			t.Logf("description: %s, type: %s", v.Descripcion, v.Tipo)

			var got string

			switch v.Tipo {
			case "alta":
				expeditionDate, err := time.Parse(fechaFormat, v.Campos.FechaExpedicionFactura)

				if err != nil {
					t.Fatalf("Error parsing FechaExpedicionFactura: %v", err)
				}

				generationDate, err := time.Parse(time.RFC3339, v.Campos.FechaHoraHusoGenRegistro)

				if err != nil {
					t.Fatalf("Error parsing FechaHoraHusoGenRegistro: %v", err)
				}

				taxAmount, err := ParseAmount(v.Campos.CuotaTotal)

				if err != nil {
					t.Fatalf("Error parsing CuotaTotal: %v", err)
				}

				totalAmount, err := ParseAmount(v.Campos.ImporteTotal)

				if err != nil {
					t.Fatalf("Error parsing ImporteTotal: %v", err)
				}

				var rec = RegistrationRecord{
					IDEmisorFactura:          v.Campos.IDEmisorFactura,
					NumSerieFactura:          v.Campos.NumSerieFactura,
					FechaExpedicionFactura:   expeditionDate,
					TipoFactura:              v.Campos.TipoFactura,
					CuotaTotal:               taxAmount,
					ImporteTotal:             totalAmount,
					PreviousHash:             v.Campos.PreviousHash,
					FechaHoraHusoGenRegistro: generationDate,
				}

				t.Logf("record: %+v", rec)

				got = registrationFingerprintInput(rec)

			case "anulacion":
				expeditionDate, err := time.Parse(fechaFormat, v.Campos.FechaExpedicionFacturaAnulada)

				if err != nil {
					t.Fatalf("Error parsing FechaExpedicionFacturaAnulada: %v", err)
				}

				generationDate, err := time.Parse(time.RFC3339, v.Campos.FechaHoraHusoGenRegistro)

				if err != nil {
					t.Fatalf("Error parsing FechaHoraHusoGenRegistro: %v", err)
				}

				var rec = CancellationRecord{
					IDEmisorFacturaAnulada:        v.Campos.IDEmisorFacturaAnulada,
					NumSerieFacturaAnulada:        v.Campos.NumSerieFacturaAnulada,
					FechaExpedicionFacturaAnulada: expeditionDate,
					PreviousHash:                  v.Campos.PreviousHash,
					FechaHoraHusoGenRegistro:      generationDate,
				}

				t.Logf("record: %+v", rec)

				got = cancellationFingerprintInput(rec)

			default:
				t.Fatalf("Unknown vector type: %s", v.Tipo)
			}

			assertGotAndVector(t, got, v)
		})
	}
}

func assertGotAndVector(t *testing.T, got string, v vector) {
	t.Helper()

	if got != v.Cadena {
		t.Errorf("Fingerprint input mismatch:\nGot:  %s\nWant: %s", got, v.Cadena)
	}

	gotHash := hashFingerprintInput(got)

	if gotHash != v.Huella {
		t.Errorf("Fingerprint hash mismatch:\nGot:  %s\nWant: %s", gotHash, v.Huella)
	}

	if len(gotHash) != 64 {
		t.Errorf("Fingerprint hash length mismatch:\nGot:  %d\nWant: 64", len(gotHash))
	}

	if strings.ToUpper(gotHash) != gotHash {
		t.Errorf("Fingerprint hash is not uppercase:\nGot:  %s\nWant: %s", gotHash, strings.ToUpper(gotHash))
	}
}

func readAndUnmarshalVector(t *testing.T, filePath string) vector {
	t.Helper()

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Error reading test data: %v", err)
	}
	var v vector
	err = json.Unmarshal(data, &v)

	if err != nil {
		t.Fatalf("Error parsing JSON: %v", err)
	}
	return v
}
