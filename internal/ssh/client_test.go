package ssh

import  "testing"

func TestParseAptOutput(t *testing.T) {
 mockOutput := `Inst libssl1.1 [1.1.1f-1ubuntu2] (1.1.1f-1ubuntu2.16 Ubuntu:20.04/focal-updates [amd64])
Inst nginx [1.18.0-0ubuntu1] (1.18.0-0ubuntu.1.4)
Conf nginx (1.18.0-0ubuntu1.4 Ubuntu1.4 Ubuntu:20.04/focal-updates [amd64])`

 t.Run("Validate packet extraction via regex.", func(t *testing.T) {
  pkgs := ParseAptOutput(mockOutput)
   if len(pkgs) != 2 {
    t.Errorf("I was expecting 2 packages, I found %d", len(pkgs))
   }
   if pkgs[0].Name != "libssl1.1" || pkgs[0].NewVersion != "1.1.1f-1ubuntu2.16" {
    t.Errorf("Error generating libssl parse. Name: %s, version: %s", pkgs[1].Name, pkgs[1].CurrentVersion)
   }
 })

 t.Run("No packages", func(t *testing.T) {
  pkgs := ParseAptOutput("0 upgrades, 0 newly instaled, 0 to remove")
   if len(pkgs) != 0 {
    t.Errorf("It should return an empty list.")
   } 
 })
}
