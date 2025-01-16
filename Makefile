include .env
export

.PHONY: run
run:
	go run cmd/retropp/main.go


report:
	go run cmd/retropp/main.go -notice
