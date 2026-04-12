// Receive signed SumUp events with typed callbacks.
//
// Set SUMUP_EVENT_SECRET and SUMUP_API_KEY, then run:
//
//	go run ./example/events
//
// Send notifications to POST /events. Deploy behind HTTPS in production.
package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sumup/sumup-go"
)

func main() {
	c := sumup.NewClient()
	handler, err := c.EventsHandler(os.Getenv("SUMUP_EVENT_SECRET"), handleUnhandled)
	if err != nil {
		log.Fatal(err)
	}
	if err := handler.OnMemberUpdated(handleMemberUpdated); err != nil {
		log.Fatal(err)
	}
	if err := handler.OnReaderCreated(handleReaderCreated); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /events", func(w http.ResponseWriter, r *http.Request) {
		// The HTTP receiver owns the body limit; the SDK accepts raw bytes.
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			var sizeError *http.MaxBytesError
			if errors.As(err, &sizeError) {
				http.Error(w, "event too large", http.StatusRequestEntityTooLarge)
			} else {
				http.Error(w, "invalid event body", http.StatusBadRequest)
			}
			return
		}
		if err := handler.Handle(r.Context(), body, r.Header.Get(sumup.EventSignatureHeader)); err != nil {
			var callbackError *sumup.EventCallbackError
			switch {
			case errors.As(err, &callbackError):
				log.Printf("event processing failed: %v", err)
				http.Error(w, "event processing failed", http.StatusInternalServerError)
			case errors.Is(err, sumup.ErrEventSignatureInvalid), errors.Is(err, sumup.ErrEventTimestampInvalid), errors.Is(err, sumup.ErrEventSignatureExpired):
				http.Error(w, "invalid event signature", http.StatusUnauthorized)
			default:
				http.Error(w, "invalid event request", http.StatusBadRequest)
			}
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	server := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: time.Minute}
	log.Println("Listening on http://localhost:8080/events")
	log.Fatal(server.ListenAndServe())
}

func handleMemberUpdated(ctx context.Context, event *sumup.MemberUpdatedEvent) error {
	// In production, deduplicate by event.ID and make updates idempotent.
	member, err := event.FetchObject(ctx)
	if err != nil {
		return err
	}
	log.Printf("member updated: event=%s member=%s", event.ID, member.ID)
	return nil
}

func handleReaderCreated(ctx context.Context, event *sumup.ReaderCreatedEvent) error {
	reader, err := event.FetchObject(ctx)
	if err != nil {
		return err
	}
	log.Printf("reader created: event=%s reader=%s", event.ID, reader.ID)
	return nil
}

func handleUnhandled(_ context.Context, event sumup.EventNotification) error {
	// Both known events without a callback and future event types arrive here.
	log.Printf("unhandled event: id=%s type=%s", event.EventID(), event.EventType())
	return nil
}
