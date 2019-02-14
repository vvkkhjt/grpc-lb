package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/resolver"

	//"github.com/coreos/etcd/clientv3"
	"strconv"
	"time"

	grpclb "aranya/grpc-lb/etcdv3"
	pb "aranya/grpc-lb/cmd/helloworld"
	"google.golang.org/grpc"
)

var (
	serv = flag.String("service", "hello_service", "service name")
	reg  = flag.String("reg", "http://localhost:2379", "register etcd address")
)

func main() {
	flag.Parse()
	// 生成命名解析
	r := grpclb.NewResolver(*reg)
	// 注册命名
	resolver.Register(r)
	// 生成链接保活
	keepAlive := keepalive.ClientParameters{
		10 * time.Second,
		20 * time.Second,
		true,
	}

	// grpc.WithInsecure: 不使用安全连接
	// grpc.WithBalancerName("round_robin"), 轮询机制做负载均衡
	// grpc.WithBlock: 握手成功才返回
	// grpc.WithKeepaliveParams: 连接保活，防止因为长时间闲置导致连接不可用
	conn, err := grpc.Dial(r.Scheme()+":///"+*serv,grpc.WithInsecure(),
		grpc.WithBalancerName(roundrobin.Name),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepAlive))
	if err != nil {
		panic(err)
	}

	client := pb.NewGreeterClient(conn)

	ticker := time.NewTicker(1000 * time.Millisecond)
	for t := range ticker.C {
		resp, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "world " + strconv.Itoa(t.Second())})
		if err == nil {
			fmt.Printf("%v: Reply is %s\n", t, resp.Message)
		}
	}
}
