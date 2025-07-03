-- =============================================================================
-- Diagram Name: tradebot
-- Created on: 30.05.2025 15:35:01
-- Diagram Version: 
-- =============================================================================

CREATE TYPE marketplaces AS ENUM('wildberries', 
	'ozon', 
	'yandex');

CREATE TYPE types AS ENUM('fbo', 
	'fbs', 
	'all');


CREATE TABLE "stocks" (
	"stockId" SERIAL NOT NULL,
	"article" varchar(64) NOT NULL,
	"updatedAt" timestamp with time zone NOT NULL DEFAULT now(),
	"marketplace" marketplaces NOT NULL,
	"stocksFbo" int4,
	"stocksFbs" int4,
	PRIMARY KEY("stockId")
);

CREATE TABLE "users" (
	"userId" SERIAL NOT NULL,
	"tgId" int8 NOT NULL,
	"statusId" int2 NOT NULL DEFAULT 1,
	"isAdmin" bool NOT NULL DEFAULT false,
	PRIMARY KEY("userId")
);

CREATE UNIQUE INDEX "IX_tgId_unique" ON "users" (
	"tgId"
);


CREATE TABLE "cabinets" (
	"cabinetsId" SERIAL NOT NULL,
	"name" varchar(64) NOT NULL,
	"clientId" varchar(64),
	"key" varchar(256) NOT NULL,
	"marketplace" marketplaces NOT NULL,
	"type" types NOT NULL,
	"userId" int4 NOT NULL,
	PRIMARY KEY("cabinetsId")
);

CREATE TABLE "orders" (
	"orderId" SERIAL NOT NULL,
	"postingNumber" varchar(32) NOT NULL,
	"marketplace" marketplaces,
	PRIMARY KEY("orderId")
);


ALTER TABLE "cabinets" ADD CONSTRAINT "Ref_cabinets_to_users" FOREIGN KEY ("userId")
	REFERENCES "users"("userId")
	MATCH SIMPLE
	ON DELETE NO ACTION
	ON UPDATE NO ACTION
	NOT DEFERRABLE;


