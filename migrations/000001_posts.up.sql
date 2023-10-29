create extension if not exists ltree;

create table if not exists posts (
    id bigserial primary key,
    title text not null,
    body text not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now()
);

create table if not exists comments (
    id bigserial primary key,
    post_id bigint not null references posts on delete cascade,
    path ltree not null,
    body text not null,
    created_at timestamp not null default now(),
    updated_at timestamp not null default now(),
    foreign key (post_id) references posts (id)
);

CREATE INDEX path_gist_idx ON comments USING GIST (path);
CREATE INDEX path_idx ON comments USING BTREE (path);


