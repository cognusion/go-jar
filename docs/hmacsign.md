# HMAC-Signed URLs

Sharing tamperproof URLs is an important feature of many applications. JAR can use "HMAC256" to create (and verify) signatures of the entire URI structure, which allows for easier caching of documents, immutability of query parameters, etc. JAR also has baked-in the concept of an expiration, so once created and signed, a signature can eventually become invalid due to age. Once verified, the signature is stripped from the request path before the request continues down the handler path.

* Normal: `/emery/food/pizzapie.jpg?width=200&quality=50`
* Signed: `/emery/food/pizzapie.jpg/76fbf049b3859331c83a51ca570f9b?width=200&quality=50`
* Signed w/ Expiration: `/emery/food/pizzapie.jpg/bf14920049162cf4686353c5384d88214f2?width=200&quality=50&expiration=1673381791311`

## Configuration

```yaml
-
    Path: /emery/
    hmacsigned: true
    cache: group1
    Options:
      hmac.key: abcdefghijk123
      hmac.expiration: 3h
    Pool: emery
```

### hmacsigned: [true|false]

**Default: false**
If set, the request **path** must have a valid HMAC signature appended to it. the **Path** *must* also have at least **hmac.key** set in the **Options** (example below), which must be the key used to create the aforementioned HMAC signture. See the **HMACSigner** **Finisher** documentation below for more information.

#### hmac.key: [string]

**Default: empty**
**REQUIRED**
Shared key to use between signer and verifier paths.

#### hmac.salt: [string]

**Default: empty**
Optional static salt to use, shared between signer and verifier paths.

#### hmac.expiration: [duration]

**Default: empty (off)**
If non-empty, must be a valid duration string. For signing, will automatically append a query parameter (before signature is computed) containing the offset expiration stamp. For verification, will be computed against the current system time if the signature is otherwise valid.

#### hmac.expirationfield: [string]

**Default: "expiration"**
Allows control over the name of the query parameter name used to store/retrieve the expiration stamp.
