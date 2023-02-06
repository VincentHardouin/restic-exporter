package main

import (
	"encoding/json"
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
	filesNew = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_backup_files_new",
		Help: "Number of new files.",
	}, []string{"repository", "snapshot_id"})
	filesChanged = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_backup_files_changed",
		Help: "Number of changed files.",
	}, []string{"repository", "snapshot_id"})
	filesUnmodified = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_backup_files_unmodified",
		Help: "Number of unmodified files.",
	}, []string{"repository", "snapshot_id"})
	dirsNew = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_backup_dirs_new",
		Help: "Number of new directories.",
	}, []string{"repository", "snapshot_id"})
	dirsChanged = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_backup_dirs_changed",
		Help: "Number of changed directories.",
	}, []string{"repository", "snapshot_id"})
	dirsUnmodified = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_backup_dirs_unmodified",
		Help: "Number of unmodified directories.",
	}, []string{"repository", "snapshot_id"})
	dataBlobs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_backup_data_blobs",
		Help: "Number of data blobs.",
	}, []string{"repository", "snapshot_id"})
	treeBlobs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_backup_tree_blobs",
		Help: "Number of tree blobs.",
	}, []string{"repository", "snapshot_id"})
	dataAdded = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_backup_data_added",
		Help: "Amount of data added.",
	}, []string{"repository", "snapshot_id"})
	totalFilesProcessed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_backup_total_files_processed",
		Help: "Total number of processed files.",
	}, []string{"repository", "snapshot_id"})
	totalBytesProcessed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_backup_total_bytes_processed",
		Help: "Total number of processed bytes.",
	}, []string{"repository", "snapshot_id"})
	totalDuration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "restic_backup_total_duration",
		Help: "Total duration of processing.",
	}, []string{"repository", "snapshot_id"})
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

	config = getConfig()

	if config.FeatureToggles.BackupSummary {
		prometheus.MustRegister(filesNew)
		prometheus.MustRegister(filesChanged)
		prometheus.MustRegister(filesUnmodified)
		prometheus.MustRegister(dirsNew)
		prometheus.MustRegister(dirsChanged)
		prometheus.MustRegister(dirsUnmodified)
		prometheus.MustRegister(dataBlobs)
		prometheus.MustRegister(treeBlobs)
		prometheus.MustRegister(dataAdded)
		prometheus.MustRegister(totalFilesProcessed)
		prometheus.MustRegister(totalBytesProcessed)
		prometheus.MustRegister(totalDuration)
	}
}

func main() {
	var (
		promPort   = flag.Int("prom.port", 9150, "port to expose prometheus metrics")
		backupPort = flag.Int("backup.port", 9151, "port to receive backup summary")
		interval   = flag.Int("interval", 60, "number of seconds between each refresh")
	)
	flag.Parse()

	log.Printf("Interval: %d", *interval)
	s := gocron.NewScheduler(time.UTC)
	_, err := s.Every(*interval).Seconds().Do(updateResticMetrics)
	if err != nil {
		log.Fatalf("Error scheduling job")
	}
	s.StartAsync()

	log.Printf("FT_BACKUP_SUMMARY: %t", config.FeatureToggles.BackupSummary)

	if config.FeatureToggles.BackupSummary {

		muxBackup := http.NewServeMux()
		save := &saveBackupSummaryHandler{}
		muxBackup.Handle("/backups/summaries", save)

		port2 := fmt.Sprintf(":%d", *backupPort)

		go func() {
			startServer("backup summary", port2, muxBackup)
		}()
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	port := fmt.Sprintf(":%d", *promPort)

	startServer("restic exporter", port, mux)
}

func startServer(serverName string, port string, mux *http.ServeMux) {
	log.Printf("Stating %s %q", serverName, port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Cannot start %s: %s", serverName, err)
	}
}

func updateResticMetrics() {
	updateStatisticsMetrics(config.Restic)
	updateSnapshotsMetrics(config.Restic)
	updateCheckStatus(config.Restic)
	if config.FeatureToggles.BackupSummary {
		updateBackupSummary(config.Restic)
	}
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

func updateBackupSummary(restic ResticConfig) {
	summary := getBackupSummary()

	filesNew.WithLabelValues(restic.Repository, summary.SnapshotID).Set(float64(summary.FilesNew))
	filesChanged.WithLabelValues(restic.Repository, summary.SnapshotID).Set(float64(summary.FilesChanged))
	filesUnmodified.WithLabelValues(restic.Repository, summary.SnapshotID).Set(float64(summary.FilesUnmodified))
	dirsNew.WithLabelValues(restic.Repository, summary.SnapshotID).Set(float64(summary.DirsNew))
	dirsChanged.WithLabelValues(restic.Repository, summary.SnapshotID).Set(float64(summary.DirsChanged))
	dirsUnmodified.WithLabelValues(restic.Repository, summary.SnapshotID).Set(float64(summary.DirsUnmodified))
	dataBlobs.WithLabelValues(restic.Repository, summary.SnapshotID).Set(float64(summary.DataBlobs))
	treeBlobs.WithLabelValues(restic.Repository, summary.SnapshotID).Set(float64(summary.TreeBlobs))
	dataAdded.WithLabelValues(restic.Repository, summary.SnapshotID).Set(float64(summary.DataAdded))
	totalFilesProcessed.WithLabelValues(restic.Repository, summary.SnapshotID).Set(float64(summary.TotalFilesProcessed))
	totalBytesProcessed.WithLabelValues(restic.Repository, summary.SnapshotID).Set(float64(summary.TotalBytesProcessed))
	totalDuration.WithLabelValues(restic.Repository, summary.SnapshotID).Set(float64(summary.TotalDuration))
}

type saveBackupSummaryHandler struct {
}

func (s *saveBackupSummaryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", 405)
	}
	var summary backupSummary

	err := json.NewDecoder(r.Body).Decode(&summary)
	if err != nil {
		log.Fatalln("There was an error decoding the request body into the struct")
	}

	saveBackupSummary(summary)

	w.Write([]byte("Backup summary saved"))
}
