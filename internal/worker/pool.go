package worker

import(
 "context"
 "log"
 "servers_updater/internal/db"
 "servers_updater/internal/ssh"
 "servers_updater/internal/domain"
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

   machine := domain.Machine {
    Host: host,
   }

   client, err := ssh.Connect(machine)
   if err != nil {
    log.Printf("[ERROR] It was not possible to establish a connection with: %s: %v", host, err)
    return nil
   }

   return ProcessSingleServer(ctx, host, database,client)	
  })
 }
 return g.Wait()
}
