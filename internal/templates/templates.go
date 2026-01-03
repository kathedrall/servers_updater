package templates

import (
 "embed"
 "fmt"
 "html/template"
 "io"
)

var emailTemplateFS embed.FS

func RenderEmail(w io.Writer, templateName string, data interface{}) error {
 fullPath := fmt.Sprintf("email/%s.html", templateName)

 tmpl, err := template.ParseFS(emailTemplateFS, fullPath)
 if err != nil {
  return fmt.Errorf("Failed to load template %s: %v", templateName, err)
 }
 
 return tmpl.Execute(w, data)
}
