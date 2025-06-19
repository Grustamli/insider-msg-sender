-- Create "message" table
CREATE TABLE "public"."message" ("id" serial NOT NULL, "recipient" character varying NOT NULL, "content" text NOT NULL, "message_id" character varying(100) NULL, "created_at" timestamp NULL DEFAULT CURRENT_TIMESTAMP, "sent_at" timestamp NULL, PRIMARY KEY ("id"));
