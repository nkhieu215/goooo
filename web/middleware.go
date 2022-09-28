// example of middlerware basic
package main

import "net/http"

// dinh nghia kieu interface
type middlerware func(http.Handler) http.Handler

// cau truc router
type router struct {
	// slice gồm các hàm middleware
	middlerwareChain []middlerware
	// mapping cấu trúc routing với name
	mux map[string]http.Handler
}

func NewRouter() *router {
	return &router{}
}

// mỗi khi gọi Use là thêm hàm middleware vào slice
func (r *router) Use(m middlerware) {
	r.middlerwareChain = append(r.middlerwareChain, m)// append tra ve 1 slice moi sau khi da them phan tu
	// ham append ko cap phat lai bo nho khi chua dat toi suc chua toi da 
}

// mỗi khi gọi Add là thêm phần routing trong đó, áp dụng các middleware vào
func (r *router) Add(route string, h http.Handler) {
	var mergedHandler = h
	// duyệt theo thứ tự ngược lại để apply middleware
	for i := len(r.middlerwareChain) - 1; i >= 0; i-- {
		mergedHandler = r.middlerwareChain[i](mergedHandler)
	}
	// cuối cùng register hàm handler vào name route tương ứng
	r.mux[route] = mergedHandler
}
