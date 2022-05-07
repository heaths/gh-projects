package logger

import (
	"io"

	"github.com/heaths/go-console"
	"github.com/heaths/go-console/pkg/colorscheme"
)

func New(con *console.Console, style string) io.Writer {
	cs := con.ColorScheme().Clone(
		colorscheme.WithTTY(con.IsStderrTTY),
	)
	return &log{
		w:     con.Stderr(),
		color: cs.ColorFunc(style),
	}
}

type log struct {
	w     io.Writer
	color func(string) string
}

func (l *log) Write(buf []byte) (int, error) {
	s := l.color(string(buf))
	return l.w.Write([]byte(s))
}
