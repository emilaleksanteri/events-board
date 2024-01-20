create extension if not exists ltree;

create table if not exists posts (
    id bigserial primary key,
    body text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table if not exists comments (
    id bigserial primary key,
    post_id bigint not null references posts on delete cascade,
    path ltree not null,
    body text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    foreign key (post_id) references posts (id)
);

CREATE INDEX path_gist_idx ON comments USING GIST (path);
CREATE INDEX path_idx ON comments USING BTREE (path);

create extension if not exists citext;
create table if not exists users(
    id bigserial primary key,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    email citext unique not null,
    name text,
    username text not null,
    profile_picture text not null
);

alter table if exists posts 
    add column if not exists user_id bigint not null references users on delete cascade;
alter table if exists comments 
    add column if not exists user_id bigint not null references users on delete cascade;

create table if not exists friend_nodes (
    id bigserial primary key,
    userId bigint not null references users(id) on delete cascade
);

create table if not exists friend_edges (
    previous_node bigint references friend_nodes(id),
    next_node bigint references friend_nodes(id),
    primary key (previous_node, next_node)
);
