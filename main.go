package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

type Config struct {
	PortMappings map[string]string `yaml:"port_mappings"`
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &config, nil
}

// 处理单个 TCP 连接
func handleConnection(clientConn net.Conn, targetAddr string) {
	defer func(clientConn net.Conn) {
		err := clientConn.Close()
		if err != nil {
			log.Println("关闭客户端连接失败:", err)
		} else {
			log.Println("关闭客户端连接成功")
		}
	}(clientConn)

	// 连接到目标 MySQL 服务器
	serverConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Println("无法连接到 MySQL 服务器:", targetAddr, "错误:", err)
		return
	}
	defer func(serverConn net.Conn) {
		err = serverConn.Close()
		if err != nil {
			log.Println("关闭服务器连接失败:", err)
		} else {
			log.Println("关闭服务器连接成功")
		}
	}(serverConn)

	// 开启双向数据转发
	go func() {
		_, err = io.Copy(serverConn, clientConn)
		if err != nil {
			log.Println("#1 数据转发失败:", err)
			return
		}
	}()
	_, err = io.Copy(clientConn, serverConn)
	if err != nil {
		log.Println("#2 数据转发失败:", err)
		return
	}
}

// 启动 TCP 监听
func startTCPListener(port string, targetAddr string, wg *sync.WaitGroup) {
	defer wg.Done()

	listenAddr := "0.0.0.0:" + port
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Printf("端口 %s 监听失败: %v\n", port, err)
		return
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Printf("关闭监听失败: %v\n", err)
		} else {
			log.Printf("端口 %s 监听已关闭\n", port)
		}
	}(listener)

	log.Printf("代理服务器启动，监听端口 %s，目标数据库 %s\n", port, targetAddr)

	for {
		clientConn, err := listener.Accept()
		if err != nil {
			log.Println("接受连接失败:", err)
			continue
		}
		go handleConnection(clientConn, targetAddr)
	}
}

func main() {
	config, err := loadConfig("config.yml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	var wg sync.WaitGroup

	// 遍历端口映射，启动多个监听
	for port, targetAddr := range config.PortMappings {
		wg.Add(1)
		go startTCPListener(port, targetAddr, &wg)
	}

	wg.Wait()
}
