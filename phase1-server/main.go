package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"
	"database/sql"
	_ "modernc.org/sqlite"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ─── 数据 ────────────────────────────────────────

// 一条短链的结构
type shortEntry struct {
	URL  string
	Hits int64
}


var db *sql.DB

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
		Help : ".005 .01 .025 .05",
	},
	[]string{"method", "path"},
)

var usEntries = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "us_entries_total",
		Help:  "Total shortened URLs created",
	},
)

func codeExists(code string) (bool, error) {
	var found string
	err := db.QueryRow("SELECT code FROM short_link WHERE code = ?", code).Scan(&found)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}


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
	ok, err := codeExists(code)
	if err != nil {
		log.Printf("检查短码是否存在失败: %v", err)
		http.Error(w,"检查短码是否存在失败", http.StatusInternalServerError)
		return
	} else {
		for ok {
			code = randomCode()
			ok, err = codeExists(code)
			if err != nil {
				log.Printf("检查短码是否存在失败: %v", err)
				http.Error(w,"检查短码是否存在失败", http.StatusInternalServerError)
				return
			}
		}
	}
	
	_, err = db.Exec(
		"INSERT INTO short_link (code ,url) VALUES (?, ?)",
		code,
		data.URL,
	)
	if err != nil {
		log.Printf("保存短链失败: %v", err)
		http.Error(w,"保存短链失败", http.StatusInternalServerError)
		return
	}
	usEntries.Inc()
	httpRequests.WithLabelValues("POST", "/shorten", "200").Inc()

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"code": "%s"}`, code)
	httpDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	code := r.PathValue("code")

    var entry shortEntry
	err := db.QueryRow(
		"SELECT url, hits FROM short_link WHERE code = ?",
		code,).Scan(&entry.URL, &entry.Hits)

		if err == sql.ErrNoRows {
			httpRequests.WithLabelValues("GET", "/r/{code}", "404").Inc()
			http.Error(w, "短码不存在", http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("查询短码失败: %v", err)
			http.Error(w, "查询短码失败", http.StatusInternalServerError)
			return
		}
	_, err = db.Exec(
		"UPDATE short_link SET hits = hits + 1 WHERE code = ?",
		code,
	)
	if err != nil {
		log.Printf("更新点击次数失败: %v", err)
		http.Error(w, "更新点击次数失败", http.StatusInternalServerError)
		return
	}	
	w.Header().Set("Location", entry.URL)
	w.WriteHeader(http.StatusFound)
	httpDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())
	httpRequests.WithLabelValues("GET", "/r/{code}", "302").Inc()
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()

    rows, err := db.Query("SELECT code, url, hits FROM short_link")
    if err != nil {
        log.Printf("查询统计失败: %v", err)
        http.Error(w, "查询统计失败", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    entriesMap := map[string]interface{}{}

    for rows.Next() {
        var code string
        var entry shortEntry

        err := rows.Scan(&code, &entry.URL, &entry.Hits)
        if err != nil {
            log.Printf("读取统计记录失败: %v", err)
            http.Error(w, "读取统计记录失败", http.StatusInternalServerError)
            return
        }

        entriesMap[code] = map[string]interface{}{
            "url":  entry.URL,
            "hits": entry.Hits,
        }
    }

    if err := rows.Err(); err != nil {
        log.Printf("遍历统计记录失败: %v", err)
        http.Error(w, "遍历统计记录失败", http.StatusInternalServerError)
        return
    }

    result := map[string]interface{}{
        "total_entries": len(entriesMap),
        "entries":       entriesMap,
    }

    jsonbytes, err := json.Marshal(result)
    if err != nil {
        log.Printf("生成统计 JSON 失败: %v", err)
        http.Error(w, "生成统计响应失败", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonbytes)

    httpDuration.WithLabelValues(r.Method, r.URL.Path).
        Observe(time.Since(start).Seconds())
    httpRequests.WithLabelValues("GET", "/stats", "200").Inc()
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
func metricsHandler() http.Handler {
    return promhttp.Handler()
}
// ─── 启动 ────────────────────────────────────────

func main() {
	var err error
	db, err = sql.Open("sqlite", "/data/u-short.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	_, err = db.Exec(
		` CREATE TABLE IF NOT EXISTS short_link(
			code TEXT PRIMARY KEY NOT NULL,
			url TEXT NOT NULL,
			hits INTEGER NOT NULL DEFAULT 0)
		
	`)
	if err != nil {
		log.Fatal(err)
	}

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
