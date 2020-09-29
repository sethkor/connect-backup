mac:
	env GOOS=darwin GOARCH=amd64 go build -o connect-backup cmd/main.go

linux:
	env GOOS=linux GOARCH=amd64 go build -o connect-backup.linux
	gzip -fk9 lexbelt.linux

clean:
	rm connect-backup.linux connect-backup.linux.gz connect-backup

publish-test:
	goreleaser --snapshot --skip-publish --rm-dist

publish:
	goreleaser --rm-dist --skip-validate