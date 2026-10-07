package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func getEvents(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(r.Context(), `
			SELECT id, venue, name, capacity, available
			FROM events
			ORDER BY id
		`)
		if err != nil {
			log.Printf("FAILED TO FETCH: %v", err)
			http.Error(w, "failed to fetch events", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type Event struct {
			ID        int    `json:"id"`
			Venue     string `json:"venue"`
			Name      string `json:"name"`
			Capacity  int    `json:"capacity"`
			Available int    `json:"available"`
		}

		var events []Event

		for rows.Next() {
			var event Event

			if err := rows.Scan(
				&event.ID,
				&event.Venue,
				&event.Name,
				&event.Capacity,
				&event.Available,
			); err != nil {
				log.Printf("FAILED TO READ EVENTS: %v", err)
				http.Error(w, "failed to read events", http.StatusInternalServerError)
				return
			}

			events = append(events, event)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
	}
}

func createBooking(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//w.Write([]byte("booking"))

		var req struct {
			Username string `json:"username"`
			EventID  int    `json:"event_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("ERROR: %v", err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin(r.Context())

		if err != nil {
			log.Printf("ERROR: %v", err)
			http.Error(w, "failed to start transaction", http.StatusInternalServerError)
			return
		}
		defer tx.Rollback(r.Context())

		result, err := tx.Exec(r.Context(),
			`UPDATE events SET available = available -1
		WHERE id=$1
		AND available > 0`,
			req.EventID)

		if err != nil {
			log.Printf("UPDATE FAILED: %v", err)
			http.Error(w, "failed to update event", http.StatusInternalServerError)
			return
		}

		rows := result.RowsAffected()

		if rows == 0 {
			http.Error(w, "event sold out", http.StatusConflict)
			return
		}

		_, err = tx.Exec(r.Context(),

			`INSERT INTO bookings (username,event_id) VALUES ($1,$2)`, req.Username, req.EventID)

		if err != nil {
			log.Printf("INSERT FAILED: %v", err)
			http.Error(w, "Failed to create booking", http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(r.Context()); err != nil {
			log.Printf("COMMIT FAILED: %v", err)
			http.Error(w, "Failed to create booking", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("booking created "))

	}
}
