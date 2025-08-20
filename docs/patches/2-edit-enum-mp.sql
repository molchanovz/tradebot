alter table stocks
    drop column marketplace;

alter table stocks
    drop column "stocksFbo";

alter table stocks
    drop column "stocksFbs";

alter table stocks
    add "countFbo" integer;

alter table stocks
    add "countFbs" integer;

alter table stocks
    add "cabinetId" integer not null
        constraint "Ref_stocks_to_cabinets"
            references cabinets;



alter table cabinets
    alter column key type varchar(1024) using key::varchar(1024);

alter table cabinets
    drop constraint "Ref_cabinets_to_users";

alter table cabinets
    drop column "userId";

alter table cabinets
    add "sheetLink" varchar(1024);


alter table orders
    drop column marketplace;

alter table orders
    add article varchar(128) not null;

alter table orders
    add count integer not null;

alter table orders
    add "cabinetId" integer not null
        constraint "Ref_orders_to_cabinets"
            references cabinets;

alter table orders
    add "createdAt" timestamp with time zone default now() not null;


alter table users
    add "cabinetIds" integer[];

