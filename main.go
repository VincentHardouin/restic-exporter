package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	snapshotsTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_stats_snapshot_total",
		Help: "Number of snapshots",
	}, []string{"repository"})
	restoreSizeTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_stats_restore_size_bytes_total",
		Help: "Size of repository in bytes",
	}, []string{"repository"})
	rawSizeTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_stats_raw_size_bytes_total",
		Help: "Size of repository in bytes",
	}, []string{"repository"})
	uncompressedSizeTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_stats_uncompressed_size_bytes_total",
		Help: "Size of repository in bytes",
	}, []string{"repository"})
	blobTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_stats_blob_total",
		Help: "Number of blob",
	}, []string{"repository"})
	fileTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_stats_file_total",
		Help: "Number of files",
	}, []string{"repository"})
	compressionRatio = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_stats_compression_ratio",
		Help: "Compression ratio",
	}, []string{"repository"})
	compressionSpaceSavingTotal = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_stats_compression_space_saving_bytes_total",
		Help: "Compression space saving",
	}, []string{"repository"})
	compressionProgress = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_stats_compression_progress_percent",
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

	s := gocron.NewScheduler(time.UTC)
	_, err := s.Every(*interval).Seconds().Do(updateResticMetrics)
	if err != nil {
		log.Fatalf("Error scheduling job")
	}
	s.StartAsync()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	port := fmt.Sprintf(":%d", *promPort)
	log.Printf("Starting restic exporter on %q/metrics", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Cannot start restic exporter: %s", err)
	}
}

func updateResticMetrics() {
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
