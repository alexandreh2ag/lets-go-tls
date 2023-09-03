package http

import (
	"time"

	"github.com/valyala/fasthttp"
)

type Client interface {
	DoTimeout(req *fasthttp.Request, resp *fasthttp.Response, timeout time.Duration) error
}
