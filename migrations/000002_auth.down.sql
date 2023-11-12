drop table if exists providers;
drop table if exists sessions;
drop table if exists users;
drop extension if exists citext;
drop index if exists users_username_idx;
drop index if exists providers_user_id_idx;
drop index if exists sessions_token_idx;
