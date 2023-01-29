# Restic Exporter

[Prometheus](https://prometheus.io/) exporter for [Restic backup system](https://github.com/restic/restic).

## Usage

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
```