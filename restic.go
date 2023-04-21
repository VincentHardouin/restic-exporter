package main

import (
	"encoding/json"
	"fmt"
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

type snapshot struct {
	Time     time.Time `json:"time"`
	Paths    []string  `json:"paths"`
	Hostname string    `json:"hostname"`
	Username string    `json:"username"`
	Excludes []string  `json:"excludes"`
	ID       string    `json:"id"`
}

type backupSummary struct {
	FilesNew            int     `json:"files_new"`
	FilesChanged        int     `json:"files_changed"`
	FilesUnmodified     int     `json:"files_unmodified"`
	DirsNew             int     `json:"dirs_new"`
	DirsChanged         int     `json:"dirs_changed"`
	DirsUnmodified      int     `json:"dirs_unmodified"`
	DataBlobs           int     `json:"data_blobs"`
	TreeBlobs           int     `json:"tree_blobs"`
	DataAdded           int     `json:"data_added"`
	TotalFilesProcessed int     `json:"total_files_processed"`
	TotalBytesProcessed int     `json:"total_bytes_processed"`
	TotalDuration       float64 `json:"total_duration"`
	SnapshotID          string  `json:"snapshot_id"`
}

var summary backupSummary

func getRestoreDataStats(restic ResticConfig) restoreDataStats {
	var stats restoreDataStats
	err := json.Unmarshal(getStats("restore-size", restic), &stats)
	if err != nil {
		return restoreDataStats{}
	}
	return stats
}

func getRawDataStats(restic ResticConfig) rawDataStats {
	var stats rawDataStats
	err := json.Unmarshal(getStats("raw-data", restic), &stats)
	if err != nil {
		return rawDataStats{}
	}

	return stats
}

func getStats(mode string, restic ResticConfig) []byte {
	cmd := exec.Command("restic", "stats", fmt.Sprintf("--mode=%s", mode), "--no-lock", "--json")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, restic.Repository)
	cmd.Env = append(cmd.Env, restic.Password)
	stdout, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println(err.Error())
	}

	return stdout
}

func getLatestSnapshotInformation(restic ResticConfig) snapshot {
	cmd := exec.Command("restic", "snapshots", "--latest=1", "--no-lock", "--json")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, restic.Repository)
	cmd.Env = append(cmd.Env, restic.Password)
	stdout, errCmd := cmd.CombinedOutput()

	if errCmd != nil {
		fmt.Println(errCmd.Error())
	}

	var snapshotInformation []snapshot
	err := json.Unmarshal(stdout, &snapshotInformation)
	if err != nil {
		fmt.Println(err.Error())
		return snapshot{}
	}
	return snapshotInformation[0]
}

func getCheckStatus(restic ResticConfig) int {
	cmd := exec.Command("restic", "check", "--no-lock")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, restic.Repository)
	cmd.Env = append(cmd.Env, restic.Password)
	err := cmd.Run()
	if err != nil {
		return 0
	}
	return 1
}

func saveBackupSummary(newSummary backupSummary) {
	summary = newSummary
}

func getBackupSummary() backupSummary {
	return summary
}
