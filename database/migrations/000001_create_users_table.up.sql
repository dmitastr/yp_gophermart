-- Table: public.users

-- DROP TABLE IF EXISTS public.users;

CREATE TABLE IF NOT EXISTS users
(
    username VARCHAR(50) NOT NULL,
    password VARCHAR(50) NOT NULL,
    created_at timestamp with time zone,
    CONSTRAINT users_pkey PRIMARY KEY (username)
);

ALTER TABLE IF EXISTS public.users
    OWNER to postgres;