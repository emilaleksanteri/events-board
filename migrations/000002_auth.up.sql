create extension if not exists citext;


create table if not exists users(
    id bigserial primary key,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    email citext unique not null,
    name text,
    username text unique not null
);

create table if not exists providers(
    id bigserial primary key,
    provider text not null,
    user_id bigint not null references users(id) on delete cascade,
    access_token text not null,
    refresh_token text,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    expires_at timestamp(0) with time zone not null,
    id_token text not null,
    access_token_secret text not null
);

create table if not exists sessions(
    id bigserial primary key,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    user_id bigint not null references users(id),
    token text not null,
    expires_at timestamp(0) with time zone not null
);

create index if not exists users_username_idx on users using gin (to_tsvector('simple', username));
create index if not exists providers_user_id_idx on providers(user_id);
create index if not exists sessions_token_idx on sessions(token);
