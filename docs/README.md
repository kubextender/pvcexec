## Getting started

[filename](manual-install.md ':include')
* List persistent volume claims: 
  ```bash
  # list PVCs to get the ones we're interested in
  kubectl get pvc
  ```
* Run it!
  ```bash
  kubectl pvcexec mc -p first.pvc.id -p second.pvc.id
  ```
* This will by default open [midnight commander](https://midnight-commander.org/) showing given PVCs in left and right panel
* Happy browsing! :blush:

Project authors: [Dragan Ljubojevic](https://github.com/ljufa) and [Dusan Odalovic](https://github.com/dodalovic)