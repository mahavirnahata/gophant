package cli

import (
	"fmt"
	"strings"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

// RouteList prints all registered routes to stdout in a table format.
func RouteList(router *gomvchttp.Router) {
	routes := router.Routes()
	if len(routes) == 0 {
		fmt.Println("No routes registered.")
		return
	}

	// Compute column widths.
	methodW, patternW, nameW := len("METHOD"), len("PATTERN"), len("NAME")
	for _, r := range routes {
		if len(r.Method) > methodW {
			methodW = len(r.Method)
		}
		if len(r.Pattern) > patternW {
			patternW = len(r.Pattern)
		}
		if len(r.Name) > nameW {
			nameW = len(r.Name)
		}
	}

	fmt.Printf("%-*s  %-*s  %-*s\n", methodW, "METHOD", patternW, "PATTERN", nameW, "NAME")
	fmt.Printf("%s  %s  %s\n", strings.Repeat("-", methodW), strings.Repeat("-", patternW), strings.Repeat("-", nameW))
	for _, r := range routes {
		fmt.Printf("%-*s  %-*s  %-*s\n", methodW, r.Method, patternW, r.Pattern, nameW, r.Name)
	}
}
