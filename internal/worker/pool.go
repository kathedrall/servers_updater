package worker

import(
 "context"
 "servers_updater/internal/db"
 "servers_updater/internal/ssh"
 "golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, hosts []string, database *db.BoltDB) error {
 g, ctx := errgroup.WithContext(ctx)
 g.SetLimit(10)

 for _, h := range hosts {
  host := h
  g.Go(func() error {
   return ssh.ProcessServer(ctx, host, database)	
  })
 }
 return g.Wait()
}
