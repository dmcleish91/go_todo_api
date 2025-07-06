-- The queries.sql file is a reference for all the SQL statements used in the application.

CREATE TABLE IF NOT EXISTS public.projects (
    project_id uuid NOT NULL DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    project_name character varying NOT NULL,
    color character varying,
    is_inbox boolean DEFAULT false,
    parent_project_id uuid,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT projects_pkey PRIMARY KEY (project_id),
    CONSTRAINT projects_parent_project_id_fkey FOREIGN KEY (parent_project_id) REFERENCES public.projects(project_id),
    CONSTRAINT projects_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id)
);

CREATE TABLE IF NOT EXISTS public.labels (
    label_id uuid NOT NULL DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    name character varying NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT labels_pkey PRIMARY KEY (label_id),
    CONSTRAINT labels_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id),
    CONSTRAINT labels_user_name_unique UNIQUE (user_id, name)
);

CREATE TABLE IF NOT EXISTS public.tasks (
    task_id uuid NOT NULL DEFAULT gen_random_uuid(),
    project_id uuid,
    user_id uuid NOT NULL,
    content text NOT NULL,
    description text,
    due_date date,
    due_datetime time without time zone,
    priority smallint,
    is_completed boolean DEFAULT false,
    completed_at timestamp with time zone,
    parent_task_id uuid,
    order integer NOT NULL DEFAULT 0,
    labels jsonb DEFAULT '[]'::jsonb,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT tasks_pkey PRIMARY KEY (task_id),
    CONSTRAINT tasks_parent_task_id_fkey FOREIGN KEY (parent_task_id) REFERENCES public.tasks(task_id),
    CONSTRAINT tasks_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id),
    CONSTRAINT tasks_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(project_id)
);


-- The queries below are used in the projects model.

-- AddProject
INSERT INTO projects (
    user_id, project_name, color, is_inbox, parent_project_id
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING project_id, user_id, project_name, color, is_inbox, parent_project_id;

-- EditProjectByID
UPDATE projects SET
    project_name = $3,
    color = $4,
    is_inbox = $5,
    parent_project_id = $6
WHERE project_id = $1 AND user_id = $2
RETURNING project_id, user_id, project_name, color, is_inbox, parent_project_id;

-- GetProjectsByUserID
SELECT project_id, user_id, project_name, color, is_inbox, parent_project_id
FROM projects
WHERE user_id = $1
ORDER BY project_name ASC;

-- DeleteProjectByID
DELETE FROM projects WHERE project_id = $1 AND user_id = $2;


-- The queries below are used in the labels model.

-- AddLabel
INSERT INTO labels (
    user_id, name
) VALUES (
    $1, $2
) RETURNING label_id, user_id, name, created_at;

-- EditLabelByID
UPDATE labels SET
    name = $3
WHERE label_id = $1 AND user_id = $2
RETURNING label_id, user_id, name, created_at;

-- GetLabelsByUserID
SELECT label_id, user_id, name, created_at
FROM labels
WHERE user_id = $1
ORDER BY name ASC;

-- DeleteLabelByID
DELETE FROM labels WHERE label_id = $1 AND user_id = $2;


-- The queries below are used in the tasks model.

-- AddTask
INSERT INTO tasks (
    project_id, user_id, content, description, due_date, due_datetime, priority, parent_task_id, order, labels
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING task_id, project_id, user_id, content, description, due_date, due_datetime, priority, is_completed, completed_at, parent_task_id, order, labels;

-- EditTaskByID
UPDATE tasks SET
    project_id = $3,
    content = $4,
    description = $5,
    due_date = $6,
    due_datetime = $7,
    priority = $8,
    is_completed = $9,
    completed_at = $10,
    parent_task_id = $11,
    "order" = $12,
    labels = $13
WHERE task_id = $1 AND user_id = $2
RETURNING task_id, project_id, user_id, content, description, due_date, due_datetime, priority, is_completed, completed_at, parent_task_id, order, labels;

-- GetTasksByUserID
SELECT task_id, project_id, user_id, content, description, due_date, due_datetime, priority, is_completed, completed_at, parent_task_id, order, labels, created_at
FROM tasks
WHERE user_id = $1
ORDER BY created_at ASC;

-- ToggleTaskCompleted
UPDATE tasks SET is_completed = NOT is_completed, completed_at = CASE WHEN is_completed THEN NULL ELSE CURRENT_TIMESTAMP END WHERE task_id = $1 AND user_id = $2;

-- DeleteTaskByID
DELETE FROM tasks WHERE task_id = $1 AND user_id = $2;

-- AddLabelToTask
UPDATE tasks SET labels = labels || $2::jsonb WHERE task_id = $1 AND user_id = $3;

-- RemoveLabelFromTask
UPDATE tasks SET labels = labels - $2 WHERE task_id = $1 AND user_id = $3;
