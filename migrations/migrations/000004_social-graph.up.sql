create table if not exists friend_nodes (
    id bigserial primary key,
    userId bigint not null
);

create table if not exists friend_edges (
    previous_node bigint references friend_nodes(id),
    next_node bigint references friend_nodes(id),
    primary key (previous_node, next_node)
);
