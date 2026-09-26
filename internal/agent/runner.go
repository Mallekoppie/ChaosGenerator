package agent

import (
	"context"
	"log"
	"time"
)

const (
	// initialReconnectDelay is the first pause after a failed register/connect.
	initialReconnectDelay = time.Second
	// maxReconnectDelay caps the exponential backoff.
	maxReconnectDelay = 30 * time.Second
	// stableConnectionDuration is how long a connection must last before the
	// backoff is reset. It stops a flapping master from hammering the server.
	stableConnectionDuration = 30 * time.Second
)

// Run keeps the agent registered and connected to the master. If registration
// or the stream fails it retries with exponential backoff until ctx is
// cancelled, so a master restart or a dropped connection no longer strands the
// agent forever.
func Run(ctx context.Context, masterAddress, port, metricsPort, version string) {
	delay := initialReconnectDelay

	for {
		if ctx.Err() != nil {
			return
		}

		agentId, err := RegisterWithMaster(masterAddress, port, metricsPort, version)
		switch {
		case err != nil:
			log.Printf("Failed to register with master: %v", err)
		case agentId == "":
			log.Printf("Master rejected the registration request")
		default:
			log.Printf("Registered with master. Agent ID: %s", agentId)

			client := NewClient(masterAddress, agentId)
			startedAt := time.Now()

			// Blocks until the stream ends.
			err = client.ConnectAndServe(ctx)

			if ctx.Err() != nil {
				return
			}

			if time.Since(startedAt) >= stableConnectionDuration {
				delay = initialReconnectDelay
			}

			if err != nil {
				log.Printf("Agent disconnected from master: %v", err)
			} else {
				log.Printf("Agent connection to master closed")
			}
		}

		log.Printf("Reconnecting in %s", delay)

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}

		delay *= 2
		if delay > maxReconnectDelay {
			delay = maxReconnectDelay
		}
	}
}
