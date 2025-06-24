# CohortBridge Data Sender

Send intersection results or matched data to another CohortBridge receiver.

## Overview

The `send` program transmits intersection results or matched raw data to a remote CohortBridge receiver. It supports two modes: sending just the intersection results (ID pairs) or sending the actual data records for matched pairs.

## Usage

```bash
./send [OPTIONS]
```

## Required Arguments

- `-intersection` - Path to intersection results file (required)

## Optional Arguments

- `-host` - Target host to send data to
- `-port` - Target port to send data to
- `-config` - Configuration file for network settings (default: config.yaml)
- `-data` - Optional raw data file to send matched records
- `-mode` - Send mode: 'intersection' or 'matched_data' (default: intersection)
- `-help` - Show help message

## Modes

### Intersection Mode
Sends only the intersection results (ID pairs) to the receiver for verification.

```bash
./send -intersection=intersection_results.csv -host=peer.example.com -port=8080
```

### Matched Data Mode
Sends actual data records for matched pairs. Requires the `-data` parameter.

```bash
./send -intersection=results.csv -data=raw_data.json \
  -mode=matched_data -config=sender_config.yaml
```

## Examples

### Basic Intersection Send
```bash
# Send intersection results using command line parameters
./send -intersection=intersection_results.csv \
  -host=192.168.1.100 -port=8080
```

### Send with Configuration File
```bash
# Use configuration file for network settings
./send -intersection=results.csv -config=sender_config.yaml
```

### Send Matched Data
```bash
# Send actual matched records (requires data file)
./send -intersection=results.csv -data=patient_data.json \
  -mode=matched_data -config=config.yaml
```

## Input Formats

### Intersection Results
CSV file produced by the `intersect` program:

```csv
id1,id2,is_match,hamming_distance,jaccard_similarity,match_score,timestamp
record_1,record_2,true,45,0.85,0.92,2024-01-01T00:00:00Z
record_3,record_4,false,120,0.25,0.15,2024-01-01T00:00:00Z
```

### Raw Data (for matched_data mode)
JSON file with raw records:

```json
[
  {
    "id": "record_1",
    "name": "John Doe",
    "dob": "1990-01-01",
    "email": "john@example.com"
  }
]
```

## Network Configuration

The sender can get network settings from:

1. Command line parameters (`-host`, `-port`)
2. Configuration file (`-config`)
3. Default config.yaml file

Priority: Command line > Config file > Defaults

Example config.yaml:
```yaml
peer:
  host: "receiver.example.com"
  port: 8080
```

## Security

- All data is transmitted over TCP connections
- The receiver must be running and listening for connections
- Network timeouts and error handling are built-in
- Only matched records are sent in matched_data mode (privacy-preserving)

## Integration

This program is designed to be called by the `agent` orchestrator as part of the complete PPRL workflow, or used standalone for custom data transmission scenarios. 