create table users (
    id bigserial,
    user_id varchar unique not null,
    name varchar unique not null,
    penalty_num int default 0,
    primary key (user_id)
);

create table rooms (
    id bigserial,
    room_id varchar unique not null,
    owner_id varchar not null,
    members varchar[] not null default '{}',
    primary key (room_id)
);