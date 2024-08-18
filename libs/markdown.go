package libs

import (
	"bytes"
	"fmt"
	"html/template"
	"portfolio/models"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

func (s *Server) ParseMarkdown(data *models.Data) error {

	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var buff bytes.Buffer
	if err := md.Convert([]byte(data.Github.HomeRaw), &buff); err != nil {
		return fmt.Errorf("failed to format home: %w", err)
	}
	data.Home.Content = template.HTML(buff.String())

	return nil
}
