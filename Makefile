all:
	go build -o git-theseus main.go

test-integration: git-theseus
	sh integration.sh