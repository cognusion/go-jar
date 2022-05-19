# poolmanager

Poolmanager is a CLI resource to assist with communicating with JAR Pools via the JAR Pool Management API.
While one can certainly use raw "curls" (and the output includes them for reference), ``poolmanager`` takes the
guess/memory work out of the syntax, and can/will also find the JARDs for you.

```bash
$ poolmanager --help
Usage of ./poolmanager:
      --command string     Command to issue, one of: [add lose list]
      --debug              Enable vociferous output
      --pool string        Pool to act upon
      --scheme string      Protocol scheme to prefix (default "https")
      --srv                Use DNS SRV to look up targets (default true)
      --srvdomain string   Domain of DNS SRV to use
      --srvsuffix string   Suffix of DNS SRV to use (e.g. dev, prod, useast1c, etc.)
      --targets string     List of JARDs to manage (if not using SRV)
      --uri string         Base path to the manager (default "/admin/pool")
```

## SRV, FTW!

Poolmanager attempts to use DNS SRV lookups against relevant SRV records start with ``_jarpool`` and are suffixed e.g. ``dev`` or ``prod``, toggleable via the ``--srvsuffix`` option. This allows us to change hosts via DNS, without changing code or config of the tools.

In case you think you know what you're doing, and only want to target specific systems, you can do, e.g. ``--srv false --targets "localhost:8080,localhost:8081"``. You may also need to change ``--scheme`` to generate the proper URLs.

## Make it so

Poolmanager knows what commands are possible, and automatically displays them under the CLI help for the ``-command string`` option. SSL Certificate errors are also ignored, so don't worry about domain conflicts at whatnot.

### add

Adding a member to a pool is as simple as ``--pool <name> --command "add <protocol://baseURLtoMember>"``.

Double-quoting the entire command is recommended.

**NOTE:** Please keep in mind we prefer IP addresses to host names, so that we don't have to do DNS lookups which can fail, be slow, or otherwise cause issues that are all eliminated by using IP addresses.

### lose

Removing a member to a pool is as simple as ``--pool <name> --command "lose <protocol://baseURLtoMember>"``. Might want to copy-paste that somewhere to make re-adding later easier.

Double-quoting the entire command is recommended.

It's worth noting that while the member will be *immediately* removed from the pool, and no new connections will be allowed, existing connections will be properly shuttled until they are closed. In general, waiting 30 seconds after issuing a ``lose`` should be sufficient for most applications.

### list

If you're not sure what pools exist on a JARD instance, issuing ``--command list`` without a pool specified, will list all of the pools. To list the members of a pool, ``--pool <name> --command list`` is the option set you're looking for: The output syntax of which is ready for use in a ``lose`` or ``add`` command.
