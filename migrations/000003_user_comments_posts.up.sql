alter table if exists posts 
    add column if not exists user_id bigint not null references users on delete cascade;
alter table if exists comments 
    add column if not exists user_id bigint not null references users on delete cascade;


create index if not exists posts_user_id on posts(user_id);
create index if not exists comments_user_id on comments(user_id);


