create table wb_stocks
(
    article text,
    date    timestamp not null,
    stock   bigint
);

create unique index wb_stocks_article_uindex
    on wb_stocks (article);

create table users
(
    chatid bigint  not null
        constraint table_name_pk
            primary key,
    state  integer not null
);

