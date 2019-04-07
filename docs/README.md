# Getting started

## Installing

!> MacOS and Linux are the only supported platforms at the moment

!> [`curl`](https://curl.haxx.se/) command line client is required for installation

```bash
bash <(curl -s https://raw.githubusercontent.com/kubextender/pvcexec/master/setup.sh)
```

> `setup.sh` script recognizes OS it's being executed on, downloads appropriate precompiled binary, and makes it 
> available to `kubectl` by being placed on `/usr/local/bin/kubectl-pvcexec`

## Running

* List persistent volume claims: 
  ```bash
  # list PVCs to get the ones we're interested in
  kubectl get pvc
  NAME       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
  testpvc1   Bound    pvc-1d62bbb1-48f6-11e9-90ea-448a5bd44db1   100Mi      RWO            nfs-client     16d
  testpvc2   Bound    pvc-27a0fffe-48f6-11e9-90ea-448a5bd44db1   100Mi      RWO            nfs-client     16d
  ```
* Run it!
  ```bash
  kubectl pvcexec mc -p testpvc1 -p testpvc2
  ```
  
* Getting help

  ```bash
  kubectl pvcexec -h
  ```

## How it works

`pvcexec` is implemented as a kubernetes plugin. It's written in Go language. How it works? It's rather simple: 

`pvcexec` expects subcommand to be passed: `mc` or `zsh`. Based on given subcommand, tool deploys pod into your kubernetes context
based on docker images, which we, opinionatedly, created [here](https://cloud.docker.com/u/kubextender/repository/list).

`pvcexec` also expects list of one or more pvc names to be mounted in given pod. They will be mounted by their name under `/mnt` directory. 

For the example given above, we will have two mounted directories: `/mnt/testpvc1` and `/mnt/testpvc2`. 

Once you're there, you can perform any file operations needed.
 
After you're done using `pvcexec` (by exiting `pod`'s shell) the tool will automatically purge the running pod. 

## Feature requests

Feel free to vote for the next [features](https://doodle.com/poll/pnu5kbwnfmcphigt) to implement

## Project authors

Project kicked off by [Dragan Ljubojevic](https://github.com/ljufa) and [Dusan Odalovic](https://github.com/dodalovic)

Contributions are highly appreciated!
