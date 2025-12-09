PWD = $(shell pwd)
PATH = $(PWD)/pkg

keypair:
	@openssl genpkey -algorithm RSA -out $(PATH)/private_key_$(ENV).pem -pkeyopt rsa_keygen_bits:2048
	openssl rsa -in $(PATH)/private_key_$(ENV).pem -out $(PATH)/public_key_$(ENV).pem -pubout
swagger:
	@swag init -d ./cmd/main,./

up:
	@docker-compose up -d

down:
	@docker-compose down
