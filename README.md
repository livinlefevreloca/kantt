# kantt

Sometimes it's useful to answer the question "what was running in
my cluster at a specific point in time?" The purpose of this tool
is to build Gantt charts that can help answer that question.

There are 2 core services defined in this repository:
- `collector` which listens for events and stores them to a
  persistent store.
- `dashboard` which provides a web interface to view the
  events.
 
Both services are intended to run within a cluster

## Repository structure

`/cmd` - Contains a directory per main package for short lived commands
  like database migration.

`/pkg` - Packages intended for import by other packages.

`/svc` - Contains a directory per main package for long lived services
  like the collector and dashboard.
