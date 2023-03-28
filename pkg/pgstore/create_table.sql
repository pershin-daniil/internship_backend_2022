CREATE TABLE wallets
(
    id              serial PRIMARY KEY,
    user_id         int         NOT NULL UNIQUE,
    account_balance int         NOT NULL,
    reserved        int         NOT NULL DEFAULT 0,
    updated_at      timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE events
(
    id         serial PRIMARY KEY,
    wallet_id  int         NOT NULL REFERENCES wallets (id),
    service_id int         NOT NULL,
    order_id   int         NOT NULL UNIQUE,
    price      int         NOT NULL,
    status     varchar     NOT NULL DEFAULT 'REQUESTED',
    datetime   timestamptz NOT NULL DEFAULT NOW()
);