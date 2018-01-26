/*
 * Create dentry Table
 */

-- ----------------------------
--  Ensure uuid-ossp is enabled
-- ----------------------------
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ----------------------------
--  Table structure for dentry
-- ----------------------------
CREATE TABLE public.dentry(
    "id" uuid NOT NULL DEFAULT uuid_generate_v4(),
    "create_date" timestamptz NOT NULL DEFAULT timezone('UTC'::text, now()),
    "update_date" timestamptz NOT NULL DEFAULT timezone('UTC'::text, now()),
    "dtab" varchar(128) NOT NULL,
    "prefix" varchar(128) NOT NULL,
    "destination" varchar(128) NOT NULL,
    "priority" int NOT NULL,
    CONSTRAINT "pk_dentry" PRIMARY KEY ("id")
) WITH (OIDS = FALSE);

-- --------------------------
--  Indexes for dentry table
-- --------------------------
CREATE INDEX "ix_dentry_create_date_desc" ON "public"."dentry" USING btree(create_date DESC);
CREATE INDEX "ix_dentry_update_date_desc" ON "public"."dentry" USING btree(update_date DESC);
