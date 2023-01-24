package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// https://github.com/restic/restic/blob/b90308180458f5d0bbe49b03fa8cc312c2e97127/cmd/restic/cmd_stats.go#L304
type restoreDataStats struct {
	TotalSize      uint64 `json:"total_size"`
	TotalFileCount uint64 `json:"total_file_count,omitempty"`
	SnapshotsCount int    `json:"snapshots_count"`
}

type rawDataStats struct {
	TotalSize              uint64  `json:"total_size"`
	TotalUncompressedSize  uint64  `json:"total_uncompressed_size,omitempty"`
	CompressionRatio       float64 `json:"compression_ratio,omitempty"`
	CompressionProgress    float64 `json:"compression_progress,omitempty"`
	CompressionSpaceSaving float64 `json:"compression_space_saving,omitempty"`
	TotalBlobCount         uint64  `json:"total_blob_count,omitempty"`
	SnapshotsCount         int     `json:"snapshots_count"`
}

var (
	snapshotsTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_snapshot_total",
		Help: "Number of snapshots",
	}, []string{"repository"})
	restoreSizeTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_restore_size_bytes_total",
		Help: "Size of repository in bytes",
	}, []string{"repository"})
	rawSizeTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_raw_size_bytes_total",
		Help: "Size of repository in bytes",
	}, []string{"repository"})
	uncompressedSizeTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_uncompressed_size_bytes_total",
		Help: "Size of repository in bytes",
	}, []string{"repository"})
	blobTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_blob_total",
		Help: "Number of blob",
	}, []string{"repository"})
	fileTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_file_total",
		Help: "Number of files",
	}, []string{"repository"})
	compressionRatio = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_compression_ratio",
		Help: "Compression ratio",
	}, []string{"repository"})
	compressionSpaceSavingTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_compression_space_saving_bytes_total",
		Help: "Compression space saving",
	}, []string{"repository"})
	compressionProgress = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_compression_progress_percent",
		Help: "Compression ",
	}, []string{"repository"})
)

func init() {
	prometheus.MustRegister(snapshotsTotal)
	prometheus.MustRegister(restoreSizeTotal)
	prometheus.MustRegister(rawSizeTotal)
	prometheus.MustRegister(uncompressedSizeTotal)
	prometheus.MustRegister(blobTotal)
	prometheus.MustRegister(fileTotal)
	prometheus.MustRegister(compressionRatio)
	prometheus.MustRegister(compressionSpaceSavingTotal)
	prometheus.MustRegister(compressionProgress)
}

func main() {
	var (
		promPort = flag.Int("prom.port", 9150, "port to expose prometheus metrics")
		interval = flag.Int("interval", 60, "number of seconds between each refresh")
	)
	flag.Parse()

	go func() {
		for {
			resticMetrics()
      time.Sleep(time.Duration(*interval) * time.Second)
		}
	}()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	port := fmt.Sprintf(":%d", *promPort)
	log.Printf("Starting restic exporter on %q/metrics", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("cannot start restic exporter: %s", err)
	}
}

func resticMetrics() {
	restoreDataStats := getRestoreDataStats()
	rawDataStats := getRawDataStats()
	restoreSizeTotal.WithLabelValues("repo").Set(float64(restoreDataStats.TotalSize))
	snapshotsTotal.WithLabelValues("repo").Set(float64(restoreDataStats.SnapshotsCount))
	fileTotal.WithLabelValues("repo").Set(float64(restoreDataStats.TotalFileCount))

	rawSizeTotal.WithLabelValues("repo").Set(float64(rawDataStats.TotalSize))
	uncompressedSizeTotal.WithLabelValues("repo").Set(float64(rawDataStats.TotalUncompressedSize))
	blobTotal.WithLabelValues("repo").Set(float64(rawDataStats.TotalBlobCount))
	compressionRatio.WithLabelValues("repo").Set(float64(rawDataStats.CompressionRatio))
	compressionSpaceSavingTotal.WithLabelValues("repo").Set(float64(rawDataStats.CompressionSpaceSaving))
	compressionProgress.WithLabelValues("repo").Set(float64(rawDataStats.CompressionProgress))
}

func getRestoreDataStats() restoreDataStats {
	var stats restoreDataStats
	err := json.Unmarshal(getStats("restore-size"), &stats)
	if err != nil {
		return restoreDataStats{}
	}
	return stats
}

func getRawDataStats() rawDataStats {
	var stats rawDataStats
	err := json.Unmarshal(getStats("raw-data"), &stats)
	if err != nil {
		return rawDataStats{}
	}

	return stats
}

func getStats(mode string) []byte {
	cmd := exec.Command("restic", "stats", fmt.Sprintf("--mode=%s", mode), "--no-lock", "--json")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, os.Getenv("RESTIC_REPOSITORY"))
	cmd.Env = append(cmd.Env, os.Getenv("RESTIC_PASSWORD"))
	stdout, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println(err.Error())
	}

	return stdout
}
