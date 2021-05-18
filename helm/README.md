# Helm chart for multi-john
Using this helm chart, you can run multi-john on a Kubernetes cluster. Based on `kompose` originally (some unnecessary remnants probably remain in manifests under `templates`).

## Usage
### tldr
```shell
helm install multi-john . \
    --namespace <namespace> \
    --set="multijohn.node.johnFile.input=$(echo a5b432ee0307be7fa23aa00461f54eee34ba9d45251b5504567d37a8da339dff | base64)" \
    --set="multijohn.node.johnFlags=--format=raw-sha256" \
    --set="multijohn.totalNodes=5"
kubectl port-forward -n <namespace> howdy-<id> 8080:8080
curl localhost:8080
```

### How it works
1. Configure `values.yaml`
    * `.multijohn.node.johnFile.input` should contain the base64 contents of your file to pass to john. 
    * Make sure to pass the appropriate `.multijohn.node.johnFlags` to tell john what kind of hash it need to process.
    * You will probably want to set the `.totalNodes`. The chart will spawn one pod per node specified. 
2. Install the chart; 
```shell
kubectl create namespace <namespace>
helm install <name> . -n <namespace> --values values.yaml
# Check the Makefile for this too. `make install`
```
3. Port-forward `howdy` and start polling to get results; 
```shell
kubectl port-forward -n <namespace> howdy-<id> 8080:8080
watch "curl -s localhost:8080 | jq"
```
4. Once you're satisfied with the results (don't forget to save them), you can uninstall the chart; 
```shell
helm uninstall <name> -n <namespace>
# Check Makefile for this too. `make uninstall`
```