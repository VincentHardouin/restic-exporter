package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
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
		Help: "Compression progression",
	}, []string{"repository"})
	snapshotsLatestTimestamp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_snapshots_latest_timestamp",
		Help: "Timestamp of the last backup",
	}, []string{"repository", "host", "username", "id", "path", "exludes"})
	checkStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_check_status",
		Help: "Result of restic check operation in the repository",
	}, []string{"repository"})
)

var config *Config

func init() {
	godotenv.Load()

	prometheus.MustRegister(snapshotsTotal)
	prometheus.MustRegister(restoreSizeTotal)
	prometheus.MustRegister(rawSizeTotal)
	prometheus.MustRegister(uncompressedSizeTotal)
	prometheus.MustRegister(blobTotal)
	prometheus.MustRegister(fileTotal)
	prometheus.MustRegister(compressionRatio)
	prometheus.MustRegister(compressionSpaceSavingTotal)
	prometheus.MustRegister(compressionProgress)
	prometheus.MustRegister(snapshotsLatestTimestamp)
	prometheus.MustRegister(checkStatus)
}

func main() {
	config = getConfig()
	var (
		promPort = flag.Int("prom.port", 9150, "port to expose prometheus metrics")
		interval = flag.Int("interval", 60, "number of seconds between each refresh")
	)
	flag.Parse()

	log.Printf("Interval: %d", *interval)
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
	updateStatisticsMetrics(config.Restic)
	updateSnapshotsMetrics(config.Restic)
	updateCheckStatus(config.Restic)
}

func updateStatisticsMetrics(restic ResticConfig) {
	restoreDataStats := getRestoreDataStats(restic)
	rawDataStats := getRawDataStats(restic)
	restoreSizeTotal.WithLabelValues(restic.Repository).Set(float64(restoreDataStats.TotalSize))
	snapshotsTotal.WithLabelValues(restic.Repository).Set(float64(restoreDataStats.SnapshotsCount))
	fileTotal.WithLabelValues(restic.Repository).Set(float64(restoreDataStats.TotalFileCount))

	rawSizeTotal.WithLabelValues(restic.Repository).Set(float64(rawDataStats.TotalSize))
	uncompressedSizeTotal.WithLabelValues(restic.Repository).Set(float64(rawDataStats.TotalUncompressedSize))
	blobTotal.WithLabelValues(restic.Repository).Set(float64(rawDataStats.TotalBlobCount))
	compressionRatio.WithLabelValues(restic.Repository).Set(float64(rawDataStats.CompressionRatio))
	compressionSpaceSavingTotal.WithLabelValues(restic.Repository).Set(float64(rawDataStats.CompressionSpaceSaving))
	compressionProgress.WithLabelValues(restic.Repository).Set(float64(rawDataStats.CompressionProgress))
}

func updateSnapshotsMetrics(restic ResticConfig) {
	latestSnapshotInformation := getLatestSnapshotInformation(restic)
	paths := strings.Join(latestSnapshotInformation.Paths, ",")
	excludes := strings.Join(latestSnapshotInformation.Excludes, ",")
	snapshotsLatestTimestamp.WithLabelValues(restic.Repository, latestSnapshotInformation.Hostname, latestSnapshotInformation.Username, latestSnapshotInformation.ID, paths, excludes).Set(float64(latestSnapshotInformation.Time.Unix()))
}

func updateCheckStatus(restic ResticConfig) {
	status := getCheckStatus(restic)
	checkStatus.WithLabelValues(restic.Repository).Set(float64(status))
}
