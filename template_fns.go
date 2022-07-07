package main

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"
	"unicode"

	"go-recipes/models"
)

var noAlpha = regexp.MustCompile(`[^0-9,\.]+`)
var templateFns map[string]any = template.FuncMap{
	"splitLines": func(u string) []string {
		return strings.Split(u, "\n")
	},
	"removeAlpha": func(u string) string {
		return noAlpha.ReplaceAllString(u, "")
	},
	"inc": func(i int) int {
		return i + 1
	},
	"hasTag": func(possible models.Tag, selected []string) bool {
		for _, s := range selected {
			if s == possible.Tag.String {
				return true
			}
		}

		return false
	},
	// used for edit
	"flattenTags": func(tags models.TagSlice) string {
		str := ""
		for i, tag := range tags {
			str += tag.Tag.String
			if i != len(tags)-1 {
				str += ","
			}
		}
		return str
	},
	// used for edit
	"flattenIngredients": func(ingredients models.IngredientSlice) string {
		str := ""
		for _, ingredient := range ingredients {
			str += fmt.Sprintf("%v\n", ingredient.Ingredient.String)
		}
		return str
	},
	// used for display
	"formatIngredients": func(ingredients models.IngredientSlice) []template.HTML {
		formatted := make([]template.HTML, 0)

		startNewList := true
		for index, ingredient := range ingredients {
			i := ingredient.Ingredient.String
			// detect if possibly a header (currently doing this sort strictly)
			if strings.HasPrefix(strings.ToLower(i), "for") && strings.HasSuffix(i, ":") {
				if !startNewList {
					// end list
					formatted = append(formatted, "</ul>")
				}

				formatted = append(formatted, template.HTML(fmt.Sprintf("<b>%v</b>", i)))
				startNewList = true
			} else {
				if startNewList {
					formatted = append(formatted, "<ul>")
					startNewList = false
				}

				// highlight amounts
				f := "<li onclick=\"strike(this)\">"
				h := false
				d := true
				for _, r := range i {
					if d && !unicode.IsLetter(r) {
						if !h {
							h = true
							f += "<b>"
						}

						f += string(r)
					} else if h {
						f += "</b>" + string(r)
						h = false
						d = false
					} else {
						d = false
						f += string(r)
					}
				}
				f += "</li>"

				formatted = append(formatted, template.HTML(f))

				if index == len(ingredients)-1 {
					formatted = append(formatted, template.HTML("</ul>"))
				}
			}
		}

		return formatted
	},
}
