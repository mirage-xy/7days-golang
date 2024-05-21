package main

import (
    "flag"
    "fmt"
    "geecache"
    "log"
    "net/http"
)

var db = map[string]string{
    "Tom":  "630",
    "Jack": "589",
    "Sam":  "567",
}

func createGroup() *geecache.Group {
    return geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
        func(key string) ([]byte, error) {
            log.Println("[SlowDB] search key", key)
            if v, ok := db[key]; ok {
                return []byte(v), nil
            }
            return nil, fmt.Errorf("%s not exist", key)
        }))
}

// 用来启动会缓存服务器，创建HTTPPool，添加节点信息，注册到gee中，启动HTTP服务
func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
    peers := geecache.NewHTTPPool(addr)
    peers.Set(addrs...)
    gee.RegisterPeers(peers)
    log.Println("geecache is running at", addr)
    log.Fatal(http.ListenAndServe(addr[7:], peers))
}

// 用来启动API服务，与用户进行交互，用户感知
func startAPIServer(apiAddr string, gee *geecache.Group) {
    http.Handle("/api", http.HandlerFunc(
        func(w http.ResponseWriter, r *http.Request) {
            key := r.URL.Query().Get("key")
            view, err := gee.Get(key)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            w.Header().Set("Content-Type", "application/octet-stream")
            w.Write(view.ByteSlice())

        }))
    log.Println("fontend server is running at", apiAddr)
    log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

// main函数需要命令行传入port和api2个参数，用来在指定端口启动HTTP服务
func main() {
    var port int
    var api bool
    flag.IntVar(&port, "port", 8001, "Geecache server port")
    flag.BoolVar(&api, "api", false, "Start a api server?")
    flag.Parse()

    apiAddr := "http://localhost:9993"
    addrMap := map[int]string{
        8001: "http://localhost:8001",
        8002: "http://localhost:8002",
        8003: "http://localhost:8003",
    }

    var addrs []string
    for _, v := range addrMap {
        addrs = append(addrs, v)
    }

    gee := createGroup()
    if api {
        go startAPIServer(apiAddr, gee)
    }

    startCacheServer(addrMap[port], []string(addrs), gee)
}

// 2024/05/10 11:34:42 geecache is running at localhost:9993
//2024/05/10 11:36:51 [Server localhost:9993] GET /_geecache/scores/Tom
//2024/05/10 11:36:51 [SlowDB] search key Tom
//2024/05/10 11:37:12 [Server localhost:9993] GET /_geecache/scores/kkk
//2024/05/10 11:37:12 [SlowDB] search key kkk
