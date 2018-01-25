/*
 * Create service Table
 */

-- ----------------------------
--  Table structure for service
-- ----------------------------
CREATE TABLE public.service(
    "create_date" timestamptz NOT NULL DEFAULT timezone('UTC'::text, now()),
    "update_date" timestamptz NOT NULL DEFAULT timezone('UTC'::text, now()),
    "logical_name" varchar(128) NOT NULL,
    "namesapce" varchar(128) NOT NULL,
    "description" text NOT NULL,
    CONSTRAINT "pk_service" PRIMARY KEY ("logical_name", "namespace")
) WITH (OIDS = FALSE);

-- --------------------------
--  Indexes for service table
-- --------------------------
CREATE INDEX "ix_service_create_date_desc" ON "public"."service" USING btree(create_date DESC);
CREATE INDEX "ix_service_update_date_desc" ON "public"."service" USING btree(update_date DESC);
