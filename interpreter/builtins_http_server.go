//go:build !js

package interpreter

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type httpRoute struct {
	method  string
	path    string
	handler Value
}

func builtinHTTPServe(e *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "httpServe expects config object"}
	}
	config, ok := objectPairs(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "httpServe expects config object"}
	}
	addrVal, ok := config["addr"]
	if !ok {
		return nil, &RuntimeError{Message: "httpServe expects addr"}
	}
	addr, ok := stringArg(addrVal)
	if !ok {
		return nil, &RuntimeError{Message: "httpServe addr must be string"}
	}
	routesVal, ok := config["routes"]
	if !ok {
		return nil, &RuntimeError{Message: "httpServe expects routes"}
	}
	routes, err := parseHTTPRoutes(routesVal)
	if err != nil {
		return nil, err
	}

	serverValue := &HTTPServer{}
	mux := http.NewServeMux()
	registerHTTPRoutes(mux, serverValue, e, routes)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, recoverableError("httpServe", "httpServe error: "+err.Error())
	}
	server := &http.Server{Addr: addr, Handler: mux}
	serverValue.Server = server

	go func() {
		err := server.Serve(listener)
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			return
		}
		if e != nil && e.runtime != nil {
			e.runtime.setFatalTaskFailure(&RuntimeError{Message: "httpServe error: " + err.Error()})
		}
	}()

	cancelCh := runtimeCancelSignal(e)
	fatalCh := runtimeFatalSignal(e)
	if cancelCh != nil || fatalCh != nil {
		go func() {
			select {
			case <-cancelCh:
			case <-fatalCh:
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			_ = serverValue.Stop(ctx)
		}()
	}

	return serverValue, nil
}

func builtinHTTPServerStop(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "httpServerStop expects server"}
	}
	server, ok := args[0].(*HTTPServer)
	if !ok {
		return nil, &RuntimeError{Message: "httpServerStop expects server"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := server.Stop(ctx); err != nil {
		return nil, recoverableError("httpServerStop", "httpServerStop error: "+err.Error())
	}
	return UnitValue, nil
}

func parseHTTPRoutes(val Value) ([]httpRoute, error) {
	arr, ok := val.(*Array)
	if !ok {
		return nil, &RuntimeError{Message: "httpServe routes must be array"}
	}
	if len(arr.Elements) == 0 {
		return nil, &RuntimeError{Message: "httpServe expects at least one route"}
	}
	out := make([]httpRoute, 0, len(arr.Elements))
	for _, entry := range arr.Elements {
		pairs, ok := objectPairs(entry)
		if !ok {
			return nil, &RuntimeError{Message: "httpServe route must be object"}
		}
		methodVal, ok := pairs["method"]
		if !ok {
			return nil, &RuntimeError{Message: "httpServe route expects method"}
		}
		method, ok := stringArg(methodVal)
		if !ok {
			return nil, &RuntimeError{Message: "httpServe route method must be string"}
		}
		pathVal, ok := pairs["path"]
		if !ok {
			return nil, &RuntimeError{Message: "httpServe route expects path"}
		}
		path, ok := stringArg(pathVal)
		if !ok {
			return nil, &RuntimeError{Message: "httpServe route path must be string"}
		}
		handler, ok := pairs["handler"]
		if !ok {
			return nil, &RuntimeError{Message: "httpServe route expects handler"}
		}
		if !isCallable(handler) {
			return nil, &RuntimeError{Message: "httpServe route handler must be callable"}
		}
		out = append(out, httpRoute{
			method:  strings.ToUpper(method),
			path:    path,
			handler: handler,
		})
	}
	return out, nil
}

func registerHTTPRoutes(mux *http.ServeMux, serverValue *HTTPServer, baseEval *Evaluator, routes []httpRoute) {
	byPath := make(map[string][]httpRoute)
	for _, route := range routes {
		byPath[route.path] = append(byPath[route.path], route)
	}
	for path, set := range byPath {
		routesAtPath := append([]httpRoute(nil), set...)
		handlerEval := &Evaluator{
			source:      baseEval.source,
			filename:    baseEval.filename,
			projectRoot: baseEval.projectRoot,
			modules:     baseEval.modules,
			runtime:     baseEval.runtime,
		}
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			route, ok := selectRouteByMethod(routesAtPath, r.Method)
			if !ok {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			reqVal, err := buildHTTPRequestValue(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			serverValue.evalMu.Lock()
			response, sig, err := handlerEval.applyFunction(route.handler, []Value{reqVal})
			serverValue.evalMu.Unlock()

			if sig != nil {
				http.Error(w, "invalid handler signal", http.StatusInternalServerError)
				return
			}
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			status, headers, body, err := parseHTTPHandlerResponse(response)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			for k, v := range headers {
				w.Header().Set(k, v)
			}
			w.WriteHeader(status)
			_, _ = io.WriteString(w, body)
		})
	}
}

func selectRouteByMethod(routes []httpRoute, method string) (httpRoute, bool) {
	method = strings.ToUpper(method)
	for _, route := range routes {
		if route.method == method {
			return route, true
		}
	}
	return httpRoute{}, false
}

func buildHTTPRequestValue(r *http.Request) (Value, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	headers := map[string]Value{}
	for key, values := range r.Header {
		headers[key] = &String{Value: strings.Join(values, ", ")}
	}
	query := map[string]Value{}
	for key, values := range r.URL.Query() {
		if len(values) == 0 {
			query[key] = &String{Value: ""}
			continue
		}
		query[key] = &String{Value: values[0]}
	}
	return &Object{Pairs: map[string]Value{
		"method":     &String{Value: r.Method},
		"path":       &String{Value: r.URL.Path},
		"query":      &Object{Pairs: query},
		"headers":    &Object{Pairs: headers},
		"body":       &String{Value: string(body)},
		"remoteAddr": &String{Value: r.RemoteAddr},
	}}, nil
}

func parseHTTPHandlerResponse(val Value) (int, map[string]string, string, error) {
	pairs, ok := objectPairs(val)
	if !ok {
		return 0, nil, "", &RuntimeError{Message: "httpServe handler must return object"}
	}
	status := 200
	if statusVal, ok := pairs["status"]; ok {
		statusInt, ok := statusVal.(*Integer)
		if !ok {
			return 0, nil, "", &RuntimeError{Message: "httpServe handler status must be integer"}
		}
		if statusInt.Value < 100 || statusInt.Value > 999 {
			return 0, nil, "", &RuntimeError{Message: "httpServe handler status out of range"}
		}
		status = int(statusInt.Value)
	}
	headers := map[string]string{}
	if headersVal, ok := pairs["headers"]; ok && headersVal != NullValue {
		parsed, err := extractHeaders(headersVal)
		if err != nil {
			return 0, nil, "", err
		}
		headers = parsed
	}
	body := ""
	if bodyVal, ok := pairs["body"]; ok && bodyVal != NullValue {
		bodyText, ok := stringArg(bodyVal)
		if !ok {
			return 0, nil, "", &RuntimeError{Message: "httpServe handler body must be string"}
		}
		body = bodyText
	}
	return status, headers, body, nil
}

func isCallable(val Value) bool {
	switch val.(type) {
	case *Function, *Builtin, *Partial:
		return true
	default:
		return false
	}
}
