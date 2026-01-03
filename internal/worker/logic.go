package worker

import(
"context"
"fmt"
"log"

"servers_updater/internal/db"
"servers_updater/internal/notify"
"servers_updater/internal/security"
"servers_updater/internal/ssh"
)

func ProcessSingleServer(ctx context.Context, host string, database *db.BoltDB) error {
 client, err := ssh.Connect(host, database)
 if err != nil {
  msg := fmt.Sprinf("Failed to connect to SSH.: %v, err")
  notify.SendEmail(host, "Connection err", msg)
  return err
 }
 defer client.Close()

 packages, err := ssh.GetUpgradablePackages(client)
 if err != nil {
  notify.SendEmail(host, "Dry-Run error", err.Error())
  return err
 }

 if len(packages) == 0 {
  log.Printf("[%s] System already updated.", host)
  return nil
 }

 for _, pkg := range packages {
  isVulnerable, info := security.CheckOS(pkg.Name, pkg.Version)
  if isVulnerable {
	  msg := fmt.Sprintf("Update aborted! The package %s (%s)  has risks: %s", pkg.Name, pkgVersion, info) 
  
   notify.SendEmail(host, "Security Risk", msg)
   return fmt.Errorf("Vulnerability found in %s", pkg.Name)
  }
 }

 err = ssh.RunFinalUpgrade(client)
 if err = != nil {
  notify.SendEmail(host, "Package installation failed.", err.Error())
  return err
 }

 log.Printf("[%s] Updated packages successful.")
 return nil
}






