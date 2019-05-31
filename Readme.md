# Simple Redis based rate limiter

### Limitations 
- All the nodes must use same rate.
- Redis keys used by this package must not used by other programs.
```
const RateLimitKey = "__RATELIMIT__"
const RequestSet = "__REQUESTSET__"
```
- Since I/O operation have their own latency, this package provides very close enough enough rate limiting but not exact.

