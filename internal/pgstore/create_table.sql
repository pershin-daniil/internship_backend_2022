CREATE TABLE users
(
    user_id         serial PRIMARY KEY,
    account_balance int NOT NULL,
    reserved        int
);

CREATE TABLE history
(
    user_id    int  NOT NULL REFERENCES users (user_id),
    service_id int  NOT NULL, -- probably references services or something else
    order_id   int  NOT NULL, -- same for orders
    price      int  NOT NULL,
    status     bool NOT NULL DEFAULT FALSE
);