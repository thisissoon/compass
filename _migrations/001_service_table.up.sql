/*
 * Create service Table
 */

-- ----------------------------
--  Ensure uuid-ossp is enabled
-- ----------------------------
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ----------------------------
--  Table structure for service
-- ----------------------------
CREATE TABLE public.service(
    "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
    "create_date" timestamptz NOT NULL DEFAULT timezone('UTC'::text, now()),
    "update_date" timestamptz NOT NULL DEFAULT timezone('UTC'::text, now()),
    "logical_name" varchar(128) NOT NULL,
    "namespace" varchar(128) NOT NULL,
    "description" text NOT NULL,
    CONSTRAINT "pk_service" PRIMARY KEY ("id"),
    CONSTRAINT "uq_service_logical_name" UNIQUE("logical_name")
) WITH (OIDS = FALSE);

-- --------------------------
--  Indexes for service table
-- --------------------------
CREATE INDEX "ix_service_create_date_desc" ON "public"."service" USING btree(create_date DESC);
CREATE INDEX "ix_service_update_date_desc" ON "public"."service" USING btree(update_date DESC);
