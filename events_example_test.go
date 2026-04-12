package sumup_test

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sumup/sumup-go"
)

func ExampleClient_EventsHandler() {
	c := sumup.NewClient()
	h, err := c.EventsHandler(os.Getenv("SUMUP_EVENT_SECRET"), func(_ context.Context, event sumup.EventNotification) error {
		log.Printf("unhandled event: %s", event.EventType())
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := h.OnMemberUpdated(func(ctx context.Context, event *sumup.MemberUpdatedEvent) error {
		member, err := event.FetchObject(ctx)
		if err != nil {
			return err
		}
		log.Printf("member updated: %s", member.ID)
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	// In an HTTP receiver, limit and read the raw body, then call
	// h.Handle(r.Context(), body, r.Header.Get(sumup.EventSignatureHeader)).
	// Acknowledge only on success.
}

func ExampleClient_ParseEventNotificationWithoutVerification() {
	c := sumup.NewClient()
	event, err := c.ParseEventNotificationWithoutVerification([]byte(`{"id":"evt_123","type":"members.updated","created_at":"2026-04-11T10:00:00Z","object":{"id":"member_123","type":"member","url":"https://api.sumup.com/v0.1/merchants/MCODE/members/member_123"}}`))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(event.EventID(), event.EventType())
	// Output: evt_123 members.updated
}
