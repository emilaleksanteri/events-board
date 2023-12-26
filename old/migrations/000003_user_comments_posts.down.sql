alter table posts if exists drop column if exists user_id;
alter table comments if exists drop column if exists user_id;

drop index if exists posts_user_id;
drop index if exists comments_user_id;

