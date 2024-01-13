drop index if exists post_likes_post_id_user_id_idx;

drop index if exists comment_likes_comment_id_user_id_idx;

drop table if exists post_likes;

drop table if exists comment_likes;

alter table if exists posts 
    drop column if exists total_likes;

alter table if exists comments
    drop column if exists total_likes;

