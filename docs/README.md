# Getting started

## Installing

```bash
bash <(curl -s https://raw.githubusercontent.com/kubextender/pvcexec/master/setup.sh)
```

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

## Architecture

`pvcexec` is implemented as a kubernetes plugin. It's written in Go language

## Feature requests

Feel free to vote for the next [features](https://doodle.com/poll/pnu5kbwnfmcphigt) to implement

## Project authors

Project kicked off by [Dragan Ljubojevic](https://github.com/ljufa) and [Dusan Odalovic](https://github.com/dodalovic)

Contributions are highly appreciated!