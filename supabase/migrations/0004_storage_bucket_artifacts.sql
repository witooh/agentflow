-- 0004_storage_bucket_artifacts.sql — Create artifacts storage bucket and policies

begin;

-- Create storage bucket `artifacts` (idempotent)
do $$
begin
  perform storage.create_bucket('artifacts', false);
exception when others then
  -- Ignore if bucket already exists or storage schema not available in local dev
  null;
end $$;

-- Ensure RLS is enabled on storage.objects
alter table if exists storage.objects enable row level security;

-- Helper: extract UUID prefix from object path `<projectId>/...`
-- We inline this via substring() to avoid extra function dependencies.

-- Drop existing policies if re-running
drop policy if exists "artifacts objects read by members" on storage.objects;
drop policy if exists "artifacts objects insert by members" on storage.objects;
drop policy if exists "artifacts objects update by owner" on storage.objects;
drop policy if exists "artifacts objects delete by owner" on storage.objects;

-- Read: project members can read objects under their project folder
create policy "artifacts objects read by members"
  on storage.objects
  for select
  to authenticated
  using (
    bucket_id = 'artifacts'
    and public.is_project_member(
      substring(name from '^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})/')::uuid
    )
  );

-- Insert: project members may upload into their project folder; owner must be self
create policy "artifacts objects insert by members"
  on storage.objects
  for insert
  to authenticated
  with check (
    bucket_id = 'artifacts'
    and owner = auth.uid()
    and public.is_project_member(
      substring(name from '^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})/')::uuid
    )
  );

-- Update: only the owner can modify, and must be a member of that project
create policy "artifacts objects update by owner"
  on storage.objects
  for update
  to authenticated
  using (
    bucket_id = 'artifacts'
    and owner = auth.uid()
    and public.is_project_member(
      substring(name from '^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})/')::uuid
    )
  )
  with check (
    bucket_id = 'artifacts'
    and owner = auth.uid()
    and public.is_project_member(
      substring(name from '^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})/')::uuid
    )
  );

-- Delete: only the owner can delete, and must be a member of that project
create policy "artifacts objects delete by owner"
  on storage.objects
  for delete
  to authenticated
  using (
    bucket_id = 'artifacts'
    and owner = auth.uid()
    and public.is_project_member(
      substring(name from '^([0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12})/')::uuid
    )
  );

commit;

