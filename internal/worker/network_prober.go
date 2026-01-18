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

var pingRegex = regexp.MustCompile(`(?:rtt|round-trip)\s+min/avg/max(?:/[a-z]+)?\s*=\s*[\d\.]+/([\d\.]+)/`)

type NetworkProfile struct {
 RTT              time.Duration
 SuggestedWorkers int
}

func AnalyzeNetwork(ctx context.Context, target string) (NetworkProfile, error) {
 if target == "" {
  target = TARGET  
 }

 cmd := exec.CommandContext(ctx, "ping", "-c", "3", "-q", target)
 output , err := cmd.CombinedOutput()
 if err != nil {
  return NetworkProfile {
   RTT:              100 * time.Millisecond,
   SuggestedWorkers: 5,
  }, nil
 } 

 rtt, err := parseLatency(string(output))
 if err != nil {
  return NetworkProfile{
   RTT:              100 * time.Millisecond,
   SuggestedWorkers: 5,
  }, nil
 }
 workers := calculateWorkers(rtt)
  return NetworkProfile{
    RTT:              rtt,
    SuggestedWorkers: workers,
  }, nil
}


func parseLatency(output string) (time.Duration, error) {
 matches := pingRegex.FindStringSubmatch(output)
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

