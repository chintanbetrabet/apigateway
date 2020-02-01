
default: build publish deploy

build: 
	docker build -f gateway.Dockerfile -t api-gateway .

publish: 
	docker tag api-gateway gcr.io/hybrid-qubole/api-gateway:2
	docker push gcr.io/hybrid-qubole/api-gateway:2

deploy:
	#kubectl delete deploy gateway -n test 
	kubectl apply -f deploy.yaml
				