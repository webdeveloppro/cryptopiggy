create or replace function blocknotify() RETURNS trigger AS $$
DECLARE
  notification json;
begin
  notification = json_build_object(
    'id', new.id,
    'hash', new.hash,
    'bits', new.bits,
    'height', new.height,
    'nonce', new.nonce,
    'version', new.version,
    'hash_prev_block', new.hash_prev_block,
    'hash_merkle_root', new.hash_merkle_root,
    'created_at', new.created_at
  );
  PERFORM pg_notify('blocks_notify',notification::text);
  RETURN NEW; 
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER blocks_notify ON block;
CREATE TRIGGER blocks_notify AFTER INSERT ON block FOR EACH ROW EXECUTE PROCEDURE blocknotify();

create or replace function addressnotify() RETURNS trigger AS $$
DECLARE
  notification json;
begin
  notification = json_build_object(
    'id', new.id,
    'hash', new.hash,
    'income', new.income,
    'outcome', new.outcome,
    'ballance', new.ballance,
    'updated_at', new.updated_at
  );
  PERFORM pg_notify('address_notify',notification::text);
  RETURN NEW; 
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER address_notify ON address;
CREATE TRIGGER address_notify AFTER INSERT ON address FOR EACH ROW EXECUTE PROCEDURE addressnotify();