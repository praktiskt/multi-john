install:
	kubectl create namespace multijohn &&\
	helm install multi-john . -n multijohn --values values.yaml

upgrade:
	helm upgrade multi-john . -n multijohn --values values.yaml

uninstall:
	helm uninstall multi-john -n multijohn &&\
	kubectl delete namespace multijohn --cascade