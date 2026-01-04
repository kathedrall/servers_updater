package worker

import(
 "context"
 "servers_updater/internal/db"
 "servers_updater/internal/ssh"
 "golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, hosts []string, database *db.BoltDB) error {
 g, ctx := errgroup.WithContext(ctx)
 limit, _ := database.GetPoolLimit()
 g.SetLimit(limit)

 for _, h := range hosts {
  host := h
  g.Go(func() error {
   log.Printf("[POOL] Stating Connection with %s...",host)

   ssh, err := ssh.Connect(host,database)
   if err != nil {
    log.Printf("[ERROR] It was not possible to establish a connection with: %s: %v", host, err)
    return nil
   }



   return ProcessSingleServer(ctx, host, database,sshClient)	
  })
 }
 return g.Wait()
}
