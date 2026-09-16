.PHONY: build-Function
build-Function:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o $(ARTIFACTS_DIR)/bootstrap -tags lambda.norpc .

.PHONY: test
test:
	go test -v ./...

.PHONY: cicd
cicd:
	aws cloudformation deploy \
		--region ap-northeast-1 \
		--stack-name "qrcode-generator-cicd" \
		--template-file "cicd.yaml" \
		--capabilities CAPABILITY_NAMED_IAM
