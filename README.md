# Restic Exporter

[Prometheus](https://prometheus.io/) exporter for [Restic backup system](https://github.com/restic/restic).
It allows you to monitor the health and performance of your Restic backups, by exposing various metrics in a format that can be scraped by Prometheus.

## Usage

### Flags

- `--prom.port` : the port used to expose Prometheus metrics. The default is 9150.
- `--backup.port` : the port used to receive the backup summary. The default is 9151. This feature is enabled by `FT_BACKUP_SUMMARY`
- `--interval` : the number of seconds between each refresh. The default interval is 60 seconds.

### Docker Compose

Add this service in your `docker-compose.yml`

```yaml
version: '3.8'

services:
   restic-exporter:
     image: nidourah/restic-exporter:1.1.0
     container_name: restic-exporter
     restart: unless-stopped
     command:
       - '--interval=3600'
     env_file:
       - repo-variables.env
```

Create `repo-variables.env` file with :

```env
RESTIC_REPOSITORY=
RESTIC_PASSWORD=
RESTIC_HOST=

FT_BACKUP_SUMMARY=true # Allow to send backup summaries with POST /backups/summaries

# You can add also add your S3 API Key :
# AWS_ACCESS_KEY_ID=
# AWS_SECRET_ACCESS_KEY=
```

## Metrics

```
# HELP restic_check_status Result of restic check operation in the repository
# TYPE restic_check_status gauge
restic_check_status{repository="/my-repo"} 1
# HELP restic_snapshots_latest_timestamp Timestamp of the last backup
# TYPE restic_snapshots_latest_timestamp gauge
restic_snapshots_latest_timestamp{exludes="node_modules",host="Vincents-MacBook-Pro.local",id="c642f301109ddb1a70e114efd51b28bbdb9194aded978449ec123d43ff03c9af",path="/backup",repository="/my-repo",username="vincenthardouin"} 1.674584814e+09
# HELP restic_stats_blob_total Number of blob
# TYPE restic_stats_blob_total gauge
restic_stats_blob_total{repository="/my-repo"} 11337
# HELP restic_stats_compression_progress_percent Compression progression
# TYPE restic_stats_compression_progress_percent gauge
restic_stats_compression_progress_percent{repository="/my-repo"} 100
# HELP restic_stats_compression_ratio Compression ratio
# TYPE restic_stats_compression_ratio gauge
restic_stats_compression_ratio{repository="/my-repo"} 1.5320921953896214
# HELP restic_stats_compression_space_saving_bytes_total Compression space saving
# TYPE restic_stats_compression_space_saving_bytes_total gauge
restic_stats_compression_space_saving_bytes_total{repository="/my-repo"} 34.72977651023845
# HELP restic_stats_file_total Number of files
# TYPE restic_stats_file_total gauge
restic_stats_file_total{repository="/my-repo"} 33839
# HELP restic_stats_raw_size_bytes_total Size of repository in bytes
# TYPE restic_stats_raw_size_bytes_total gauge
restic_stats_raw_size_bytes_total{repository="/my-repo"} 4.42548767e+08
# HELP restic_stats_restore_size_bytes_total Size of repository in bytes
# TYPE restic_stats_restore_size_bytes_total gauge
restic_stats_restore_size_bytes_total{repository="/my-repo"} 2.056417114e+09
# HELP restic_stats_snapshot_total Number of snapshots
# TYPE restic_stats_snapshot_total gauge
restic_stats_snapshot_total{repository="/my-repo"} 3
# HELP restic_stats_uncompressed_size_bytes_total Size of repository in bytes
# TYPE restic_stats_uncompressed_size_bytes_total gauge
restic_stats_uncompressed_size_bytes_total{repository="/my-repo"} 6.78025512e+08

# If FT_BACKUP_SUMMARY is enable

# HELP restic_backup_data_added Amount of data added.
# TYPE restic_backup_data_added gauge
restic_backup_data_added{repository="/my-repo",snapshot_id="2394d34"} 17599
# HELP restic_backup_data_blobs Number of data blobs.
# TYPE restic_backup_data_blobs gauge
restic_backup_data_blobs{repository="/my-repo",snapshot_id="2394d34ff79a06cb"} 0
# HELP restic_backup_dirs_changed Number of changed directories.
# TYPE restic_backup_dirs_changed gauge
restic_backup_dirs_changed{repository="/my-repo",snapshot_id="2394d34ff79a06cb"} 7
# HELP restic_backup_dirs_new Number of new directories.
# TYPE restic_backup_dirs_new gauge
restic_backup_dirs_new{repository="/my-repo",snapshot_id="2394d34ff79a06cb"} 0
# HELP restic_backup_dirs_unmodified Number of unmodified directories.
# TYPE restic_backup_dirs_unmodified gauge
restic_backup_dirs_unmodified{repository="/my-repo",snapshot_id="2394d34ff79a06cb"} 1717
# HELP restic_backup_files_changed Number of changed files.
# TYPE restic_backup_files_changed gauge
restic_backup_files_changed{repository="/my-repo",snapshot_id="2394d34ff79a06cb"} 0
# HELP restic_backup_files_new Number of new files.
# TYPE restic_backup_files_new gauge
restic_backup_files_new{repository="/my-repo",snapshot_id="2394d34ff79a06cb"} 0
# HELP restic_backup_files_unmodified Number of unmodified files.
# TYPE restic_backup_files_unmodified gauge
restic_backup_files_unmodified{repository="/my-repo",snapshot_id="2394d34ff79a06cb"} 11277
# HELP restic_backup_total_bytes_processed Total number of processed bytes.
# TYPE restic_backup_total_bytes_processed gauge
restic_backup_total_bytes_processed{repository="/my-repo",snapshot_id="2394d34ff79a06cb"} 7.02664621e+08
# HELP restic_backup_total_duration Total duration of processing.
# TYPE restic_backup_total_duration gauge
restic_backup_total_duration{repository="/my-repo",snapshot_id="2394d34ff79a06cb"} 0.591441166
# HELP restic_backup_total_files_processed Total number of processed files.
# TYPE restic_backup_total_files_processed gauge
restic_backup_total_files_processed{repository="/my-repo",snapshot_id="2394d34ff79a06cb"} 11277
# HELP restic_backup_tree_blobs Number of tree blobs.
# TYPE restic_backup_tree_blobs gauge
restic_backup_tree_blobs{repository="/my-repo",snapshot_id="2394d34ff79a06cb"} 6

```