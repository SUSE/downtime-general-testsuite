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
