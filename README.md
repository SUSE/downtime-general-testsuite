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
- deploy kubecf using catapult, we suggest using our `kubecf_helper.sh` mentioned below using:
  ```
  $> export KUBECF_CHART=/absolut/path/to/the/base/helm/chart_bundle.tgz
  $> ./helpers/kubecf_helper.sh deploy
  ```
  if you want to run with HA and eirini use:
  ```
  $> HA=true ENABLE_EIRINI=true ./helpers/kubecf_helper.sh deploy
  ```
  instead of the last command above.

  **Note:** If you do not deploy with HA, the downtime test will alwyays fail.
- export your catapult path
  ```
  $> export CATAPULT_DIR=/absolut/path/to/your/catapult/repo
  ```
- export the absolute path to the target helm chart tarball (the suite relies on the kubcf_helper mentioned below)
  ```
  $> export KUBECF_TARGET_CHART=/absolut/path/to/the/target/helm/chart_bundle.tgz
  ```
- got to this repos path
- start the test suite
  ```
  $> ginkgo -r .
  ```

# kubecf_helper.sh

There is a helper script in the helpers directory of this repo that can help you set up clusters

## Usage

```
Usage: ./helpers/kubecf_helper.sh <deploy|upgrade|clean>

 deploy   - install kubecf with given chart from KUBECF_CHART
 upgrade  - upgrade existing kubecf with given chart from KUBECF_CHART
 password - get the admin password
 clean    - uninstall kubecf
```

If you want the cluster to be deployed with eirini use
```
$> ENABLE_EIRINI=true ./helpers/kubecf_helpers.sh deploy
```

## Environment variables

- `KUBECF_CHART` - absolute path to the bundle tarball of kubecf
- `CATAPULT_DIR` - absolute path to the catapult repository on your machine used to set up k8s
