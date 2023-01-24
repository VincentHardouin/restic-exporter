package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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
			getResticMetrics()
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

func getResticMetrics() {
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