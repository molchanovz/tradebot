-- =============================================================================
-- Diagram Name: tradebot
-- Created on: 15.04.2025 10:31:54
-- Diagram Version: 
-- =============================================================================

CREATE TYPE marketplaces AS ENUM('wildberries', 
	'ozon', 
	'yandex');


CREATE TABLE "stocks" (
	"stock_id" SERIAL NOT NULL,
	"article" varchar(64) NOT NULL,
	"updated_at" timestamp with time zone NOT NULL DEFAULT now(),
	"marketplace" marketplaces NOT NULL,
	"stocks_fbo" int4,
	"stocks_fbs" int4,
	PRIMARY KEY("stock_id")
);

CREATE TABLE "users" (
	"user_id" SERIAL NOT NULL,
	"tg_id" int8 NOT NULL,
	"status" int2 NOT NULL DEFAULT 1,
	PRIMARY KEY("user_id")
);

CREATE TABLE "cabinets" (
	"cabinets_id" int4 NOT NULL,
	"name" varchar(64) NOT NULL,
	"client_id" varchar(64),
	"key" varchar(256) NOT NULL,
	"user_id" int4 NOT NULL,
	PRIMARY KEY("cabinets_id")
);


ALTER TABLE "cabinets" ADD CONSTRAINT "Ref_cabinets_to_users" FOREIGN KEY ("user_id")
	REFERENCES "users"("user_id")
	MATCH SIMPLE
	ON DELETE NO ACTION
	ON UPDATE NO ACTION
	NOT DEFERRABLE;


