﻿-- =============================================================================
-- Diagram Name: tradebot
-- Created on: 29.07.2025 14:20:39
-- Diagram Version: 
-- =============================================================================

CREATE TYPE marketplaces AS ENUM('WB', 
	'OZON', 
	'YANDEX');

CREATE TYPE types AS ENUM('fbo', 
	'fbs', 
	'all');


CREATE TABLE "stocks" (
	"stockId" SERIAL NOT NULL,
	"article" varchar(64) NOT NULL,
	"updatedAt" timestamp with time zone NOT NULL DEFAULT now(),
	"countFbo" int4,
	"countFbs" int4,
	"cabinetId" int4 NOT NULL,
	PRIMARY KEY("stockId")
);

CREATE TABLE "cabinets" (
	"cabinetsId" SERIAL NOT NULL,
	"name" varchar(64) NOT NULL,
	"clientId" varchar(64),
	"key" varchar(1024) NOT NULL,
	"marketplace" marketplaces NOT NULL,
	"type" types NOT NULL,
	"sheetLink" varchar(1024),
	"statusId" int4 NOT NULL,
	PRIMARY KEY("cabinetsId")
);

CREATE TABLE "orders" (
	"orderId" SERIAL NOT NULL,
	"postingNumber" varchar(32) NOT NULL,
	"article" varchar(128) NOT NULL,
	"count" int4 NOT NULL,
	"cabinetId" int4 NOT NULL,
	"createdAt" timestamp with time zone NOT NULL DEFAULT now(),
	"statusId" int4 NOT NULL,
	PRIMARY KEY("orderId")
);

CREATE TABLE "vfsHashes" (
	"hash" varchar(40) NOT NULL,
	"namespace" varchar(32) NOT NULL,
	"extension" varchar(4) NOT NULL,
	"fileSize" int4 NOT NULL DEFAULT 0,
	"width" int4 NOT NULL DEFAULT 0,
	"height" int4 NOT NULL DEFAULT 0,
	"blurhash" text,
	"error" text,
	"createdAt" timestamp with time zone NOT NULL DEFAULT now(),
	"indexedAt" timestamp with time zone,
	PRIMARY KEY("hash","namespace")
);

CREATE INDEX "IX_vfsHashes_indexedAt" ON "vfsHashes" USING BTREE (
	"indexedAt"
);


CREATE TABLE "vfsFolders" (
	"folderId" SERIAL NOT NULL,
	"parentFolderId" int4,
	"title" varchar(255) NOT NULL,
	"isFavorite" bool DEFAULT false,
	"createdAt" timestamp NOT NULL DEFAULT now(),
	"statusId" int4 NOT NULL,
	CONSTRAINT "vfsFolders_pkey" PRIMARY KEY("folderId")
);

CREATE INDEX "IX_FK_vfsFolders_folderId_vfsFolders" ON "vfsFolders" USING BTREE (
	"parentFolderId"
);


CREATE INDEX "IX_FK_vfsFolders_statusId_vfsFolders" ON "vfsFolders" USING BTREE (
	"statusId"
);


CREATE TABLE "vfsFiles" (
	"fileId" SERIAL NOT NULL,
	"folderId" int4 NOT NULL,
	"title" varchar(255) NOT NULL,
	"path" varchar(255) NOT NULL,
	"params" text,
	"isFavorite" bool DEFAULT false,
	"mimeType" varchar(255) NOT NULL,
	"fileSize" int4 DEFAULT 0,
	"fileExists" bool NOT NULL DEFAULT true,
	"createdAt" timestamp NOT NULL DEFAULT now(),
	"statusId" int4 NOT NULL,
	CONSTRAINT "vfsFiles_pkey" PRIMARY KEY("fileId")
);

CREATE INDEX "IX_FK_vfsFiles_folderId_vfsFiles" ON "vfsFiles" USING BTREE (
	"folderId"
);


CREATE INDEX "IX_FK_vfsFiles_statusId_vfsFiles" ON "vfsFiles" USING BTREE (
	"statusId"
);


