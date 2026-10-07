CREATE TABLE events(
    id SERIAL PRIMARY KEY, 
    venue TEXT NOT NULL,
    name TEXT NOT NULL,
    capacity INT NOT NULL,
    available INT NOT NULL
);

-- events
-- id | venue        | name         | capacity | available
-- ---+--------------+--------------+----------+-----------
-- 1  | PVR Orion    | Avengers     | 200      | 200
-- 2  | PVR Orion    | Interstellar | 150      | 150
-- 3  | INOX Garuda  | Dune         | 180      | 180

CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL,
    event_id INT NOT NULL REFERENCES events(id),
    booked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);