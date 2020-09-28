# CUTS - Cluster upgrade tests

## Workflow

- clone catapult
  ```
  $> git clone https://github.com/SUSE/catapult && cd catapult
  ```
- set up your k8s, if necesarry specify a `BACKEND`
  ```
  $> make k8s
  ```
  or
  ```
  $> BACKEND=gke make k8s
  ```
- export your catapult path
  ```
  $> export CATAPULT_PATH=/absolut/path/to/your/catapult/repo
  ```
- got to this repos path
- start the test suite
  ```
  $> ginkgo -r .
  ```
