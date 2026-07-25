module github.com/benapetr/gotty

go 1.22.12

require (
	github.com/NYTimes/gziphandler v1.1.1
	github.com/codegangsta/cli v1.22.17
	github.com/creack/pty v1.1.24
	github.com/fatih/structs v1.1.0
	github.com/gorilla/websocket v1.5.3
	github.com/pkg/errors v0.9.1
	github.com/yudai/hcl v0.0.0-20151013225006-5fa2393b3552
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.7 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
)

replace github.com/codegangsta/cli => github.com/urfave/cli v1.22.17
