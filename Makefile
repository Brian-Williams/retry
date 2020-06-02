# as of writing there is no 1.12-1.14 compliant way to version tool dependencies
# go: github.com/githubnemo/CompileDaemon upgrade => v1.2.1
# go: golang.org/x/tools => v0.0.0-20200530233709-52effbd89c51

DIR := $(dir $(lastword $(MAKEFILE_LIST)))


daemon:
	CompileDaemon -command="go generate $(DIR)" -exclude-dir=.git -exclude=errenum_string.go
