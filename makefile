build:
	@go build -o gotion .

run: build
	@./gotion