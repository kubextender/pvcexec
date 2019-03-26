# Getting started

* Navigate to [project's Github latest release](https://github.com/kubextender/pvcexec/releases/latest) and download the binary for your platform
  * **Linux**: kubectl-pvcexec_x.y.z_linux_amd64
  * **MacOS**: kubectl-pvcexec_x.y.z_darwin_amd64
* Make the binary available on your `$PATH`, e.g `mv kubectl-pvcexec_x.y.z_linux_amd64 /usr/local/bin/`
* List persistent volume claims: 
  ```bash
  # list PVCs to get the ones we're interested in
  $ kubectl get pvc
  # Run it!
  $ kubectl pvcexec mc -p first.pvc.id -p second.pvc.id
  ```
* This will by default open [midnight commander](https://midnight-commander.org/) showing given PVCs in left and right panel
* Happy browsing! :blush:

Project authors: [Dragan Ljubojevic](https://github.com/ljufa) and [Dusan Odalovic](https://github.com/dodalovic)