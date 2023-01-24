package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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
