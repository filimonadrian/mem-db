# mem-db
An in-memory database that persists on disk



## Structure of the project

1. DBService
    - service which keeps data into a map[string]int
    - it contains:
        - snapshotter - takes periodic snapshots
        - WAL - every registered word is appended in a wal file - for event traking
        - Recovery mechanisms from WAL(recovery from event log - not efficient, but safer)

2. WordService
    - service which receives the requests from client and process the words
    - it has an API server and routines for managing data
    - contains the endpoints for the clients

3. NodeService
    - It's responsible for managing the replication and partitioning(TODO) mechanisms
    - Node name is the actual host of the service
        - I used the kubernetes host for the pod
        - Ex. `"node1-service.default.svc.cluster.local"`
        - But if you want to run on local machine, you should set a different host to connect to
        - it makes sense to be a different host, We don't want to replicate db on the same host

    - Has a http server for connecting to other nodes
        - if node is master:
            - starts some routines for replicating requests to workers and sending heartbeat
            - when worker start, it receives from master a snapshot of its database
            - master broadcasts the workers names when a new worker is added in the network
            - forwards the requests to the workers for replication
            - handles the leader election if the master fails(works only in kubernetes)
                - the replication process in handled with a custom http service
                - leader-election algo is used only if the master dies

        
## TODO

1. Partitioning
    - this can be easily prepared, but hardcoded
    - All master nodes should be configured into every config file
    - The map of nodes will have map[master_host]<range_of_starting_of_words>
    - The partitioning will be too hardcoded
    - A better solution would be to implement a ring for consistent hashing

2. GRPC