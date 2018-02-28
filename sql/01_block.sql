DROP TABLE IF EXISTS block CASCADE;
CREATE TABLE block (
  id serial PRIMARY KEY,
  bits int not null default 0,
  height int not null default 0 UNIQUE,
  nonce	 bigint not null default 0,
  version int not null default 0,
  hash_prev_block varchar(64) not null default '',
  hash_merkle_root varchar(64) not null default '',
  created_at timestamp not null default now(),
  hash varchar(64) not null default '' UNIQUE
);

DROP TABLE IF EXISTS transaction CASCADE;
CREATE TABLE transaction (
  id serial PRIMARY KEY,
  hash varchar(64) not null default '' UNIQUE,
  block_id integer references block(id) ON DELETE CASCADE,
  has_witness boolean not null default false
);

DROP TABLE IF EXISTS address CASCADE;
CREATE TABLE address (
  id serial primary key,
  hash varchar(64) not null default '' UNIQUE, 
  income bigint not null default 0,
  outcome bigint not null default 0,
  ballance  bigint not null default 0,
  updated_at timestamp not null default now()
);

DROP TABLE IF EXISTS txin CASCADE;
create table txin (
  id serial primary key,
  transaction_id integer references transaction(id) ON DELETE CASCADE,
  amount  bigint not null default 0,
  address_id integer references address(id) ON DELETE CASCADE DEFAULT NULL,
  prev_out varchar(64) not null default '',
  size int not null default 0,
  signature_script text default null,
  sequence bigint not null default 0,
  witness varchar(256) not null default ''
);

DROP TABLE IF EXISTS txout CASCADE;
CREATE TABLE txout (
  id serial primary key,
  transaction_id integer references transaction(id) ON DELETE CASCADE,
  address_id integer references address(id) ON DELETE CASCADE,
  val bigint not null default 0,
  pk_script  text default null
);

/*
CREATE TABLE address_log (
  id primary key,
  address_id  reference,
  type
  amount
  created_at timestamp
)
*/
