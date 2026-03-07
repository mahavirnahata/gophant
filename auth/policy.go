package auth

import (
	"reflect"
	"strings"

	gomvchttp "github.com/mahavirnahata/gophant/http"
)

type PolicyRegistrar struct {
	Gate *Gate
}

func NewPolicyRegistrar(g *Gate) *PolicyRegistrar {
	return &PolicyRegistrar{Gate: g}
}

func (r *PolicyRegistrar) Register(prefix string, policy any) {
	if r.Gate == nil || policy == nil {
		return
	}
	pv := reflect.ValueOf(policy)
	pt := pv.Type()
	for i := 0; i < pt.NumMethod(); i++ {
		m := pt.Method(i)
		if m.Type.NumIn() != 2 {
			continue
		}
		if m.Type.In(1) != reflect.TypeOf(&gomvchttp.Context{}) {
			continue
		}
		if m.Type.NumOut() != 1 || m.Type.Out(0).Kind() != reflect.Bool {
			continue
		}
		ability := methodToAbility(m.Name)
		name := prefix + "." + ability
		r.Gate.Define(name, func(c *gomvchttp.Context) bool {
			out := m.Func.Call([]reflect.Value{pv, reflect.ValueOf(c)})
			return out[0].Bool()
		})
	}
}

func methodToAbility(name string) string {
	switch name {
	case "ViewAny":
		return "viewAny"
	case "View":
		return "view"
	case "Create":
		return "create"
	case "Update":
		return "update"
	case "Delete":
		return "delete"
	default:
		return strings.ToLower(name[:1]) + name[1:]
	}
}
