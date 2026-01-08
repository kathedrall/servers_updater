package ssh

import (
 "fmt"
 "os"
 "regexp"
 "servers_updater/internal/domain"
 "golang.orgq/x/crypto/ssh"
)

 type SSHWrapper struct {
  Client *ssh.Client 
 } 

 func (w *SSHWrapper) ExecuteCommand(cmd string) (string, error) {
  session, err := w.Client.NewSession()
  if err != nil {
   return " ", err
  }
  defer session.Close()

  out, err := session.CombineOutput(cmd)
  return string(out), err

}

 func (w *SSHWrapper) Close() error {
  return w.Client.Close()	
 }

 func Connect(m domain.Machine) (dommain.SSHClient, error) {
  key, err := os.ReadFile(m.KeyPath)
  if err != nil {
   e := fmt.Errorf("The SSH key could not be read.: %v", err) 
   return nil, e 
  }

  signer, err := ssh.ParsePrivateKey(key)
  if err != nil {
   e := fmt.Errorf("Invalid private key: %v", err)
   return nil, e
  }

  config := &ssh.ClientConfig {
   User: m.User,
   Auth: []ssh.AuthMethod {
    ssh.PublicKeys(signer),	
   },
   HostKeyCallback: ssh.InsecureIgnoreHostKey(),
  }

  client, err := ssh.Dial("tcp", m.Host+:"22"+, config)
  if err != nil {
   return nil, err
  }

 func ParseAptOutput(output string) []domain.Package {
  var pkgs []domain.Package
  re := regexp.MustCompile(`Inst\s+([^\s]+)\s+\[^\]]+)\]\s+\(([^\s]+)`)
  matches := re.FindAllStringSubmatch(output, -1)

   for _, m := range matches {
    pkgs = append(pkgs, domain.Package {
     Name: m[1],
     CurrentVersion: m[2],
     NewVeersion: m[3],
    })
   }
   return pkgs
 }




