DROP TABLE IF EXISTS btc_price;

CREATE TABLE btc_price(
  id serial PRIMARY KEY,
  price decimal(10, 2) not null default 0,
  created_at timestamp not null default now()
);

create index created_at on btc_price(created_at);