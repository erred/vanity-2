package serve

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"text/template"
	"time"

	"go.seankhliao.com/svcrunner/v2/basehttp"
	"go.seankhliao.com/svcrunner/v2/observability"
	"go.seankhliao.com/webstyle"
	"go.seankhliao.com/webstyle/webstatic"
)

var (
	//go:embed index.md
	indexRaw []byte

	//go:embed repo.md.tpl
	repoRaw string
	repoTpl = template.Must(template.New("").Parse(repoRaw))

	headRaw = `
    <meta
      name="go-import"
      content="go.seankhliao.com/{{ .Repo }} git https://{{ .Source }}/{{ .Repo }}">
    <meta
      name="go-source"
      content="{{ .Host }}/{{ .Repo }}
        https://{{ .Source }}/{{ .Repo }}
        https://{{ .Source }}/{{ .Repo }}/tree/master{/dir}
        https://{{ .Source }}/{{ .Repo }}/blob/master{/dir}/{file}#L{line}">
`
	headTpl = template.Must(template.New("").Parse(headRaw))
)

type Server struct {
	o *observability.O

	svr *basehttp.Server

	host   string
	source string

	ts     time.Time
	render webstyle.Renderer

	index []byte
}

func New(ctx context.Context, c *Cmd) *Server {
	svr := basehttp.New(ctx, &c.basehttp)
	s := &Server{
		o:   svr.O,
		svr: svr,

		host:   c.host,
		source: c.source,
		ts:     time.Now(),
		render: webstyle.NewRenderer(webstyle.TemplateCompact),
	}

	svr.Mux.Handle("/", s)
	webstatic.Register(svr.Mux)

	return s
}

func (s *Server) Run(ctx context.Context) error {
	var err error
	s.index, err = s.render.RenderBytes(indexRaw, webstyle.Data{})
	if err != nil {
		return fmt.Errorf("render index page: %w", err)
	}
	return s.svr.Run(ctx)
}

func (s *Server) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx, span := s.o.T.Start(r.Context(), "serve vanity")
	defer span.End()

	if r.Method != http.MethodGet {
		s.o.HTTPErr(ctx, "GET only", errors.New("method not allowed"), rw, http.StatusMethodNotAllowed)
		return
	}

	slogHTTPRequest := slog.Group("http_request",
		slog.String("method", r.Method),
		slog.String("url", r.URL.String()),
		slog.String("proto", r.Proto),
		slog.String("user_agent", r.UserAgent()),
		slog.String("remote_address", r.RemoteAddr),
		slog.String("referrer", r.Referer()),
		slog.String("x-forwarded-for", r.Header.Get("x-forwarded-for")),
		slog.String("forwarded", r.Header.Get("forwarded")),
	)

	p := strings.TrimPrefix(r.URL.Path, "/")
	if p == "" { // index
		http.ServeContent(rw, r, "index.html", s.ts, bytes.NewReader(s.index))
		s.o.L.LogAttrs(ctx, slog.LevelInfo, "served index page",
			slogHTTPRequest,
		)
		return
	}
	repo, _, _ := strings.Cut(p, "/")
	data := map[string]string{"Repo": repo, "Source": s.source, "Host": s.host}

	var buf1 bytes.Buffer
	err := repoTpl.Execute(&buf1, data)
	if err != nil {
		s.o.HTTPErr(ctx, "render repo", err, rw, http.StatusInternalServerError)
		return
	}
	var buf2 bytes.Buffer
	err = headTpl.Execute(&buf2, data)
	if err != nil {
		s.o.HTTPErr(ctx, "render head", err, rw, http.StatusInternalServerError)
		return
	}

	var buf3 bytes.Buffer
	err = s.render.Render(&buf3, &buf1, webstyle.Data{
		Head: buf2.String(),
	})
	if err != nil {
		s.o.HTTPErr(ctx, "render html", err, rw, http.StatusInternalServerError)
		return
	}
	_, err = io.Copy(rw, &buf3)
	if err != nil {
		s.o.HTTPErr(ctx, "write response", err, rw, http.StatusInternalServerError)
		return
	}

	s.o.L.LogAttrs(ctx, slog.LevelInfo, "served module page",
		slog.Group("vanity",
			slog.String("repo", repo),
		),
		slogHTTPRequest,
	)
}
