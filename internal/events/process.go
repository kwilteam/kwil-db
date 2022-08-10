package events

import (
	"context"
	"fmt"
)

func (e *EventFeed) processEvents(ctx context.Context, ch chan map[string]interface{}) {
	for {
		select {
		case ev := <-ch:
			// Here we can go through and define what we want to do with the event
			if ev["ktype"] == "HelloWorld" {
				fmt.Println("Received HelloWorld event")
			}
		}
	}
}
