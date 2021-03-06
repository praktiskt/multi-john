# multi-john

Run [John the Ripper](https://github.com/openwall/john), but coordinated on many machines.

## Image

Sporadic releases on Docker hub; `praktiskt/multi-john:latest`.

## Helm chart

The easiest way to run it on many machines is to use the Helm chart and run it on Kubernetes. See the [helm directory](./helm). You could probably play around with the `docker-compose.yml` too, but I've primarily used it for development.

## How it works

`multi-john` runs a few services:

- `etcd` - used to coordinate different workers and log results.
- `worker` - Runs `john` and ships results to `etcd`.
- `howdy` - Small service to expose results. Queries `etcd` to expose the results.

![image](https://user-images.githubusercontent.com/21065360/128094542-4d2f7afb-c0e7-474c-a1f2-0e3d603a761d.png)

If no workers are started, no active session will be created. Once at least one worker has started, a session is created and workers are able to claim a slot if there are slots available (configured with `TOTAL_NODES`). If all workers terminate, the session will eventually be deleted (and results purged).

## Development

```
make standalone-etcd
make run # runs *.go
```
