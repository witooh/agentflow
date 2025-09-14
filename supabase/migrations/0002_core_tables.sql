-- 0002_core_tables.sql — Core tables and indexes
-- Creates: projects, project_members, requirements, artifacts, tasks,
--          task_dependencies, agent_runs, messages, memories

-- Ensure required extensions
create extension if not exists pgcrypto;

begin;

-- Projects
create table if not exists public.projects (
  id uuid primary key default gen_random_uuid(),
  name text not null,
  description text,
  status text check (status in ('active','archived')) default 'active',
  owner_id uuid not null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

alter table public.projects
  add constraint projects_owner_fk
  foreign key (owner_id) references auth.users(id) on delete restrict;

create index if not exists idx_projects_owner on public.projects(owner_id);
create index if not exists idx_projects_status on public.projects(status);

-- Project members
create table if not exists public.project_members (
  project_id uuid not null references public.projects(id) on delete cascade,
  user_id uuid not null,
  role text check (role in ('owner','admin','member')) default 'member',
  created_at timestamptz not null default now(),
  invited_by uuid,
  primary key (project_id, user_id)
);

alter table public.project_members
  add constraint project_members_user_fk
  foreign key (user_id) references auth.users(id) on delete cascade;

alter table public.project_members
  add constraint project_members_invited_by_fk
  foreign key (invited_by) references auth.users(id) on delete set null;

create index if not exists idx_project_members_role on public.project_members(role);

-- Requirements
create table if not exists public.requirements (
  id uuid primary key default gen_random_uuid(),
  project_id uuid not null references public.projects(id) on delete cascade,
  kind text not null check (kind in ('business','functional','nonfunctional','constraint')),
  title text not null,
  content text,
  source text,
  status text check (status in ('draft','approved','rejected')) default 'draft',
  created_by uuid,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

alter table public.requirements
  add constraint requirements_created_by_fk
  foreign key (created_by) references auth.users(id) on delete set null;

create index if not exists idx_requirements_project on public.requirements(project_id);
create index if not exists idx_requirements_kind on public.requirements(kind);
create index if not exists idx_requirements_status on public.requirements(status);

-- Artifacts
create table if not exists public.artifacts (
  id uuid primary key default gen_random_uuid(),
  project_id uuid not null references public.projects(id) on delete cascade,
  name text,
  path text not null,
  bucket text not null default 'artifacts',
  mime_type text,
  size_bytes bigint,
  checksum text,
  version text,
  created_by uuid,
  created_at timestamptz not null default now()
);

alter table public.artifacts
  add constraint artifacts_created_by_fk
  foreign key (created_by) references auth.users(id) on delete set null;

create unique index if not exists ux_artifacts_project_path on public.artifacts(project_id, path);
create index if not exists idx_artifacts_project on public.artifacts(project_id);

-- Tasks
create table if not exists public.tasks (
  id uuid primary key default gen_random_uuid(),
  project_id uuid not null references public.projects(id) on delete cascade,
  title text not null,
  description text,
  role text not null check (role in ('pm','sa','architect','techlead','fe','be','qa','devops','writer')),
  estimate_hour int check (estimate_hour >= 0) default 0,
  status text not null check (status in ('todo','in_progress','review','done','blocked')) default 'todo',
  priority smallint default 0,
  assignee_id uuid,
  due_date date,
  started_at timestamptz,
  completed_at timestamptz,
  created_by uuid,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

alter table public.tasks
  add constraint tasks_assignee_fk
  foreign key (assignee_id) references auth.users(id) on delete set null;

alter table public.tasks
  add constraint tasks_created_by_fk
  foreign key (created_by) references auth.users(id) on delete set null;

create index if not exists idx_tasks_project on public.tasks(project_id);
create index if not exists idx_tasks_status on public.tasks(status);
create index if not exists idx_tasks_role on public.tasks(role);
create index if not exists idx_tasks_assignee on public.tasks(assignee_id);
create index if not exists idx_tasks_created_at on public.tasks(created_at desc);

-- Task dependencies
create table if not exists public.task_dependencies (
  id uuid primary key default gen_random_uuid(),
  project_id uuid not null references public.projects(id) on delete cascade,
  task_id uuid not null references public.tasks(id) on delete cascade,
  depends_on_task_id uuid not null references public.tasks(id) on delete cascade,
  created_at timestamptz not null default now(),
  constraint task_dep_self check (task_id <> depends_on_task_id)
);

create unique index if not exists ux_task_dependencies_pair on public.task_dependencies(task_id, depends_on_task_id);
create index if not exists idx_task_dependencies_project on public.task_dependencies(project_id);

-- Agent runs
create table if not exists public.agent_runs (
  id uuid primary key default gen_random_uuid(),
  project_id uuid not null references public.projects(id) on delete cascade,
  task_id uuid references public.tasks(id) on delete set null,
  agent text not null,
  status text not null check (status in ('started','blocked','output_ready','finished','failed')),
  input jsonb,
  output jsonb,
  error text,
  metadata jsonb,
  started_at timestamptz not null default now(),
  finished_at timestamptz
);

create index if not exists idx_agent_runs_project on public.agent_runs(project_id);
create index if not exists idx_agent_runs_task on public.agent_runs(task_id);
create index if not exists idx_agent_runs_status on public.agent_runs(status);
create index if not exists idx_agent_runs_started_at on public.agent_runs(started_at desc);

-- Messages
create table if not exists public.messages (
  id uuid primary key default gen_random_uuid(),
  project_id uuid not null references public.projects(id) on delete cascade,
  task_id uuid references public.tasks(id) on delete set null,
  run_id uuid references public.agent_runs(id) on delete set null,
  sender_type text not null check (sender_type in ('human','agent','system')),
  sender_id uuid,
  agent text,
  role text,
  content text,
  metadata jsonb,
  created_at timestamptz not null default now()
);

alter table public.messages
  add constraint messages_sender_fk
  foreign key (sender_id) references auth.users(id) on delete set null;

create index if not exists idx_messages_project on public.messages(project_id);
create index if not exists idx_messages_task on public.messages(task_id);
create index if not exists idx_messages_run on public.messages(run_id);
create index if not exists idx_messages_created_at on public.messages(created_at desc);

-- Memories (pgvector)
create table if not exists public.memories (
  id uuid primary key default gen_random_uuid(),
  project_id uuid not null references public.projects(id) on delete cascade,
  task_id uuid references public.tasks(id) on delete set null,
  scope text not null check (scope in ('project','task','global')) default 'project',
  source text,
  content text,
  embedding vector(1536) not null,
  metadata jsonb,
  created_at timestamptz not null default now()
);

create index if not exists idx_memories_project on public.memories(project_id);
create index if not exists idx_memories_task on public.memories(task_id);
create index if not exists idx_memories_embedding on public.memories using ivfflat (embedding vector_cosine_ops) with (lists = 100);

commit;