CREATE TABLE "users" (
	"userId" SERIAL NOT NULL,
	"tgId" int8 NOT NULL,
	"login" varchar(64),
	"password" varchar(64),
	"authKey" varchar(32),
	"createdAt" timestamp with time zone NOT NULL DEFAULT now(),
	"lastActivityAt" timestamp with time zone,
	"isAdmin" bool NOT NULL DEFAULT False,
	"cabinetIds" int4[],
	"statusId" int4 NOT NULL,
	PRIMARY KEY("userId")
);

CREATE INDEX "IX_FK_users_statusId_users" ON "users" USING BTREE (
	"statusId"
);


CREATE UNIQUE INDEX "IX_tgId_unique" ON "users" (
	"tgId"
);


CREATE TABLE "statuses" (
	"statusId" SERIAL NOT NULL,
	"title" varchar(255) NOT NULL,
	"alias" varchar(64) NOT NULL,
	CONSTRAINT "statuses_pkey" PRIMARY KEY("statusId")
);


ALTER TABLE "stocks" ADD CONSTRAINT "Ref_stocks_to_cabinets" FOREIGN KEY ("cabinetId")
	REFERENCES "cabinets"("cabinetsId")
	MATCH SIMPLE
	ON DELETE NO ACTION
	ON UPDATE NO ACTION
	NOT DEFERRABLE;

ALTER TABLE "cabinets" ADD CONSTRAINT "Ref_cabinets_to_statuses" FOREIGN KEY ("statusId")
	REFERENCES "statuses"("statusId")
	MATCH SIMPLE
	ON DELETE NO ACTION
	ON UPDATE NO ACTION
	NOT DEFERRABLE;

ALTER TABLE "orders" ADD CONSTRAINT "Ref_orders_to_cabinets" FOREIGN KEY ("cabinetId")
	REFERENCES "cabinets"("cabinetsId")
	MATCH SIMPLE
	ON DELETE NO ACTION
	ON UPDATE NO ACTION
	NOT DEFERRABLE;

ALTER TABLE "orders" ADD CONSTRAINT "Ref_orders_to_statuses" FOREIGN KEY ("statusId")
	REFERENCES "statuses"("statusId")
	MATCH SIMPLE
	ON DELETE NO ACTION
	ON UPDATE NO ACTION
	NOT DEFERRABLE;

ALTER TABLE "vfsFolders" ADD CONSTRAINT "Ref_vfsFolders_to_statuses" FOREIGN KEY ("statusId")
	REFERENCES "statuses"("statusId")
	MATCH SIMPLE
	ON DELETE NO ACTION
	ON UPDATE NO ACTION
	NOT DEFERRABLE;

ALTER TABLE "vfsFolders" ADD CONSTRAINT "Ref_vfsFolders_to_vfsFolders" FOREIGN KEY ("parentFolderId")
	REFERENCES "vfsFolders"("folderId")
	MATCH SIMPLE
	ON DELETE NO ACTION
	ON UPDATE NO ACTION
	NOT DEFERRABLE;

ALTER TABLE "vfsFiles" ADD CONSTRAINT "Ref_vfsFiles_to_statuses" FOREIGN KEY ("statusId")
	REFERENCES "statuses"("statusId")
	MATCH SIMPLE
	ON DELETE NO ACTION
	ON UPDATE NO ACTION
	NOT DEFERRABLE;

ALTER TABLE "vfsFiles" ADD CONSTRAINT "Ref_vfsFiles_to_vfsFolders" FOREIGN KEY ("folderId")
	REFERENCES "vfsFolders"("folderId")
	MATCH SIMPLE
	ON DELETE NO ACTION
	ON UPDATE NO ACTION
	NOT DEFERRABLE;

ALTER TABLE "users" ADD CONSTRAINT "Ref_users_to_statuses" FOREIGN KEY ("statusId")
	REFERENCES "statuses"("statusId")
	MATCH SIMPLE
	ON DELETE NO ACTION
	ON UPDATE NO ACTION
	NOT DEFERRABLE;


