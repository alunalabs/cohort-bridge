# CohortBridge Intersection Finder

Find the intersection between two tokenized datasets using privacy-preserving record linkage techniques.

## Overview

The `intersect` program compares two tokenized datasets and identifies matching records based on similarity metrics like Hamming distance and Jaccard similarity. It uses Bloom filters and MinHash signatures to perform secure fuzzy matching without exposing the underlying data.

## Usage

```bash
./intersect [OPTIONS]
```

## Required Arguments

- `-dataset1` - Path to first tokenized dataset file (required)
- `-dataset2` - Path to second tokenized dataset file (required)

## Optional Arguments

- `-output` - Output file for intersection results (default: intersection_results.csv)
- `-config` - Optional configuration file for advanced settings
- `-hamming-threshold` - Maximum Hamming distance for match (default: 100)
- `-jaccard-threshold` - Minimum Jaccard similarity for match (default: 0.5)
- `-batch-size` - Processing batch size for streaming mode (default: 1000)
- `-streaming` - Enable streaming mode for large datasets
- `-help` - Show help message

## Examples

### Basic Intersection
```bash
# Find intersection between two datasets
./intersect -dataset1=tokenized_a.csv -dataset2=tokenized_b.csv
```

### Custom Thresholds
```bash
# Use custom thresholds for more/less strict matching
./intersect -dataset1=data1.csv -dataset2=data2.csv \
  -hamming-threshold=50 -jaccard-threshold=0.8
```

### Streaming Mode for Large Datasets
```bash
# Enable streaming for large datasets to manage memory usage
./intersect -dataset1=large1.csv -dataset2=large2.csv \
  -streaming -batch-size=500
```

## Input Format

The input datasets should be tokenized CSV files produced by the `tokenize` program:

```csv
id,bloom_filter,minhash,timestamp
record_1,base64_encoded_bloom_filter,"[signature_array]",2024-01-01T00:00:00Z
record_2,base64_encoded_bloom_filter,"[signature_array]",2024-01-01T00:00:00Z
```

## Output Format

The intersection results are saved as CSV:

```csv
id1,id2,is_match,hamming_distance,jaccard_similarity,match_score,timestamp
record_1,record_2,true,45,0.85,0.92,2024-01-01T00:00:00Z
record_3,record_4,false,120,0.25,0.15,2024-01-01T00:00:00Z
```

## Performance

- **In-memory mode**: Fastest but requires enough RAM for both datasets
- **Streaming mode**: Memory-efficient for large datasets, processes in batches
- **Thresholds**: Lower Hamming threshold = stricter matching, higher Jaccard threshold = stricter matching

## Integration

This program is designed to be called by the `agent` orchestrator or used standalone for batch processing of tokenized datasets. 