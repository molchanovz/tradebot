
create table orders
(
    "orderId"       serial not null
        primary key,
    "postingNumber" varchar(32) not null,
    marketplace     marketplaces
);

