create table if not exists post_likes (
    id bigserial primary key,
    post_id bigint not null references posts on delete cascade,
    user_id bigint not null references users on delete cascade,
    created_at timestamptz not null default now()
);

create unique index if not exists post_likes_post_id_user_id_idx on post_likes (post_id, user_id);

create table if not exists comment_likes (
    id bigserial primary key,
    comment_id bigint not null references comments on delete cascade,
    user_id bigint not null references users on delete cascade,
    created_at timestamptz not null default now()
);

create unique index if not exists comment_likes_comment_id_user_id_idx on comment_likes (comment_id, user_id);

alter table if exists posts 
    add column if not exists total_likes bigint not null default 0;

alter table if exists comments 
    add column if not exists total_likes bigint not null default 0;

