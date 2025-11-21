# jar

Package jar is a readily-embeddable feature-rich proxy-focused AWS-aware
distributed-oriented resiliency-enabling URL-driven superlative-laced
***elastic application link***. At its core, JAR is "just" a load-balancing
proxy taking cues from HAProxy (resiliency, zero-drop restarts, performance)
and Apache HTTPD (virtualize everything) while leveraging over 20 years
of systems engineering experience to provide robust features with exceptional
stability.

JAR has been in production use since 2018 and handles millions of connections a day across heterogeneous application stacks.

Consumers will want to 'cd cmd/jard; go build; #enjoy'

* [Roadmap](docs/Roadmap.md)
* [Configuration Docs](docs/Configuration.md)
* [Features](docs/Features.md)
  * [Consistent-Hashing Pools](docs/consistenthashing.md)
  * [HMAC signature verification](docs/hmacsign.md)
  * [TUS-compatible Uploads](docs/tus.md)
  * [Distributed Content-caching](docs/cache.md)
  * [Middleware plugins](docs/Plugins.md)
* [GoDoc for embedding](docs/godoc.md)
* [FAQ](docs/FAQ.md)

## Code Stability Warning

Only tagged releases on *master* are considered stable. While *master* is always buildable, revisions between tagged releases are considered "development grade" and may not work as intended/described/expected.

### Long-term Support Branches
The *lts1* branch is currently the only long-term support branch.

*lts1* was considered feature-complete just before the *v1.2.0* release (circa December 2022) and has only received select fixes and security-related updates since. There are known code-quality issues, optimizations, enhancements, etc. that have been addressed elsewhere.

**Use the *lts1* branch if you need stability.**

### Pre-2.0.0

**Use the *lts1* branch if you need stability.**

Code released since the *v1.7.x* tagged release is especially ***under-tested***. The deprecation of AWS SDK v1 strongly encouraged an update to AWS SDK v2, which was non-trivial. That upgrade required updates of numerous other subsystems, not the least of which was TUS, which itself was a v1->v2 update. **BOTH AWS and TUS subsystems are *under-tested* until *v2.0.0* is released.** The tests for TUS also quazi-required using a different client for the testing, which is also less-than-known to us.

All tests are **passing**.

**Use the *lts1* branch if you need stability.**

#### What about v1.8.0?

There will be more *v1* releases for fixes, but the massive changes required to migrate to AWS SDK v2, TUS v2, etc. have made it untenable to safely release that into the v1 ecosystem. Commits since *v1.7* will trend towards *v2*.

**Use the *lts1* branch if you need stability.**

## Grinder
The load-generator tool included with JAR has been separated and expanded at [grinder](https://github.com/cognusion/grinder/).