.PHONY: demo
demo: install
	@clear
	@go run . --nats

.PHONY: preload-images
preload-images:
	# NATS images (derived dynamically from the chart)
	helm template my-nats nats/nats \
		| grep -E '^\s+image:' \
		| awk '{print $$2}' \
		| tr -d '"' \
		| sort -u \
		| xargs -I{} sh -c 'docker pull {} && kind load docker-image {}'
	# cert-manager images (derived dynamically from the charts)
	for image in $$(helm template cert-manager jetstack/cert-manager --namespace cert-manager --set crds.enabled=true \
		| grep -E '^\s+image:' | awk '{print $$2}' | tr -d '"' | sort -u) \
		$$(helm template cert-manager-csi-driver jetstack/cert-manager-csi-driver --namespace cert-manager \
		| grep -E '^\s+image:' | awk '{print $$2}' | tr -d '"' | sort -u); do \
		docker pull $$image && kind load docker-image $$image; \
	done
	# runtime-enforcer public images
	docker pull otel/opentelemetry-collector-contrib:0.147.0
	kind load docker-image otel/opentelemetry-collector-contrib:0.147.0
	# runtime-enforcer local-only images (already on this machine, not in any registry)
	kind load docker-image ghcr.io/flavio/runtime-enforcer/operator:v0.4.0-demo1
	kind load docker-image ghcr.io/flavio/runtime-enforcer/agent:v0.4.0-demo1

.PHONY: install
install: preload-images
	helm install cert-manager jetstack/cert-manager \
		--namespace cert-manager \
		--create-namespace \
		--set crds.enabled=true \
		--wait
	helm install cert-manager-csi-driver jetstack/cert-manager-csi-driver \
		--namespace cert-manager \
		--wait
	helm install runtime-enforcer runtime-enforcer/runtime-enforcer \
		--namespace runtime-enforcer \
		--create-namespace \
		--set operator.image.repository=ghcr.io/flavio/runtime-enforcer/operator \
		--set operator.image.tag=v0.4.0-demo1 \
		--set agent.image.repository=ghcr.io/flavio/runtime-enforcer/agent \
		--set agent.image.tag=v0.4.0-demo1 \
		--wait
