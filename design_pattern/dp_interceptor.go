package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

func exampleInterceptor() {
	println("exampleInterceptor")

	s := NewSimpleHttpSvr()
	s.HandleFunc("/v1", func(writer http.ResponseWriter, request *http.Request) error {
		fmt.Fprintf(writer, "v1's content!\n")
		return nil
	})
	s.AddInterceptor(Auth, Logging)
	//panic(http.ListenAndServe(":8080", s))

	// 本地测试
	ts := httptest.NewServer(s)
	defer ts.Close()

	req, err := http.NewRequest("POST", ts.URL+"/v1", nil)
	if err != nil {
		log.Fatalf("http.NewRequest: %v", err)
	}
	req.Header.Set("Authorization", "Bearer xxx")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("http.Get: %v", err)
	}
	defer resp.Body.Close()
	assertEqual(resp.StatusCode, http.StatusOK)

	b, _ := io.ReadAll(resp.Body)
	assertEqual(string(b), "v1's content!\n")
}

// Logging 日志拦截器
func Logging(r *http.Request, w http.ResponseWriter, invoker HandlerFunc) error {
	//fmt.Printf("Before-Logging: %s\n", r.RequestURI)
	err := invoker(w, r)
	if err == nil {
		//fmt.Printf("After-Logging: [OK] %s -200-\n", r.RequestURI)
		return nil
	}

	code := http.StatusInternalServerError
	_code := r.Context().Value("code")
	if c, _ := _code.(int); c > 0 {
		code = c
	}
	_ = code
	//fmt.Printf("After-Logging: [ERR] %s -%d- [%s]\n", r.RequestURI, code, err)
	return err
}

// Auth 拦截器
func Auth(r *http.Request, w http.ResponseWriter, invoker HandlerFunc) error {
	// do auth ...
	// 若认证失败，可以返回错误，而不会继续处理请求
	if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
		r.WithContext(context.WithValue(r.Context(), "code", 401))
		//fmt.Printf("Before-Auth: [FAIL] %s\n", r.RequestURI)
		return fmt.Errorf("invalid authorization Header format")
	}
	//fmt.Printf("Before-Auth: [OK] %s\n", r.RequestURI)
	return invoker(w, r)
}

type HandlerFunc func(http.ResponseWriter, *http.Request) error

type InterceptorFunc func(r *http.Request, w http.ResponseWriter, invoker HandlerFunc) error

type simpleHttpServer struct {
	mux              map[string]HandlerFunc
	interceptors     []InterceptorFunc
	unaryInterceptor InterceptorFunc
}

func (s *simpleHttpServer) HandleFunc(route string, handler func(http.ResponseWriter, *http.Request) error) {
	s.mux[route] = handler
}

func (s *simpleHttpServer) AddInterceptor(inter ...InterceptorFunc) {
	s.interceptors = append(s.interceptors, inter...)
	// 需要将拦截器链组合为一个拦截器链
	s.unaryInterceptor = UnaryInterceptor(s.interceptors...)
}

func (s *simpleHttpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := s.mux[r.RequestURI]
	if h == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err := s.unaryInterceptor(r, w, h)
	if err != nil {
		code := r.Context().Value("code")
		if c, _ := code.(int); c > 0 {
			http.Error(w, err.Error(), c)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
}

// UnaryInterceptor 组合多个拦截器为一个拦截器链
func UnaryInterceptor(interceptors ...InterceptorFunc) InterceptorFunc {
	return func(r *http.Request, w http.ResponseWriter, invoker HandlerFunc) error {
		// 定义一个递归函数，用于依次执行拦截器链中的拦截器
		var chain func(r *http.Request, w http.ResponseWriter, i int) error
		chain = func(r *http.Request, w http.ResponseWriter, i int) error {
			if i == len(interceptors) {
				// 所有拦截器已执行完毕，调用原始处理函数
				return invoker(w, r)
			}
			// 执行当前拦截器，并将下一个拦截器作为参数传递给它
			return interceptors[i](r, w, func(resp http.ResponseWriter, req *http.Request) error {
				return chain(req, resp, i+1)
			})
		}
		// 开始执行拦截器链
		return chain(r, w, 0)
	}
}

func NewSimpleHttpSvr() *simpleHttpServer {
	return &simpleHttpServer{mux: make(map[string]HandlerFunc), unaryInterceptor: func(r *http.Request, w http.ResponseWriter, invoker HandlerFunc) error {
		return invoker(w, r)
	}}
}
