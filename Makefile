PWD = $(shell pwd)
BASE_PATH = $(PWD)

keypair:
	@openssl genpkey -algorithm RSA -out $(BASE_PATH)/private_key_$(ENV).pem -pkeyopt rsa_keygen_bits:2048 
	@openssl rsa -in $(BASE_PATH)/private_key_$(ENV).pem -out $(BASE_PATH)/public_key_$(ENV).pem -pubout

swagger:
	@swag init -g cmd/main/main.go -d . --output doc --parseDependency --parseInternal

up:
	@docker-compose up -d

down:
	@docker-compose down
