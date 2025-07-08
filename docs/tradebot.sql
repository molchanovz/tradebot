-- =============================================================================
-- Diagram Name: tradebot
-- Created on: 08.07.2025 10:40:13
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

CREATE TABLE "users" (
	"userId" SERIAL NOT NULL,
	"tgId" int8 NOT NULL,
	"statusId" int2 NOT NULL DEFAULT 1,
	"isAdmin" bool NOT NULL DEFAULT false,
	"cabinetIds" int4[],
	PRIMARY KEY("userId")
);

CREATE UNIQUE INDEX "IX_tgId_unique" ON "users" (
	"tgId"
);


CREATE TABLE "cabinets" (
	"cabinetsId" SERIAL NOT NULL,
	"name" varchar(64) NOT NULL,
	"clientId" varchar(64),
	"key" varchar(1024) NOT NULL,
	"marketplace" marketplaces NOT NULL,
	"type" types NOT NULL,
	"sheetLink" varchar(1024),
	PRIMARY KEY("cabinetsId")
);

CREATE TABLE "orders" (
	"orderId" SERIAL NOT NULL,
	"postingNumber" varchar(32) NOT NULL,
	"article" varchar(128) NOT NULL,
	"count" int4 NOT NULL,
	"cabinetId" int4 NOT NULL,
	"createdAt" timestamp with time zone NOT NULL DEFAULT now(),
	PRIMARY KEY("orderId")
);


ALTER TABLE "stocks" ADD CONSTRAINT "Ref_stocks_to_cabinets" FOREIGN KEY ("cabinetId")
	REFERENCES "cabinets"("cabinetsId")
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


