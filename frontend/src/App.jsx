import './App.css'
import { useEffect, useState } from "react";

const API = "http://localhost:8080";

function App() {
  const [events, setEvents] = useState([]);
  const [username, setUsername] = useState("");
  const [eventId, setEventId] = useState("");
  const [message, setMessage] = useState("");

  useEffect(() => {
    fetch(`${API}/events`)
      .then((res) => res.json())
      .then(setEvents)
      .catch(() => setMessage("Failed to load events"));
  }, []);

  async function bookTicket(e) {
    e.preventDefault();

    const res = await fetch(`${API}/bookings`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        username,
        event_id: Number(eventId),
      }),
    });

    const text = await res.text();
    setMessage(text);

    if (res.ok) {
      setUsername("");
      setEventId("");
      // Refresh availability
      const events = await fetch(`${API}/events`).then((r) => r.json());
      setEvents(events);
    }
  }


  return (
    <>
      <main>
        <form onSubmit={bookTicket}>
          <h1>Ticket Booking</h1>

          <input
            placeholder="Username"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
          />

          <select
            value={eventId}
            onChange={(e) => setEventId(e.target.value)}
            required
          >
            <option value="">Select event</option>

            {events.map((event) => (
              <option key={event.id} value={event.id}>
                {event.venue} — {event.name} ({event.available} left)
              </option>
            ))}
          </select>

          <button type="submit">Book Ticket</button>

          {message && <p>{message}</p>}
        </form>
      </main>
    </>
  )
}

export default App
