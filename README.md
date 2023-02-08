We want to build an application which will showcase leader election using zookeeper.

Flow:
- Application starts up
- Connects with Zookeeper
- Tries to become leader
- If it becomes leader
    - it starts producing "<leaderid> - <increasing number>" and sending to self [ token production module ]
    - it also sends these to other replicas [ send module ]
- If it doesn't become leader, it listens to messages from leader if any and appends to its log [ receive module ]

## membership

- we also need a znode in zk which will store all "active" replicas. This will be done by creating an ephemeral node in zk [ membership module ]

## leader unable to connect to zk

leader:
- the leader stops all leader activities once it knows that it's session with zk is over
- until it can connect to zk, it just sits there idle [ connected check ]
- when it can connect, it again puts its candidature (check for old candidature entry first), and becomes a replica [ checkleadership module ]
- it has to somehow sort through its own log, vs where the current leader is at [ logfix module ]

replicas:
- one of the replica should get notified by zk that it should check if its the new leader [ checkleadership module ]
- this replica becomes leader.
    - new leader now starts producing its own set of entries after the ones it had received from previous leader [ token production module ]
- other replicas start "following" the new leader
    - starts listening to new leader [ receive module ]
    - this might also require some log re-write [ logfix module ]

## Architecture

[App 1] -- [App 2] -- [App 3]
    |---|----|---|
[ZK 1] -- [ZK 2] -- [ZK 3]

## To run

docker-compose -f docker-compose.yml up --force-recreate --build -d

## To stop

docker-compose down -v
(zookeeper volumes need to be removed else we don't get a clean slate on new runs in zk)