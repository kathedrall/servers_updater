package security

 import (
  "bytes"
  "encoding/json"
  "net/http"
  "time"
 )

 const (
  ECOSYSTEM = "Debian:11"
  URL = "http://api.osv.dev/v1/query"
 )

 type osvQuery struct {
  Version string `json:"veersion"`
  Package struct {
   Name string `json:"name"`
   Ecosystem string `json:"ecosystem"`
  } `json:"package"`
 }

 func CheckOSV(pkgName string, version string) (bool, string) {
  query := osvQuery {Version: version}
  query.Package.Name = pkgName
  query.Package.Ecosystem = ECOSYSTEM

  body, _ := json.Marshal(query)
  
  client := &http.Client{Timeout: 5 * time.Second}
  resp, err := client.Post(URL, "application/json", bytes.NewBuffer(body))
   if err != nil {
    e := "API connection failed."
    return false, e
   }
   defer resp.Body.Close()

   var result map[string]interface{}
   if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
    e := "Error processing API response."
    return false , e
   }

   if vulns, ok := result["vulns"]; ok && vulns != nil {
    v := "This package contains critical vulnerabilities listed in osv.dev."
    return true, v
   }
  return false, ""
}
