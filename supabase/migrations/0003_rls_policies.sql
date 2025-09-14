-- 0003_rls_policies.sql — RLS policies for core tables
-- Covers: projects, project_members, requirements, artifacts, tasks,
--         task_dependencies, agent_runs, messages, memories

begin;

-- Helper functions for membership checks
create or replace function public.is_project_owner(pid uuid)
returns boolean
language sql
stable
as $$
  select exists (
    select 1 from public.projects p
    where p.id = pid and p.owner_id = auth.uid()
  );
$$;

create or replace function public.is_project_member(pid uuid)
returns boolean
language sql
stable
as $$
  select public.is_project_owner(pid)
         or exists (
           select 1
           from public.project_members pm
           where pm.project_id = pid
             and pm.user_id = auth.uid()
         );
$$;

create or replace function public.is_project_admin(pid uuid)
returns boolean
language sql
stable
as $$
  select public.is_project_owner(pid)
         or exists (
           select 1
           from public.project_members pm
           where pm.project_id = pid
             and pm.user_id = auth.uid()
             and pm.role in ('owner','admin')
         );
$$;

-- Enable RLS on all core tables
alter table if exists public.projects enable row level security;
alter table if exists public.project_members enable row level security;
alter table if exists public.requirements enable row level security;
alter table if exists public.artifacts enable row level security;
alter table if exists public.tasks enable row level security;
alter table if exists public.task_dependencies enable row level security;
alter table if exists public.agent_runs enable row level security;
alter table if exists public.messages enable row level security;
alter table if exists public.memories enable row level security;

-- Projects
drop policy if exists "projects_select_members" on public.projects;
drop policy if exists "projects_insert_owner_only" on public.projects;
drop policy if exists "projects_update_owner_admin" on public.projects;
drop policy if exists "projects_delete_owner_only" on public.projects;

create policy "projects_select_members"
  on public.projects
  for select
  to authenticated
  using (public.is_project_member(id));

create policy "projects_insert_owner_only"
  on public.projects
  for insert
  to authenticated
  with check (owner_id = auth.uid());

create policy "projects_update_owner_admin"
  on public.projects
  for update
  to authenticated
  using (public.is_project_admin(id))
  with check (public.is_project_admin(id));

create policy "projects_delete_owner_only"
  on public.projects
  for delete
  to authenticated
  using (owner_id = auth.uid());

-- Project members
drop policy if exists "project_members_select_members" on public.project_members;
drop policy if exists "project_members_insert_admin" on public.project_members;
drop policy if exists "project_members_update_owner_only" on public.project_members;
drop policy if exists "project_members_delete_owner_only" on public.project_members;

create policy "project_members_select_members"
  on public.project_members
  for select
  to authenticated
  using (public.is_project_member(project_id));

create policy "project_members_insert_admin"
  on public.project_members
  for insert
  to authenticated
  with check (public.is_project_admin(project_id));

create policy "project_members_update_owner_only"
  on public.project_members
  for update
  to authenticated
  using (public.is_project_owner(project_id))
  with check (public.is_project_owner(project_id));

create policy "project_members_delete_owner_only"
  on public.project_members
  for delete
  to authenticated
  using (public.is_project_owner(project_id));

-- Requirements
drop policy if exists "requirements_select_members" on public.requirements;
drop policy if exists "requirements_insert_members" on public.requirements;
drop policy if exists "requirements_update_members" on public.requirements;
drop policy if exists "requirements_delete_members" on public.requirements;

create policy "requirements_select_members"
  on public.requirements
  for select
  to authenticated
  using (public.is_project_member(project_id));

create policy "requirements_insert_members"
  on public.requirements
  for insert
  to authenticated
  with check (
    public.is_project_member(project_id)
    and (created_by is null or created_by = auth.uid())
  );

create policy "requirements_update_members"
  on public.requirements
  for update
  to authenticated
  using (public.is_project_member(project_id))
  with check (public.is_project_member(project_id));

create policy "requirements_delete_members"
  on public.requirements
  for delete
  to authenticated
  using (public.is_project_member(project_id));

-- Artifacts
drop policy if exists "artifacts_select_members" on public.artifacts;
drop policy if exists "artifacts_insert_members" on public.artifacts;
drop policy if exists "artifacts_update_members" on public.artifacts;
drop policy if exists "artifacts_delete_members" on public.artifacts;

create policy "artifacts_select_members"
  on public.artifacts
  for select
  to authenticated
  using (public.is_project_member(project_id));

