package gofw

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	NOTFOUND string = "NOT FOUND"
	FOUND string = "FOUND"
)

type (
	HandlerFunc func(*Context)

	Router struct {
		tree map[string]*node
	}

	//todo 结构优化
	node struct {
		label byte
		prefix string
		child []*node
		handlerFunc HandlerFunc
	}
)

func NewRouter() *Router {
	router := &Router {
		tree: make(map[string]*node),
	}
	return router
}

func (r *Router) insert(method, path string, h HandlerFunc) {
	//todo map 判断标准方法
	n := r.tree[method]
	if n == nil {
		//初始化struct的方法
		n = &node{}
	}

	n.Add(path, h)
	fmt.Println(n)
	r.tree[method] = n
}

func (n *node) findChildWithLabel(l byte) *node {
	for _, c := range n.child {
		if c.label == l {
			return c
		}
	}
	return nil
}

func (n *node) Add(path string, h HandlerFunc) {
	search := path

	for {
		sl := len(search)
		pl := len(n.prefix)
		l := 0

		max := pl
		if sl < max {
			max = sl
		}

		// 这一句话不是没用的，是为了计算l
		for; l < max && search[l] == n.prefix[l]; l++ {

		}

		// 其实第一种情况只会在第一次插入路由的情况出现，因为所有path的第一个字符都是“/”
		if l == 0 {
			n.label = search[0]
			n.prefix = search
			if h != nil {
				n.handlerFunc = h
			}
		} else if l < pl {
			n1 := &node {
				n.prefix[l:][0],
				n.prefix[l:],
				n.child,
				n.handlerFunc,
			}
			fmt.Println("n1:", n1)
			n.label = n.prefix[0]
			n.prefix = n.prefix[:l]

			n.child = append(n.child, n1)

			// 被包含的情况，新添加的node变父亲，原来的变儿子
			if l == sl {
				n.handlerFunc = h
			} else {
				prefix := search[l:]
				n2 := &node{
					prefix[0],
					prefix,
					nil,
					h,
				}
				n.child = append(n.child, n2)
			}
		} else if l < sl {
			search = search[l:]
			c := n.findChildWithLabel(search[0])
			if c != nil {
				// 继续搜索
				n = c
				continue
			}
			n1 := &node {
				search[0],
				search,
				nil,
				h,
			}
			n.child = append(n.child, n1)
		} else {
			if h != nil {
				n.handlerFunc = h
			}
		}
		return
	}
}

func (n *node) Find(path string) (*node, string) {
	var search = path

	for {
		// 一下一连串的10行代码得到两个字符串的交集 eg: abc&ab = ab
		l := 0
		sl := len(search)
		pl := len(n.prefix)

		max := pl
		if sl < max {
			max = sl
		}
		for; l < max && search[l] == n.prefix[l]; l++ {

		}

		if l == pl {
			search = search[l:]
		} else {
			// 参数路由
			for _, v := range n.child {
				if v.label == ':' {
					return v, search
				}
			}
			// 全部匹配
			for _, v := range n.child {
				if v.label == '*' {
					return v, ""
				}
			}
			return nil, ""
		}
		// 绝对路由
		if search == "" {
			return n, ""
		}
		// child中的接点通过label来查找
		if n1 := n.findChildWithLabel(search[0]); n1 != nil {
			n = n1
			continue
		}
	}
}

func (r *Router) Add(method, path string, h HandlerFunc) {
	if path == "" {
		panic("gofw: path cannot be empty")
	}
	//为了有个根字符
	if path[0] != '/' {
		path = "/" + path
	}

	for i, l := 0, len(path); i < l; i++ {
		if path[i] == ':' {
			if path[i-1] == '/' {
				r.insert(method, path[:i], nil)
				r.insert(method, path, h)
				return
			}
		} else if path[i] == '*' {
			if path[i-1] == '/' {
				r.insert(method, path[:i], nil)
				r.insert(method, path, h)
				return
			}
		}
	}

	r.insert(method, path, h)
}

func (r *Router) ServeHTTP(rw http.ResponseWriter, re *http.Request) {
	ctx := &Context {
		Request: re,
		Response: rw,
	}
	// 程序执行的异常， nil handlerFunc
	defer Recover(ctx)

	t := time.Now()
	method := re.Method
	uri := re.URL.Path

	tree, ok := r.tree[method]
	// 避免nil 的panic
	if !ok {
		http.NotFound(rw, re)
		return
	}
	n, paramname := tree.Find(uri)
	if n == nil || n.handlerFunc == nil {
		http.NotFound(rw, re)
		dur := time.Since(t)
		log.Printf("\033[31;1m%s %10s %10s %10s\033[0m", method, uri, dur.String(), NOTFOUND)
		return
	}

	// 设置参数获取
	if paramname != "" {
		ctx.setParamname(n.prefix[1:])
		ctx.setParamvalue(paramname)
	}
	n.handlerFunc(ctx)
	dur := time.Since(t)
	log.Printf("\033[32;1m%s %10s %10s %10s\033[0m", method, uri, dur.String(), FOUND)
}