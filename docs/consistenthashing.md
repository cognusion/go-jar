# Consistent-hashing Pools

## Various Means of Load-Balancing

### Round-robin, weighted/fair queuing, etc load-balancing

Treating every request atomically (or in classification groups) and balancing quazi-blindly across available instances is one of the purest means of ensuring availablility for idempotent requests. When combined with buffering this can provide completely seamless responses even in the event of significant-but-non-total outages, and mitigate many "thundering herd" cascading failover scenarios.

There are many use-cases, however, where a user/session/IP/browser/whatever is better served by being "pinned" to a serving instance. This could be because of heavy session-management, optimized file storage, locality, or other valid (and invalid :) ) reasons.

### Cookie-based Session Assignment

The `StickyCookie`-based load-balancing works exceptionally well for dynamic **Pool**s, where membership may scale up and down, as the cookie will keep a session consistently pinned as long as that member is available, and then algorithmically reassign the session to a different member if the pinned member becomes unavailable. This mechanism also well-handles situations where there are large concentrations of sessions behind a NAT or other IP masquerade that make unauthenticated sharding complicated.

Unfortunately, cookies are imperfect media for *reliably* distributing and assuring the pinning of arbitrary sessions, as corporate or user policy may curtail the use of cookies, and API-based access is complicated by needing to capture and regurgitate the cookie.

### Consistent-hashing Requests

Consistent hashing allows for the allocation of requests using various request-provided mechanisms, without having to rely on a dedicated cookie in order to do so, and handles both scaling up and down gracefully without many "thundering herd" scenarios. Whatever item is chosen as the key is hashed, the keys are distributed among partitions and partitions are distributed among members. In a stable situation, the same key will be in the same partition will be assigned to the same member. In a scaling or degrading situation, partitions will be reassigned, and thus the keys they held will be shifted.

## JAR Consistent-hashing Pools

The key item may be any request header value, cookie value, the requestor's IP address, the hostname they are trying to connect to, or the full URL they are requesting (or anything else you can think of that may have sufficient cardinality for your application). The value of that item is hashed and assigned to a partition, which has been assigned to a **pool member**. In the event the composition of the pool changes, that partition may be reassigned and thus subsequent requests with the same key will be transparently reassigned as well.

### Making a Great Hashkey

JAR allows for per-pool consistent-hashing configuration because different pools may have access to different key material. For example, a "first touch" pool where requests are unauthenticated and could be from anywhere on the Internet is very different than an authenticated-only client application pool. They keys (haha) to picking a good source are:

* ensuring that every request going into that pool *has* that source (i.e. don't pick an internal-use request header when few-if-any requests will have that header at all)
* ensuring that the cardinality of that source is sufficiently diverse to allow for proper sharding (i.e. if 60% of your requests come from a single NAT/VPN address, don't use the IP address as your source)
* if no single thing has both the necessary availability *and* cardinality, is there a combination that does? (unimplemented, but trivial (TODO))

### Convergence

In theory, a request will have the material for the key which will be hashed, its partition located, and the request forwarded on for resolution. Every time. In practice, however, there are timing issues where a partition may be reassigned while there are requests in-flight (and thus lost), or just as partitions are being reassigned (and thus delayed). Many of these may be mitigated by Buffering, so that the lost request can be replayed, but that requires one to understand the possible risks of duplicating (or triplicating, etc.) requests to any given pool. Partition reassignment is extremely fast- generally sub-microsecond after a failure is detected or members are deliberately added or removed.
