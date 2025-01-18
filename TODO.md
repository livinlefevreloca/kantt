# TODO's

## Collector
* getting fk violation on fk_pod_owner in collector because a default value of zero is getting passed in as the pod_owner_id.
    - This is most likey due to a type of owner we have not handled (DaemonSet?)

## Database
* Check indexing for pods table based on the pods endpoint

