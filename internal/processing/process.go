package processing

import (
	"context"
	"fmt"
)

func (ep *EventProcessor) ProcessEvents(ctx context.Context, ch chan map[string]interface{}) {
	go func() {
		for {
			select {
			case ev := <-ch:
				// Here we can go through and define what we want to do with the event
				switch ev["ktype"] {
				case "NewStake":
					yuh := ev["from"]
					fmt.Println(yuh)
				}
			}
		}
	}()
}
