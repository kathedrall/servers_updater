package worker

import (
	"context"
	"fmt"
	"log"

	"servers_updater/internal/db"
	"servers_updater/internal/domain"
	"servers_updater/internal/notify"
	"servers_updater/internal/pkgmanager"

	"servers_updater/internal/security"
)

func ProcessSingleServer(ctx context.Context, host string, database *db.BoltDB, client domain.SSHClient) error {
	defer client.Close()

	log.Printf("[%s] Initiating audit...", host)

	var machine domain.Machine
	machine.Host = host
	if err := pkgmanager.DetectOs(client, &machine); err != nil {
		e := fmt.Errorf("Failed to identify the OS: %v", err)
		return e
	}

	manager, err := pkgmanager.GetManager(machine.OSName)
	if err != nil {
		log.Printf("[%s] %v (OS: %s)", host, err, machine.OSName)
		return nil
	}

	output, err := client.ExecuteCommand(manager.GetCheckCommand())
	if err != nil && machine.OSName != "centos" && machine.OSName != "rhel" {
		msg := fmt.Sprintf("failure in the dry run in %s: %v", host, err)
		sendErrorNotification(host, database, msg)
		return err
	}

	packages, err := manager.ParseOutput(output)
	if err != nil {
		e := fmt.Errorf("Error no parse: %v", err)
		return e
	}
	if len(packages) == 0 {
		log.Printf("[%s] System is up-to-date (%s).", host, machine.PrettyName)
		return nil
	}

	for _, pkg := range packages {
		isVulnerable, info := security.CheckOSV(pkg.Name, pkg.NewVersion)
		if isVulnerable {
			msg := fmt.Sprintf("VULNERABILITY: %s (Version: %s) - Details: %s", pkg.Name, pkg.NewVersion, info)
			sendVulnerabilityNotification(host, database, packages, msg)

			return fmt.Errorf("Secutiry Risk detected %s ", pkg.Name)
		}
	}

	log.Printf("[%s] Updating %d packages via %T...", host, len(packages), manager)
	_, err = client.ExecuteCommand(manager.GetUpdateCommand())
	if err != nil {
		sendErrorNotification(host, database, "Upgrade error:"+err.Error())
		return err
	}

	log.Printf("[%s] System updated successfully.", host)
	return nil
}

func sendErrorNotification(host string, database *db.BoltDB, msg string) {
	cfg, err := database.GetSMTPConfig()
	if err != nil {
		return
	}

	recipients, err := database.ListRecipient()
	if err != nil || len(recipients) == 0 {
		return
	}

	notify.SendHtmlEmail(
		cfg,
		recipients,
		domain.EmailData{
			Host:         host,
			Status:       "Error",
			ErrorMessage: msg,
		})
}

func sendVulnerabilityNotification(host string, database *db.BoltDB, pkgs []domain.Package, msg string) {
	cfg, err := database.GetSMTPConfig()
	if err != nil {
		return
	}

	recipients, err := database.ListRecipient()
	if err != nil || len(recipients) == 0 {
		return
	}

	notify.SendHtmlEmail(
		cfg,
		recipients,
		domain.EmailData{
			Host:         host,
			Status:       "Vulnerability",
			Packages:     pkgs,
			ErrorMessage: msg,
		})
}
