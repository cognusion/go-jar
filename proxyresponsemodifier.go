package jar

import (
	"net/http"
)

// ProxyResponseModifier is a type interface compatible with oxy/forward, to allow the proxied response
// to be modified at proxy-time, before the Handlers will see the response. This is of special importance
// for responses which need absolute mangling before a response is completed e.g. streaming/chunked responses
type ProxyResponseModifier func(resp *http.Response) error

// ProxyResponseModifierChain is an encapsulating type to chain multiple ProxyResponseModifier funcs for
// sequential execution as a single ProxyResponseModifier
type ProxyResponseModifierChain struct {
	prms []ProxyResponseModifier
}

// Add appends the provided ProxyResponseModifier to the ProxyResponseModifierChain
func (p *ProxyResponseModifierChain) Add(prm ProxyResponseModifier) {
	p.prms = append(p.prms, prm)
}

// ToProxyResponseModifier returns a closure ProxyResponseModifier that will sequentially execute each
// encapsulated ProxyResponseModifier, discontinuing and returning an error as soon as one is noticed
func (p *ProxyResponseModifierChain) ToProxyResponseModifier() ProxyResponseModifier {
	return func(resp *http.Response) error {
		for _, f := range p.prms {
			if err := f(resp); err != nil {
				return err
			}
		}
		return nil
	}
}