create policy "artifacts_insert_members"
  on public.artifacts
  for insert
  to authenticated
  with check (
    public.is_project_member(project_id)
    and (created_by is null or created_by = auth.uid())
  );

create policy "artifacts_update_members"
  on public.artifacts
  for update
  to authenticated
  using (public.is_project_member(project_id))
  with check (public.is_project_member(project_id));

create policy "artifacts_delete_members"
  on public.artifacts
  for delete
  to authenticated
  using (public.is_project_member(project_id));

-- Tasks
drop policy if exists "tasks_select_members" on public.tasks;
drop policy if exists "tasks_insert_members" on public.tasks;
drop policy if exists "tasks_update_members" on public.tasks;
drop policy if exists "tasks_delete_members" on public.tasks;

create policy "tasks_select_members"
  on public.tasks
  for select
  to authenticated
  using (public.is_project_member(project_id));

create policy "tasks_insert_members"
  on public.tasks
  for insert
  to authenticated
  with check (
    public.is_project_member(project_id)
    and (created_by is null or created_by = auth.uid())
  );

create policy "tasks_update_members"
  on public.tasks
  for update
  to authenticated
  using (public.is_project_member(project_id))
  with check (public.is_project_member(project_id));

create policy "tasks_delete_members"
  on public.tasks
  for delete
  to authenticated
  using (public.is_project_member(project_id));

-- Task dependencies
drop policy if exists "task_deps_select_members" on public.task_dependencies;
drop policy if exists "task_deps_insert_members" on public.task_dependencies;
drop policy if exists "task_deps_delete_members" on public.task_dependencies;

create policy "task_deps_select_members"
  on public.task_dependencies
  for select
  to authenticated
  using (public.is_project_member(project_id));

create policy "task_deps_insert_members"
  on public.task_dependencies
  for insert
  to authenticated
  with check (public.is_project_member(project_id));

create policy "task_deps_delete_members"
  on public.task_dependencies
  for delete
  to authenticated
  using (public.is_project_member(project_id));

-- Agent runs
drop policy if exists "agent_runs_select_members" on public.agent_runs;
drop policy if exists "agent_runs_insert_members" on public.agent_runs;
drop policy if exists "agent_runs_update_members" on public.agent_runs;
drop policy if exists "agent_runs_delete_members" on public.agent_runs;

create policy "agent_runs_select_members"
  on public.agent_runs
  for select
  to authenticated
  using (public.is_project_member(project_id));

create policy "agent_runs_insert_members"
  on public.agent_runs
  for insert
  to authenticated
  with check (public.is_project_member(project_id));

create policy "agent_runs_update_members"
  on public.agent_runs
  for update
  to authenticated
  using (public.is_project_member(project_id))
  with check (public.is_project_member(project_id));

create policy "agent_runs_delete_members"
  on public.agent_runs
  for delete
  to authenticated
  using (public.is_project_member(project_id));

-- Messages
drop policy if exists "messages_select_members" on public.messages;
drop policy if exists "messages_insert_members" on public.messages;
drop policy if exists "messages_update_members" on public.messages;
drop policy if exists "messages_delete_members" on public.messages;

create policy "messages_select_members"
  on public.messages
  for select
  to authenticated
  using (public.is_project_member(project_id));

create policy "messages_insert_members"
  on public.messages
  for insert
  to authenticated
  with check (public.is_project_member(project_id));

create policy "messages_update_members"
  on public.messages
  for update
  to authenticated
  using (public.is_project_member(project_id))
  with check (public.is_project_member(project_id));

create policy "messages_delete_members"
  on public.messages
  for delete
  to authenticated
  using (public.is_project_member(project_id));

-- Memories
drop policy if exists "memories_select_members" on public.memories;
drop policy if exists "memories_insert_members" on public.memories;
drop policy if exists "memories_update_members" on public.memories;
drop policy if exists "memories_delete_members" on public.memories;

create policy "memories_select_members"
  on public.memories
  for select
  to authenticated
  using (public.is_project_member(project_id));

create policy "memories_insert_members"
  on public.memories
  for insert
  to authenticated
  with check (public.is_project_member(project_id));

create policy "memories_update_members"
  on public.memories
  for update
  to authenticated
  using (public.is_project_member(project_id))
  with check (public.is_project_member(project_id));

create policy "memories_delete_members"
  on public.memories
  for delete
  to authenticated
  using (public.is_project_member(project_id));

commit;

