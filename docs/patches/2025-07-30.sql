alter table public.users
alter column "statusId" type integer using "statusId"::integer;

alter table public.users
    alter column "statusId" drop default;

-- column reordering is not supported public.users."statusId"

create table public."vfsHashes"
(
    hash        varchar(40)                            not null,
    namespace   varchar(32)                            not null,
    extension   varchar(4)                             not null,
    "fileSize"  integer                  default 0     not null,
    width       integer                  default 0     not null,
    height      integer                  default 0     not null,
    blurhash    text,
    error       text,
    "createdAt" timestamp with time zone default now() not null,
    "indexedAt" timestamp with time zone,
    primary key (hash, namespace)
);

create index "IX_vfsHashes_indexedAt"
    on public."vfsHashes" ("indexedAt");

alter table public.users
    add login varchar(64) not null;

-- column reordering is not supported public.users.login

alter table public.users
    add password varchar(64) not null;

-- column reordering is not supported public.users.password

alter table public.users
    add "authKey" varchar(32);

-- column reordering is not supported public.users."authKey"

alter table public.users
    add "createdAt" timestamp with time zone default now() not null;

-- column reordering is not supported public.users."createdAt"

alter table public.users
    add "lastActivityAt" timestamp with time zone;

-- column reordering is not supported public.users."lastActivityAt"

create index "IX_FK_users_statusId_users"
    on public.users ("statusId");

create table public.statuses
(
    "statusId" serial
        primary key,
    title      varchar(255) not null,
    alias      varchar(64)  not null
);

alter table public.cabinets
    add "statusId" integer not null
        constraint "Ref_cabinets_to_statuses"
            references public.statuses;

alter table public.orders
    add "statusId" integer not null
        constraint "Ref_orders_to_statuses"
            references public.statuses;

create table public."vfsFolders"
(
    "folderId"       serial
        primary key,
    "parentFolderId" integer
        constraint "Ref_vfsFolders_to_vfsFolders"
            references public."vfsFolders",
    title            varchar(255)            not null,
    "isFavorite"     boolean   default false,
    "createdAt"      timestamp default now() not null,
    "statusId"       integer                 not null
        constraint "Ref_vfsFolders_to_statuses"
            references public.statuses
);

create index "IX_FK_vfsFolders_folderId_vfsFolders"
    on public."vfsFolders" ("parentFolderId");

create index "IX_FK_vfsFolders_statusId_vfsFolders"
    on public."vfsFolders" ("statusId");

create table public."vfsFiles"
(
    "fileId"     serial
        primary key,
    "folderId"   integer                 not null
        constraint "Ref_vfsFiles_to_vfsFolders"
            references public."vfsFolders",
    title        varchar(255)            not null,
    path         varchar(255)            not null,
    params       text,
    "isFavorite" boolean   default false,
    "mimeType"   varchar(255)            not null,
    "fileSize"   integer   default 0,
    "fileExists" boolean   default true  not null,
    "createdAt"  timestamp default now() not null,
    "statusId"   integer                 not null
        constraint "Ref_vfsFiles_to_statuses"
            references public.statuses
);

create index "IX_FK_vfsFiles_folderId_vfsFiles"
    on public."vfsFiles" ("folderId");

create index "IX_FK_vfsFiles_statusId_vfsFiles"
    on public."vfsFiles" ("statusId");

alter table public.users
    add constraint "Ref_users_to_statuses"
        foreign key ("statusId") references public.statuses;

