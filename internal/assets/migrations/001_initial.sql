-- +migrate Up

drop function if exists trigger_set_updated_at cascade;
CREATE FUNCTION trigger_set_updated_at() RETURNS trigger
    LANGUAGE plpgsql
AS $$ BEGIN NEW.updated_at = NOW() at time zone 'utc'; RETURN NEW; END; $$;

create domain int_256 as numeric not null
    check (value > -(2 ^ 256) and value < 2 ^ 256)
    check (scale(value) = 0);

create table if not exists transactions(
    hash bytea primary key,
    block_height bigint,
    index integer,
    raw_tx bytea,
    tx_result jsonb,
    tx_timestamp timestamp without time zone not null default now(),
    created_at timestamp without time zone not null default now()
);

create table if not exists transfers(
    id bigserial primary key,
    index bytea unique,
    status integer not null default 0,
    created_at timestamp without time zone not null default now(),
    updated_at timestamp without time zone not null default now(),
    creator text,
    rarimo_tx bytea references transactions(hash),
    rarimo_tx_timestamp timestamp without time zone not null default now(), -- for sorting purposes without joining
    -- below are transfer details fields
    origin text not null,
    tx bytea not null,
    event_id bigint not null,
    from_chain text not null,
    to_chain text not null,
    receiver text not null,
    amount int_256 not null,
    bundle_data bytea,
    bundle_salt bytea,
    token_index text not null
);

create table if not exists votes(
    id bigserial primary key,
    transfer_index bytea,
    choice integer not null default 0,
    rarimo_transaction bytea references transactions(hash),
    created_at timestamp without time zone not null default now()
);

create index if not exists votes_transfer_index on votes using btree(transfer_index);

create table if not exists confirmations(
    id bigserial primary key,
    transfer_index bytea unique,
    rarimo_transaction bytea references transactions(hash),
    created_at timestamp without time zone not null default now()
);

create index if not exists confirmations_transfer_index on confirmations using hash(transfer_index);

create table if not exists approvals(
    id bigserial primary key,
    transfer_index bytea,
    rarimo_transaction bytea references transactions(hash),
    created_at timestamp without time zone not null default now()
);

create index if not exists approvals_transfer_index on approvals using hash(transfer_index);

create table if not exists rejections(
    id bigserial primary key,
    transfer_index bytea,
    rarimo_transaction bytea references transactions(hash),
    created_at timestamp without time zone not null default now()
);

create index if not exists rejections_transfer_index on rejections using hash(transfer_index);

create table if not exists collections(
    id bigserial primary key,
    index bytea not null unique,
    metadata jsonb not null default '{}'::jsonb,
    created_at timestamp without time zone not null default now(),
    updated_at timestamp without time zone not null default now()
);

CREATE TRIGGER set_updated_at BEFORE UPDATE ON collections FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

create table if not exists collection_chain_mappings(
    collection bigint not null references collections(id),
    network int not null,
    address bytea not null,
    token_type integer,
    wrapped boolean,
    decimals integer,
    created_at timestamp without time zone not null default now(),
    updated_at timestamp without time zone not null default now(),
    primary key (collection, network)
);

create index if not exists collection_chain_mappings_collection on collection_chain_mappings using btree(collection);
create index if not exists collection_chain_mappings_network on collection_chain_mappings using btree(network);

CREATE TRIGGER set_updated_at BEFORE UPDATE ON collection_chain_mappings FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

create table if not exists items(
    id bigserial primary key,
    index bytea not null unique,
    collection bigint references collections(id),
    metadata jsonb not null default '{}'::jsonb,
    created_at timestamp without time zone not null default now(),
    updated_at timestamp without time zone not null default now()
);

create index if not exists item_collection on items using btree(collection);

CREATE TRIGGER set_updated_at BEFORE UPDATE ON items FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

create table if not exists item_chain_mappings(
    item bigint not null references items(id),
    network int not null,
    address bytea not null,
    token_id bytea,
    created_at timestamp without time zone not null default now(),
    updated_at timestamp without time zone not null default now(),
    primary key (item, network)
);

create index if not exists item_chain_mapping_item on item_chain_mappings using btree(item);
create index if not exists item_chain_mapping_network on item_chain_mappings using btree(network);
create index if not exists item_chain_mapping_address on item_chain_mappings using btree(address);

CREATE TRIGGER set_updated_at BEFORE UPDATE ON item_chain_mappings FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

create table if not exists withdrawals(
    hash bytea primary key,
    block_height bigint,
    tx_result jsonb,
    success bool,
    created_at timestamp without time zone not null default now()
);

-- +migrate Down
drop table if exists withdrawals;

drop trigger if exists set_updated_at on item_chain_mappings;
drop index if exists item_chain_mapping_item;
drop index if exists item_chain_mapping_network;
drop index if exists item_chain_mapping_address;
drop table if exists item_chain_mappings;

drop trigger if exists set_updated_at on items;
drop index if exists item_collection;
drop table if exists items;

drop trigger if exists set_updated_at on collection_chain_mappings;
drop index if exists collection_chain_mappings_collection;
drop index if exists collection_chain_mappings_network;
drop table if exists collection_chain_mappings;

drop trigger if exists set_updated_at on collections;
drop table if exists collections;

drop index if exists rejections_transfer_index;
drop table if exists rejections;

drop index if exists approvals_transfer_index;
drop table if exists approvals;

drop index if exists confirmations_transfer_index;
drop table if exists confirmations;

drop index if exists votes_transfer_index;
drop table if exists votes;

drop table if exists transfers;
drop table if exists transactions;

drop domain if exists int_256;
