package link

import (
	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/ooclab/es/session"
	"github.com/ooclab/es/tunnel"
)

type requestHandler struct {
	router *session.Router
}

func newRequestHandler(routes []session.Route) *requestHandler {
	h := &requestHandler{
		router: session.NewRouter(),
	}
	h.router.AddRoutes([]session.Route{
		{"/echo", h.echo},
	})
	h.router.AddRoutes(routes)
	return h
}

func (h *requestHandler) Handle(m *session.EMSG) *session.EMSG {
	var resp *session.Response
	var err error

	req := &session.Request{}
	if err = json.Unmarshal(m.Payload, &req); err != nil {
		logrus.Errorf("json unmarshal session request failed: %s", err)
		resp = &session.Response{Status: "json-unmarshal-request-error"}
	} else {
		resp, err = h.router.Dispatch(req)
		if err != nil {
			logrus.Errorf("dispatch request failed: %s", err)
			resp = &session.Response{Status: "dispatch-request-error"}
		}
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		logrus.Errorf("json marshal response failed: %s", err)
		resp = &session.Response{Status: "json-marshal-response-error"}
	}

	return &session.EMSG{
		Type:    session.MsgTypeResponse,
		ID:      m.ID,
		Payload: payload,
	}
}

func (h *requestHandler) echo(req *session.Request) (*session.Response, error) {
	return &session.Response{Status: "success", Body: req.Body}, nil
}

func defaultTunnelCreateHandler(manager *tunnel.Manager) session.RequestHandlerFunc {
	return func(r *session.Request) (resp *session.Response, err error) {
		cfg := &tunnel.TunnelConfig{}
		if err = json.Unmarshal(r.Body, &cfg); err != nil {
			logrus.Errorf("tunnel create: unmarshal tunnel config failed: %s", err)
			resp.Status = "load-tunnel-map-error"
			return
		}

		logrus.Debugf("got config for tunnel create: %s", cfg)

		t, err := manager.TunnelCreate(cfg)
		if err != nil {
			logrus.Errorf("create tunnel failed: %s", err)
			resp.Status = "create-tunnel-failed"
			return
		}

		body, _ := json.Marshal(tunnelCreateBody{ID: t.ID})
		resp = &session.Response{
			Status: "success",
			Body:   body,
		}
		return
	}
}
