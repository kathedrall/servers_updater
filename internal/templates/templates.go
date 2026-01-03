package templates

import (
 "embed"
 "fmt"
 "html/template"
 "io"
 "io/fs"
)

//go:embed email/*.html
var emailTemplateFS embed.FS

func RenderEmail(w io.Writer, templateName string, data interface{}) error {
 fsys, err := fs.Sub(emailTemplateFS, "email")
 if err != nil {
   return fmt.Errorf("Error accessing subdirectory.: %w", err)
 }

 fileName := fmt.Sprintf("%s.html", templateName)
 tmpl, err := template.ParseFS(fsys, fileName)
 if err != nil {
  return fmt.Errorf("Failed to load template %s: %v", templateName, err)
 }
 
 return tmpl.Execute(w, data)
}
