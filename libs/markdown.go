package libs

import (
	"bytes"
	"fmt"
	"html/template"
	"os"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

func (s *Server) ParseMarkdown(data *Data) error {

	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	bytesArray, err := os.ReadFile(s.cfg.AboutMePath)
	if err != nil {
		return err
	}
	var buff bytes.Buffer
	if err := md.Convert(bytesArray, &buff); err != nil {
		return fmt.Errorf("failed to format home: %w", err)
	}
	data.Home.Content = template.HTML(buff.String())

	return nil
}
