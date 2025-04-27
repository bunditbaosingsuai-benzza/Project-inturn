package main

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"net"
	"regexp"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	tcpServerAddr = "0.0.0.0:5050"
	redisAddr     = "localhost:6379"
	redisPassword = ""
	redisDB       = 0
	ipSetKey      = "dangerous_ips"
)

type LogEntry struct {
	Source  string `json:"source"`
	IP      string `json:"ip"`
	Message string `json:"message"`
}

func main() {
	ctx := context.Background()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Unable to connect to Redis: %v", err)
	}
	log.Println("Successfully connected to Redis")

	listener, err := net.Listen("tcp", tcpServerAddr)
	if err != nil {
		log.Fatalf("Unable to start TCP server: %v", err)
	}
	defer listener.Close()

	log.Printf("Waiting for log data %s", tcpServerAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		go handleConnection(conn, redisClient, ctx)
	}
}

func handleConnection(conn net.Conn, redisClient *redis.Client, ctx context.Context) {
	defer conn.Close()

	log.Printf("Get connected from %s", conn.RemoteAddr().String())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		logData := scanner.Text()
		processLog(logData, redisClient, ctx)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading data: %v", err)
	}
}

func processLog(logData string, redisClient *redis.Client, ctx context.Context) {

	var logEntry LogEntry
	if err := json.Unmarshal([]byte(logData), &logEntry); err != nil {

		extractAndProcessIP(logData, redisClient, ctx)
		return
	}

	if logEntry.IP != "" {
		checkAndDisplayIP(logEntry.IP, redisClient, ctx)
	} else {

		extractAndProcessIP(logData, redisClient, ctx)
	}
}

func extractAndProcessIP(logData string, redisClient *redis.Client, ctx context.Context) {

	ipRegex := regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	ips := ipRegex.FindAllString(logData, -1)

	for _, ip := range ips {
		checkAndDisplayIP(ip, redisClient, ctx)
	}
}

func checkAndDisplayIP(ip string, redisClient *redis.Client, ctx context.Context) {

	if !isValidIP(ip) {
		return
	}

	isDangerous := checkIfDangerousIP(ip, redisClient, ctx)

	if isDangerous {

		log.Printf("⚠️ dangerous IP: %s - Date Time: %s", ip, time.Now().Format(time.RFC1123))
	} else {

		log.Printf(" IP: %s (Not dangerous)", ip)
	}
}

func isValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

func checkIfDangerousIP(ip string, redisClient *redis.Client, ctx context.Context) bool {
	isDangerous, err := redisClient.SIsMember(ctx, ipSetKey, ip).Result()
	if err != nil {
		log.Printf("Failed to check IP in Redis: %v", err)
		return false
	}
	return isDangerous
}
