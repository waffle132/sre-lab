package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ─── 数据 ────────────────────────────────────────

// 一条短链的结构
type shortEntry struct {
	URL  string
	Hits int64
}

// 全局内存表：短码 → 短链信息
var store = make(map[string]shortEntry)

// ─── 指标 ────────────────────────────────────────

var httpRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests",
	},
	[]string{"method", "path", "status"},
)

var httpDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: "http_request_duration_seconds",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method", "path"},
)

var usEntries = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "us_entries",
	},
)

// ─── 五个接口 ────────────────────────────────────

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("读取请求体失败:", err)
	}
	defer r.Body.Close()

	var data struct{ URL string }
	err = json.Unmarshal(body, &data)
	if err != nil {
		fmt.Println("解析JSON失败:", err)
	}
	log.Println("url地址为", data.URL)

	var code string
	code = randomCode()
	_, ok := store[code]
	for ok {
		code = randomCode()
		_, ok = store[code]
	}

	store[code] = shortEntry{URL: data.URL}
	usEntries.Inc()
	httpRequests.WithLabelValues("POST", "/shorten", "200").Inc()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"code": "%s"}`, code)
	httpDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	code := r.PathValue("code")

	entry, ok := store[code]
	if ok {
		entry.Hits += 1
		store[code] = entry
	} else {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Location", entry.URL)
	w.WriteHeader(http.StatusFound)
	httpDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
	httpRequests.WithLabelValues("GET", "/r/{code}", "302").Inc()
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	entriesMap := map[string]interface{}{}
	for code, entry := range store {
		entriesMap[code] = map[string]interface{}{
			"url":  entry.URL,
			"hits": entry.Hits,
		}
	}
	result := map[string]interface{}{
		"total_entries": len(store),
		"entries":       entriesMap,
	}
	jsonbytes, _ := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonbytes)
	httpDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
	httpRequests.WithLabelValues("GET", "/stats", "200").Inc()
}

func metricsHandler() http.Handler {
	return promhttp.Handler()
}

// ─── 工具 ────────────────────────────────────────

func randomCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// ─── 启动 ────────────────────────────────────────

func main() {
	rand.Seed(time.Now().UnixNano())

	prometheus.MustRegister(httpRequests)
	prometheus.MustRegister(usEntries)
	prometheus.MustRegister(httpDuration)

	http.HandleFunc("GET /healthz", healthzHandler)
	http.HandleFunc("POST /shorten", shortenHandler)
	http.HandleFunc("GET /r/{code}", redirectHandler)
	http.HandleFunc("GET /stats", statsHandler)
	http.Handle("GET /metrics", metricsHandler())

	fmt.Println("u-short 启动在 :8080")
	http.ListenAndServe(":8080", nil)
}