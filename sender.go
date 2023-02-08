package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// if we're the leader, sender sends all the entry-s to other replicas
// if we're not the leader, sender does nothing
func SendToReplicas(ctx context.Context, token string) error {
	log := GetLogger(ctx)
	replicasLock.Lock()
	defer replicasLock.Unlock()

	log.Debugf("sending token %s to members - %+v", token, replicas)

	for _, member := range replicas {
		if member.ID == myID {
			continue
		}

		// create io.Reader from token
		buf, err := json.Marshal(token)
		if err != nil {
			return fmt.Errorf("error marshalling token - %s", err)
		}

		resp, err := http.Post(
			fmt.Sprintf("http://%s:%d/token/receive", member.Host, member.Port),
			"application/json",
			bytes.NewBuffer(buf),
		)

		if err != nil {
			fmt.Printf("error sending token to member - %s, token - %s, member - %v\n", err, token, member)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("non-ok response from member - %d, token - %s, member - %v\n", resp.StatusCode, token, member)
			continue
		}
	}

	return nil
}
