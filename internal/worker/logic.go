package worker

import(
"context"
"fmt"
"log"

"servers_updater/internal/domain"
"servers_updater/internal/db"
"servers_updater/internal/notify"
"servers_updater/internal/security"
"servers_updater/internal/ssh"
)

func ProcessSingleServer(ctx context.Context, host string, database *db.BoltDB, client domain.SSHClient) error {
 defer client.Close()

 log.Printf("[%s] Initiating audit...", host)

 output, err := client.ExecuteCommand("sudo apt-get upgrade -s")
 if err != nil {
  msg := fmt.Sprintf("failure in the dry run in %s: %v", host, err)
  sendErrorNotification(host, database, msg)
  return err
 }
 
 packages := ssh.ParseAptOutput(output)
 if len(packages) = 0 {
  log.Printf("[%s] OK, no packages to update.", host)
  return nil
 }

 for _, pkg := range packages {
  isVulnerable, info := security.CheckOSV(pkg.Name, pkg.NewVersion)
  if isVulnerable {
   msg := fmt.Sprintf("VULNERABILITY: %s (Version: %s) - Details: %s", pkg.Name, pkg.NewVersion, info)
   sendVulnerabilityNotification(host, database, packages, msg)

   return fmt.Errorf("Secutiry: %s is vulnerable", pkg.Name)
  }
 }

 log.Printf("[%s] No risks found. Applying packages...",host)
 _, err = client.ExecuteCommand("sudo DEBIAN_FRONTEND=noninteractive apt-get upgrade -y")
 if err != nil {
  sendErrorNotification(host, database, "Upgrade error:"+err.Error())
  return err
}
 
 log.Printf("[%s] System updated successfully.", host)
 return nil
}

func sendErrorNotification(host string, database *db.BoltDB, msg string) {
 cfg, _ := database.GetSMTPConfig()
 notify.SendHTMLEmail(cfg, domain.EmailData{Host: host, Status: "Error", ErrorMessage: msg})
}

func sendVulnerabilityNotification(host string, database *db.BoltDB, pkgs []domain.Package, msg string) {
 cfg, _ := database.GetSMTPConfig()
 notify.SendHTMLEmail(cfg, domain.EmailData{Host: host, Status: "Vulnerability", Packages: pkgs, ErrorMessage: msg})
}






