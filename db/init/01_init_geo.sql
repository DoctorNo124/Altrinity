\connect geodb;

CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS areas (
    id SERIAL PRIMARY KEY,
    name TEXT,
    polygon GEOGRAPHY(POLYGON, 4326)
);

CREATE TABLE IF NOT EXISTS stops (
    id SERIAL PRIMARY KEY,
    area_id INT REFERENCES areas(id),
    name TEXT,
    location GEOGRAPHY(POINT, 4326)
);

CREATE TABLE IF NOT EXISTS assignments (
    id SERIAL PRIMARY KEY,
    volunteer_id UUID,
    stop_id INT REFERENCES stops(id),
    assigned_at TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS volunteer_positions
(
    id SERIAL PRIMARY KEY,
    volunteer_id uuid,
    "position" geography(Point,4326),
    updated_at timestamp without time zone DEFAULT now(),
    full_name text COLLATE pg_catalog."default",
    CONSTRAINT volunteer_positions_pkey PRIMARY KEY (id),
    CONSTRAINT unique_volunteer_id UNIQUE (volunteer_id)
        INCLUDE(volunteer_id)
)

CREATE TABLE IF NOT EXISTS routes (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL,
  points jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);