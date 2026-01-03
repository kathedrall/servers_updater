package templates

import (
 "bytes"
 "servers_updater/internal/domain"
 "os"
 "testing"
 "time"
)

func TestRenderEmailReport(t *testing.T) {
 now := time.Now().UTC()
 data := domain.EmailData{
  Host: "srv-production-db-01",
  Status: "Vulnerabilidade",
  Date: &now,
  Packages: []domain.Package{
    {Name: "openssl", CurrentVersion: "1.1.1f", NewVersion: "1.1.1n"},
    {Name: "libc6",   CurrentVersion: "2.31-13", NewVersion: "2.31-15"},
    {Name: "linux-image-amd64", CurrentVersion: "5.10.0-8", NewVersion: "5.10.0-10"},
  },
  ErrorMessage: "Warning: 3 critical packages detected in the OSV.dev database.",
}

 var buf bytes.Buffer
 err := RenderEmail(&buf, "report", data)
 if err != nil {
  t.Fatalf("Error rendering template: %v", err)
 }

 if buf.Len() == 0 {
  t.Error("The output buffer is empty; rendering failed.")
 }
 
 err = os.WriteFile("debug_report.html", buf.Bytes(), 0644)
 if err != nil {
  t.Logf("Could not create debug file: %v", err)
 }

 t.Log("Rendering test complete. Please check 'debug report.html' to validate the visual appearance.")
}
