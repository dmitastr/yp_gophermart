-- Table: public.orders

-- DROP TABLE IF EXISTS public.orders;

CREATE TABLE IF NOT EXISTS orders
(
    username VARCHAR(50) NOT NULL,
    order_id VARCHAR(50) NOT NULL,
    status VARCHAR(50),
    accrual double precision,
    uploaded_at timestamp with time zone,
    CONSTRAINT orders_pkey PRIMARY KEY (order_id, username),
    CONSTRAINT order_id_unique UNIQUE (order_id),
    CONSTRAINT username_fk FOREIGN KEY (username)
        REFERENCES users (username) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

ALTER TABLE IF EXISTS orders
    OWNER to postgres;

-- Table: public.withdrawals

-- DROP TABLE IF EXISTS public.withdrawals;

CREATE TABLE IF NOT EXISTS withdrawals
(
    username VARCHAR(50) NOT NULL,
    order_id VARCHAR(50) NOT NULL,
    sum double precision,
    processed_at timestamp with time zone,
    CONSTRAINT withdrawals_pkey PRIMARY KEY (username, order_id),
    CONSTRAINT username_fk FOREIGN KEY (username)
        REFERENCES users (username) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);

ALTER TABLE IF EXISTS withdrawals
    OWNER to postgres;

-- View: public.balance

-- DROP VIEW public.balance;

CREATE OR REPLACE VIEW balance
AS
SELECT d.username,
       COALESCE(d.debit, 0) - COALESCE(c.credit, 0) AS current,
       c.credit AS withdrawn
FROM ( SELECT orders.username,
              sum(orders.accrual) AS debit
       FROM orders
       GROUP BY orders.username) d
         LEFT JOIN ( SELECT withdrawals.username,
                            sum(withdrawals.sum) AS credit
                     FROM withdrawals
                     GROUP BY withdrawals.username) c ON d.username = c.username;

ALTER TABLE balance
    OWNER TO postgres;

