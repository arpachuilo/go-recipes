package main

// import (
// 	"fmt"
// 	"html/template"
// 	"net/http"
// 	"os/exec"
// 	"sync"
// 	"time"
//
// 	"github.com/arpachuilo/go-registrable"
// 	"github.com/labstack/echo/v4"
// )
//
// type RecipeCluster struct {
// 	db       string
// 	output   string
// 	clusters int
//
// 	mux   sync.Mutex
// 	after time.Duration
// 	timer *time.Timer
// }
//
// func NewRecipeCluster(db, output string, clusters int) *RecipeCluster {
// 	return &RecipeCluster{
// 		db:       db,
// 		clusters: clusters,
// 		output:   output,
//
// 		after: time.Duration(time.Second * 30),
// 	}
// }
//
// func (self *RecipeCluster) Run() {
// 	self.mux.Lock()
// 	defer self.mux.Unlock()
//
// 	if self.timer != nil {
// 		self.timer.Stop()
// 	}
//
// 	self.timer = time.AfterFunc(self.after, func() {
// 		self.cluster()
// 	})
// }
//
// func (self *RecipeCluster) cluster() error {
// 	// scrape URL
// 	out, err := exec.Command("python3", "cmds/cluster/cluster.py", self.db, fmt.Sprintf("%v", self.clusters), self.output).CombinedOutput()
// 	fmt.Println(string(out))
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//
// func (self App) ServeCluster() registrable.Registration {
// 	// read templates dynamically for debug
// 	tmplFullName := "recipe_clustering"
// 	fullTemplate := template.Must(template.New("base").Funcs(templateFns).ParseFiles(
// 		"templates/base.gohtml",
// 		"templates/nav.gohtml",
// 		"templates/recipe_search.gohtml",
// 		"templates/recipe_cluster.gohtml",
// 	))
//
// 	self.TemplateRenderer.Add(tmplFullName, fullTemplate)
// 	return EchoHandlerRegistration{
// 		Path:        "/cluster",
// 		Methods:     []Method{GET},
// 		RequireAuth: self.Auth.enabled,
// 		HandlerFunc: func(c echo.Context) error {
// 			return c.Render(http.StatusOK, tmplFullName, nil)
// 		},
// 	}
// }
