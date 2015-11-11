# Swarm Benchmarking Tool

```
NAME:
   swarm-bench - Swarm Benchmarking Tool

USAGE:
   swarm-bench [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --concurrency, -c "1"	Number of multiple requests to perform at a time. Default is one request at a time.
   --requests, -n "1"		Number of containers to start for the benchmarking session. The default is to just start a single container.
   --image, -i 			Image to use for benchmarking.
   --help, -h			show help
   --version, -v		print the version
```

## Example

```
$ swarm-bench -c 2 -n 10 -i nginx
[ 10%] 1/10 containers started
[ 20%] 2/10 containers started
[ 30%] 3/10 containers started
[ 40%] 4/10 containers started
[ 50%] 5/10 containers started
[ 60%] 6/10 containers started
[ 70%] 7/10 containers started
[ 80%] 8/10 containers started
[ 90%] 9/10 containers started
[100%] 10/10 containers started

Time taken for tests: 2.772266804s
Time per container: 228ms [50th] | 596ms [90th] | 675ms [99th]
```

