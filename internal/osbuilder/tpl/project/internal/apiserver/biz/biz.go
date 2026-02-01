{{- $D := or .Web .MQ .Job -}}
package biz

import (
	"github.com/google/wire"
    {{- if and .Web .Web.WithUser }}
    "github.com/onexstack/onexstack/pkg/authz"
    {{- end}}

    "{{.M.ModuleName}}/internal/{{$D.Name}}/store"
    {{- if and .Web .Web.WithUser }}
    userv1 "{{.M.ModuleName}}/internal/{{$D.Name}}/biz/v1/user"
    {{- end}}
    {{- if $D.Clients }}
    "{{.M.ModuleName}}/internal/{{$D.Name}}/pkg/clientset"
    {{- end}}
    {{- if and .Web .Web.WithWS }}
    wsv1 "{{.M.ModuleName}}/internal/{{$D.Name}}/biz/v1/websocket"
    {{- end}}
)

// ProviderSet declares dependency injection rules for the business logic layer.
var ProviderSet = wire.NewSet(NewBiz, wire.Bind(new(IBiz), new(*biz)))

// IBiz defines the access points for various business logic modules.
type IBiz interface {
    {{- if and .Web .Web.WithUser }}
    // UserV1 gets the user business interface.
    UserV1() userv1.UserBiz
    {{- end}}
    {{- if and .Web .Web.WithWS }}
    // WSV1 gets the WebSocket related interface.
    WSV1() wsv1.WSBiz
    {{- end}}
}

// biz is the concrete implementation of the business logic IBiz.
type biz struct {
    store store.IStore
    {{- if and .Web .Web.WithUser }}
    authz *authz.Authz
    {{- end}}
    {{- if $D.Clients }}
    clientset clientset.Interface
    {{- end}}
}

// Ensure biz implements IBiz at compile time.
var _ IBiz = (*biz)(nil)

// NewBiz creates and returns a new instance of the business logic layer.
func NewBiz(store store.IStore{{- if and .Web .Web.WithUser }}, authz *authz.Authz{{- end -}}{{- if $D.Clients }}, clientset clientset.Interface{{- end -}}) *biz {
    return &biz{store: store{{- if and .Web .Web.WithUser }}, authz: authz{{end}}{{- if $D.Clients }}, clientset: clientset{{- end -}}}
}

{{- if and .Web .Web.WithUser }}
// UserV1 returns an instance that implements the UserBiz interface.
func (b *biz) UserV1() userv1.UserBiz {
    return userv1.New(b.store, b.authz)
}
{{- end}}

{{- if and .Web .Web.WithWS }}
// WSV1 returns an instance that implements the WSBiz interface.
func (b *biz) WSV1() wsv1.WSBiz {
    return wsv1.New(b.store{{- if $D.Clients }}, b.clientset{{- end -}})
}
{{- end}}
