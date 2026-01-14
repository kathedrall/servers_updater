package worker

import (
 "context"
 "fmt"
 "os/exec"
 "regexp"
 "strconv"
 "time"
)

const TARGET = "8.8.8.8"

type NetworkProfile struct {
 RTT              time.Duration
 SuggestedWorkers int
}

func AnalyzeNetwork(ctx context.Context, target string) (NetworkProfile, error) {
 if target == "" {
  target = TARGET  
 }

 output,err := runPingCommand(ctx, target)
 if err != nil {
  n := NetworkProfile{}
 return n, err
 }

 avgRTT, err := parseLatency(output)
 if err != nil {
  n := NetworkProfile{}
 return n, err
 }

 np := NetworkProfile {
  RTT              : avgRTT,
  SuggestedWorkers : calculateWorkers(avgRTT),
 }

return np, nil 
}

func runPingCommand(ctx context.Context, target string) (string, error) {
 cmd := exec.CommandContext(ctx, "ping", "-c", "4", "-i", "0-2", target)
 out , err := cmd.CombinedOutput()
 if err != nil {
  n := ""
  e := fmt.Errorf("ping falied: %w", err)
 return n, e
 }
return string(out), nil
}

func parseLatency(output string) (time.Duration, error) {
 re := regexp.MustCompile(`min/avg/max/mdev\s+=\d+\.\d+/(\d+\.\d+)/`)
 matches := re.FindStringSubmatch(output)
 if len(matches) < 2 {
  e := fmt.Errorf("could not parse ping output")
 return 0, e
 }
 avgRTTStr := matches[1]
 avgRTTFloat, err := strconv.ParseFloat(avgRTTStr, 64)
 if err != nil {
  e := fmt.Errorf("Error converting latency value.: %v", err)
  return 0, e
 }
 td := time.Duration(avgRTTFloat * float64(time.Millisecond)) 

return td, nil
}

func calculateWorkers(rtt time.Duration) int {
 ms := rtt.Milliseconds()

 switch {
 case ms < 20:
   return 50
 case ms <  60:
   return 30
 case ms < 150:
   return 15
 default:
   return 5
 }
}

