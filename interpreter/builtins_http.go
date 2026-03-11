package interpreter

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
)

func registerHTTPBuiltins() {
	builtins["http"] = &Builtin{Name: "http", Fn: builtinHTTP}
	builtins["httpServe"] = &Builtin{Name: "httpServe", Fn: builtinHTTPServe}
	builtins["httpServerStop"] = &Builtin{Name: "httpServerStop", Fn: builtinHTTPServerStop}
}

func builtinHTTP(e *Evaluator, args []Value) (Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, &RuntimeError{Message: "http expects request object or (url, opts?)"}
	}
	if urlStr, ok := stringArg(args[0]); ok {
		method := "GET"
		mode := streamTypeBytes
		body := ""
		headers := map[string]string{}

		if len(args) == 2 && !Equivalent(args[1], NullValue) {
			opts, ok := objectPairs(args[1])
			if !ok {
				return nil, &RuntimeError{Message: "http source opts must be object"}
			}
			if err := rejectUnknownObjectKeys(opts, "http source opts", []string{"method", "headers", "body", "type"}); err != nil {
				return nil, err
			}
			if methodVal, ok := opts["method"]; ok && !Equivalent(methodVal, NullValue) {
				parsed, ok := stringArg(methodVal)
				if !ok {
					return nil, &RuntimeError{Message: "http source method must be string"}
				}
				method = strings.ToUpper(strings.TrimSpace(parsed))
				if method == "" {
					return nil, &RuntimeError{Message: "http source method must not be empty"}
				}
			}
			if headersVal, ok := opts["headers"]; ok && !Equivalent(headersVal, NullValue) {
				parsed, err := extractHeaders(headersVal)
				if err != nil {
					return nil, err
				}
				headers = parsed
			}
			if bodyVal, ok := opts["body"]; ok && !Equivalent(bodyVal, NullValue) {
				parsed, ok := stringArg(bodyVal)
				if !ok {
					return nil, &RuntimeError{Message: "http source body must be string"}
				}
				body = parsed
			}
			if typeVal, ok := opts["type"]; ok && !Equivalent(typeVal, NullValue) {
				parsed, err := parseStreamType(typeVal, "type")
				if err != nil {
					return nil, err
				}
				mode = parsed
			}
		}

		return &StreamSourceValue{
			name: "http",
			open: func(e *Evaluator) (streamIterator, error) {
				reqDone := make(chan struct{})
				ctx, cancel := context.WithCancel(context.Background())
				cancelCh := runtimeCancelSignal(e)
				fatalCh := runtimeFatalSignal(e)
				runGuarded(e.runtime, "http stream cancel watcher", func() {
					select {
					case <-cancelCh:
						cancel()
					case <-fatalCh:
						cancel()
					case <-reqDone:
						cancel()
					}
				})

				var reqBody io.Reader
				if body != "" {
					reqBody = strings.NewReader(body)
				}
				req, err := http.NewRequestWithContext(ctx, method, urlStr, reqBody)
				if err != nil {
					close(reqDone)
					cancel()
					return nil, recoverableError("http", "http request error: "+err.Error())
				}
				for k, v := range headers {
					req.Header.Set(k, v)
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					close(reqDone)
					cancel()
					if errors.Is(err, context.Canceled) {
						return nil, canceledError()
					}
					return nil, recoverableError("http", "http error: "+err.Error())
				}

				reader := &StreamReader{
					reader: resp.Body,
					closer: resp.Body,
					mode:   mode,
				}
				base := &streamReaderIterator{
					reader:      reader,
					size:        defaultStreamReadSize,
					closeOnExit: true,
				}
				return &cleanupStreamIterator{
					upstream: base,
					cleanup: func() {
						close(reqDone)
						cancel()
					},
				}, nil
			},
		}, nil
	}
	if len(args) != 1 {
		return nil, &RuntimeError{Message: "http expects request object"}
	}
	reqObj, ok := objectPairs(args[0])
	if !ok {
		return nil, &RuntimeError{Message: "http expects object request"}
	}
	methodVal, ok := reqObj["method"]
	if !ok {
		methodVal = &String{Value: "GET"}
	}
	method, ok := stringArg(methodVal)
	if !ok {
		return nil, &RuntimeError{Message: "http method must be string"}
	}
	urlVal, ok := reqObj["url"]
	if !ok {
		return nil, &RuntimeError{Message: "http expects url"}
	}
	urlStr, ok := stringArg(urlVal)
	if !ok {
		return nil, &RuntimeError{Message: "http url must be string"}
	}
	var body io.Reader
	if bodyVal, ok := reqObj["body"]; ok && bodyVal != NullValue {
		bodyStr, ok := stringArg(bodyVal)
		if !ok {
			return nil, &RuntimeError{Message: "http body must be string"}
		}
		body = strings.NewReader(bodyStr)
	}

	reqDone := make(chan struct{})
	defer close(reqDone)

	ctx, cancel := context.WithCancel(context.Background())
	cancelCh := runtimeCancelSignal(e)
	fatalCh := runtimeFatalSignal(e)
	runGuarded(e.runtime, "http cancel watcher", func() {
		select {
		case <-cancelCh:
			cancel()
		case <-fatalCh:
			cancel()
		case <-reqDone:
			cancel()
		}
	})

	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return nil, recoverableError("http", "http request error: "+err.Error())
	}
	if headersVal, ok := reqObj["headers"]; ok && headersVal != NullValue {
		headers, err := extractHeaders(headersVal)
		if err != nil {
			return nil, err
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			if cancelCh != nil {
				select {
				case <-cancelCh:
					return nil, canceledError()
				default:
				}
			}
			if fatalCh != nil {
				select {
				case <-fatalCh:
					return nil, runtimeFatalError(e)
				default:
				}
			}
			return nil, canceledError()
		}
		return nil, recoverableError("http", "http error: "+err.Error())
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, recoverableError("http", "http read error: "+err.Error())
	}
	return httpResponseObject(resp, data), nil
}

type cleanupStreamIterator struct {
	upstream streamIterator
	cleanup  func()
	once     sync.Once
}

func (i *cleanupStreamIterator) Next() (Value, bool, error) {
	if i == nil || i.upstream == nil {
		return nil, true, nil
	}
	return i.upstream.Next()
}

func (i *cleanupStreamIterator) Close() error {
	if i == nil {
		return nil
	}
	if i.cleanup != nil {
		i.once.Do(i.cleanup)
	}
	if i.upstream != nil {
		return i.upstream.Close()
	}
	return nil
}
