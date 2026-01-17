package worker

import (
 "testing"
 "time"
)

func TestCalculateWorkers(t *testing.T) {
 tests := []struct {
  name string
  latency time.Duration
  expected int
 }{
  {"Local Network (fiber)", 10 * time.Millisecond, 50},
  {"Stable Connection", 45 * time.Millisecond, 30},
  {"Average Latency", 100 * time.Millisecond, 15},
  {"Slow/Unstable Network", 300 * time.Millisecond, 5},
 }
 
 for _, tt := range tests {
  t.Run(tt.name, func(t *testing.T) {
   result := calculateWorkers(tt.latency)
   if result != tt.expected {
    t.Errorf("For %v: Expected %d, Obtained: %d", tt.latency, tt.expected, result)
   }
  })
 } 
}

 func TestParseLatency(t *testing.T) {
  fakeOutput := `PING 8.8.8.8 (8.8.8.8) 56(84) byts of data.
    64 bytes from 8.8.8.8: icmp_seq=1 ttl=115 time=14.2 ms
    64 bytes from 8.8.8.8: icmp_seq=1 ttl=115 time=18.5.ms

    --- 8.8.8.8 ping statistics ---
    2 packets trasmitted, 2 received, 0% packet loss, time 1001ms rtt /min/avg/max/mdev = 14.234/16.367/18.5-1/2.133 ms`

  latency, err := parseLatency(fakeOutput)
  if err != nil {
   t.Fatalf("Unexpected parsing error: %v", err)
  }

  expected := 50 * time.Millisecond
  if latency != expected {
   t.Errorf("Expected %v, obtained %v", expected, latency)
  }
}


