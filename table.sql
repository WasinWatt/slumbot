create table users (
	user_id varchar not null,
    name varchar not null,
    penalty_num int not null default 0,
    is_admin boolean not null default false,
    primary key (user_id)
);

create table rooms (
    room_id varchar not null,
    owner_id varchar not null,
    members varchar[] not null default '{}',
    available boolean not null default true,
    primary key (room_id)
);

