# Helm chart for multi-john
Using this helm chart, you can run multi-john on a Kubernetes cluster. Based on `kompose` originally, some unnecessary remnants probably remain in manifests under `templates`.

## Usage
1. Configure `values.yaml`, at minimum `.multijohn.node.johnFile.input` to contain the base64 contents of your file. You will probably want to set the `totalNodes` too.
2. Install the chart; 
```
kubectl create namespace <namespace>
helm install <name> . -n <namespace> --values values.yaml
# Check the Makefile for this too. `make install`
```
3. Port-forward `howdy` and start polling to get results; 
```
kubectl port-forward -n <namespace> howdy-<id> 8080:8080
watch "curl -s localhost:8080 | jq"
```
4. Once you're satisfied with the results (don't forget to save them), you can uninstall the chart; 
```
helm uninstall <name> -n <namespace>
# Check Makefile for this too. `make uninstall`
```