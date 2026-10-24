create table items (
  id bigint not null primary key,
  name varchar(100) not null,
  note text,
  big bigint not null,
  price decimal(10, 2) not null,
  data blob not null,
  flag boolean not null,
  meta json,
  created datetime not null,
  day date not null
);
