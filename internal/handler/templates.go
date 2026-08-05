package handler

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"sort"
)

// templateFuncs son funciones auxiliares disponibles en todos los templates.
var templateFuncs = template.FuncMap{
	"percent": func(v float64) string { return fmt.Sprintf("%.0f", v*100) },
}

// pageDef describe el set de archivos que renderiza una página o partial.
// files son rutas relativas a <root>/web/templates, en el MISMO orden que los
// handlers usaban con ParseFiles (base primero). root nombra el template raíz:
// para partials es el basename del archivo (semántica de ParseFiles de nivel
// superior, que permite Execute sobre la raíz); para páginas queda vacío y se
// ejecutan por nombre ("base").
type pageDef struct {
	files []string
	root  string
}

// pageDefs es la única fuente de verdad de qué templates existen y cómo se
// arman; los handlers solo referencian su clave. La carga itera las claves en
// orden alfabético para que el resultado sea determinista.
var pageDefs = map[string]pageDef{
	"dashboard":               {files: []string{"base.html", "dashboard.html"}},
	"empleados-list":          {files: []string{"base.html", "empleados/list.html"}},
	"empleados-detail":        {files: []string{"base.html", "empleados/detail.html"}},
	"empleados-form":          {files: []string{"empleados/_form.html"}, root: "_form.html"},
	"empleados-import-form":   {files: []string{"empleados/_import_form.html"}, root: "_import_form.html"},
	"empleados-import-result": {files: []string{"empleados/_import_result.html"}, root: "_import_result.html"},
	"empleados-perfil-form":   {files: []string{"empleados/_perfil_form.html"}, root: "_perfil_form.html"},
	"empleados-row":           {files: []string{"empleados/_row.html"}, root: "_row.html"},
	"estimulos-apply-form":    {files: []string{"estimulos/_apply_form.html"}, root: "_apply_form.html"},
	"estimulos-list":          {files: []string{"base.html", "estimulos/list.html"}},
	"estimulos-table":         {files: []string{"estimulos/_table.html"}, root: "_table.html"},
	"incentivos-detail":       {files: []string{"base.html", "incentivos/detail.html"}},
	"incentivos-form":         {files: []string{"incentivos/_form.html"}, root: "_form.html"},
	"incentivos-list":         {files: []string{"base.html", "incentivos/list.html"}},
	"login":                   {files: []string{"login.html"}, root: "login.html"},
	"nudges-card":             {files: []string{"nudges/_card.html"}, root: "_card.html"},
	"nudges-detail":           {files: []string{"base.html", "nudges/detail.html"}},
	"nudges-form":             {files: []string{"nudges/_form.html"}, root: "_form.html"},
	"nudges-list":             {files: []string{"base.html", "nudges/list.html", "nudges/_card.html"}},
	"recomendacion":           {files: []string{"_recomendacion.html"}, root: "_recomendacion.html"},
	"riesgos":                 {files: []string{"_riesgos.html"}, root: "_riesgos.html"},
	"stats":                   {files: []string{"_stats.html"}, root: "_stats.html"},
	"distribucion":            {files: []string{"_distribucion.html"}, root: "_distribucion.html"},
	"efectividad":             {files: []string{"_efectividad.html"}, root: "_efectividad.html"},
}

// TemplateSet contiene los templates parseados UNA VEZ en el arranque, con
// rutas absolutas derivadas de una raíz explícita: el render es independiente
// del directorio de trabajo y no re-parsea por request.
type TemplateSet struct {
	pages map[string]*template.Template
}

// ParseTemplates parsea todos los sets de páginas/partials desde la raíz dada
// (<root>/web/templates). Cualquier archivo faltante o error de parseo se
// propaga como error: el arranque debe fallar ANTES de servir tráfico.
func ParseTemplates(root string) (*TemplateSet, error) {
	dir := filepath.Join(root, "web", "templates")
	names := make([]string, 0, len(pageDefs))
	for name := range pageDefs {
		names = append(names, name)
	}
	sort.Strings(names)

	ts := &TemplateSet{pages: make(map[string]*template.Template, len(pageDefs))}
	for _, name := range names {
		def := pageDefs[name]
		paths := make([]string, 0, len(def.files))
		for _, f := range def.files {
			paths = append(paths, filepath.Join(dir, f))
		}
		t, err := template.New(def.root).Funcs(templateFuncs).ParseFiles(paths...)
		if err != nil {
			return nil, fmt.Errorf("parse templates %q (%v): %w", name, def.files, err)
		}
		ts.pages[name] = t
	}
	return ts, nil
}

// execute renderiza el set indicado. name vacío ejecuta la raíz (partials);
// cualquier otro valor ejecuta ese template dentro del set (páginas: "base").
func (ts *TemplateSet) execute(w io.Writer, page, name string, data any) error {
	t, ok := ts.pages[page]
	if !ok {
		return fmt.Errorf("page %q no está cargada en el set de templates", page)
	}
	if name == "" {
		return t.Execute(w, data)
	}
	return t.ExecuteTemplate(w, name, data)
}
