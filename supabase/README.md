# Supabase — Local Setup & pgvector

This folder contains database migrations and notes for running Supabase locally or applying schema to a hosted project.

What’s included
- Migration `0001_enable_pgvector.sql` enabling the `vector` extension (aka pgvector).

Prerequisites
- Supabase CLI installed: https://supabase.com/docs/guides/cli
- Docker running (for local dev via `supabase start`).

Environment keys (.env)
- Copy `.env.example` to `.env` at repo root.
- Fill the following from Supabase Dashboard → Project Settings → API:
  - `SUPABASE_URL`
  - `SUPABASE_ANON_KEY`
  - `SUPABASE_SERVICE_ROLE_KEY` (server-only; never expose to browser)

Local development (Docker)
1) Start stack
   - `supabase start`
2) Reset DB and apply all migrations in this folder
   - `supabase db reset`
3) (Optional) Verify extension
   - `supabase db query "select extname from pg_extension where extname = 'vector';"`

Hosted project (apply migrations)
- Link once: `supabase link --project-ref <PROJECT_REF>`
- Push schema: `supabase db push`

Notes
- The migration is idempotent (`create extension if not exists vector;`).
- Keep future schema changes as new numbered files under `supabase/migrations/`.
