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


