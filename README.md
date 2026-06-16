# ywan - a lazy task management CLI

yawn is a bare-bones taks management system that lives in your terminal. It allows you to organize tasks, jump between tasks, keep track of progress, and easily generate a hype repo. It is self contained, meaning it won't modify the repos where you do your work.

# tl;dr
```
$ yawn init # init yawn in first run
$ yawn new my-task # create a new task
$ yawn switch my-task # start working on your new task
[my-task] $ yawn open # manually update task's readme
[my-task] $ yawn done # done with task; merge to parent
```

# operations
## `init` - initialized `yawn`
## `new [--in <parent>] <name>` - create a new task
## `switch` - switch tasks
## `open` - open a tasks readme
## `prioritize <name>` - add a task to today's todo
## `list [--in <name>] [--today]` - list your tasks
## `update <status>` - update the current task's status
## `done` - shorthand for `update done` 
## `archive` - shorthand for `update archive`
## `carry-over` - carry over yesterday's incomplete tasks into today's todo

# design
`yawn init` creates a git repo in `~` with the base branch named `done`
`yawn new [--in <parent>] <name>` creates a branch off `<parent>` (or `done`) titled `<name>`. In this branch it creares a directory with the same name, containing the following files:
```
<name>
├── README.md
└── config.yaml
```

`README.md` will initially contain a title. `config.yaml` will contain thefollowing:

```
directory:
status: pending
rc: 
cleanup: 
```

You will be prompted to fill in these items. `directory` specifies the home directory of the task. When you `yawn switch <name>` yawn will `git switch` the root repo to the task's branch, `cd` into the task's directory, then perform the run commands (`rc`). yawn will perform `cleanup` when switching out of a task. An example of a task in a python project `foo` might look like so
```
directory: /Users/you/git/foo/
rc:
  - source .venv/bin/activate
cleanup:
  - deactivate
```

`yawn open` will open the current task's readme, and add a timestamp title.

`yawn done` (or `yawn update done`) will merge (or open a PR\*) the current task into its parent.
