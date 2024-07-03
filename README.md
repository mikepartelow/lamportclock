# Lamport Clock

Trivial demo of Lamport Clocks.

## Usage:

```bash
% make
go test -race ./...
ok  	mp/lc	(cached)
go run -race main.go
AbsoluteId  SenderId  producer.Clock  event.Clock
1           0         11              1
2           2         2               1
3           1         2               12
Lamport: false
---
AbsoluteId  SenderId  producer.Clock  event.Clock
1           0         11              12
2           2         13              14
3           1         15              16
Lamport: true
---
```
