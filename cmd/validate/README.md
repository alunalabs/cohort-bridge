# CohortBridge Validate Tool

The `validate` tool performs end-to-end validation of the privacy-preserving record linkage (PPRL) pipeline. It compares matching results against ground truth data to evaluate accuracy, precision, recall, and F1-score.

## Features

- **Interactive Mode**: Guided setup with prompts for all parameters
- **Command Line Mode**: Batch processing with arguments
- **Comprehensive Validation**: Precision, recall, F1-score, and detailed analysis
- **Multiple Data Formats**: Support for both raw and tokenized datasets
- **Noise Injection**: Optional random bits for robustness testing
- **Detailed Reporting**: CSV output with false positives, false negatives, and statistics

## Usage

### Interactive Mode (Recommended)

Simply run the tool without arguments to enter interactive mode:

```bash
./validate
```

The tool will guide you through:

1. **Configuration Files**: Select two config files for datasets A and B
2. **Ground Truth File**: Choose the CSV file containing expected matches
3. **Validation Parameters**: Set thresholds and noise levels
4. **Output Configuration**: Specify where to save results

### Command Line Mode

For batch processing or automation:

```bash
./validate <config1> <config2> <ground_truth> [output_file] [candidate_threshold] [hamming_threshold] [random_bits_percent]
```

**Arguments:**
- `config1`: Configuration file for dataset A (required)
- `config2`: Configuration file for dataset B (required) 
- `ground_truth`: CSV file with expected matches (required)
- `output_file`: Results output path (default: `out/validation_results.csv`)
- `candidate_threshold`: Minimum similarity score (0.0-1.0, default: 0.95)
- `hamming_threshold`: Maximum Hamming distance for matches (default: 100)
- `random_bits_percent`: Noise injection percentage (0.0-1.0, default: 0.0)

### Examples

**Interactive mode:**
```bash
./validate
```

**Command line examples:**
```bash
# Basic validation
./validate config_a.yaml config_b.yaml data/expected_matches.csv

# Custom thresholds
./validate config_a.yaml config_b.yaml ground_truth.csv results.csv 0.90 80 0.05

# Robustness testing with 5% random noise
./validate config_a.yaml config_b.yaml ground_truth.csv results.csv 0.95 100 0.05
```

## Input Files

### Configuration Files

Standard CohortBridge YAML configuration files. Can point to:
- Raw CSV datasets (will be tokenized during validation)
- Pre-tokenized datasets (loads existing tokens)

Example configuration:
```yaml
database:
  type: csv
  filename: "data/dataset_a.csv"
  is_tokenized: false
  fields:
    - first_name
    - last_name
    - date_of_birth
```

### Ground Truth File

CSV file with expected matches. Format:
```csv
id1,id2
patient_001,patient_102
patient_003,patient_098
patient_005,patient_067
```

Where `id1` and `id2` are record IDs that should match between the two datasets.

## Validation Parameters

### Candidate Threshold (0.0-1.0)
- Minimum similarity score for a record pair to be considered a potential match
- Higher values = stricter matching, fewer candidates
- Default: 0.95

### Hamming Threshold (integer)
- Maximum Hamming distance allowed between Bloom filters for a match
- Lower values = stricter matching, more precise
- Default: 100

### Jaccard Threshold (0.0-1.0)
- Minimum Jaccard similarity for MinHash comparison
- Higher values = stricter matching
- Default: 0.8

### Random Bits Percentage (0.0-1.0)
- Percentage of random bits added to Bloom filters for robustness testing
- Simulates real-world noise and variations
- 0.0 = no noise, 1.0 = maximum noise
- Default: 0.0

## Output

### Console Output

During validation, the tool displays:
- Dataset loading progress
- Matching pipeline execution
- Real-time validation statistics
- Performance metrics summary

Example output:
```
📊 Loaded 150 ground truth matches
📂 Loading datasets...
✅ Dataset 1: 1000 records
✅ Dataset 2: 1200 records

🔄 Running matching pipeline...
📈 Validating results...

📈 Validation Results:
===================
True Positives: 142
False Positives: 8
False Negatives: 8
Precision: 94.67%
Recall: 94.67%
F1-Score: 94.67%
```

### CSV Results File

Detailed results saved to specified output file with columns:
- `match_type`: TP (true positive), FP (false positive), FN (false negative)
- `id1`, `id2`: Record identifiers
- `predicted_match`: Whether algorithm predicted a match
- `ground_truth_match`: Whether this was an expected match
- `similarity_score`: Calculated similarity score
- `hamming_distance`: Bloom filter Hamming distance
- `jaccard_similarity`: MinHash Jaccard similarity

## Best Practices

### Dataset Preparation
1. Ensure both datasets have consistent field naming
2. Clean and standardize data before validation
3. Create comprehensive ground truth with diverse match types

### Parameter Tuning
1. Start with default parameters
2. Adjust candidate threshold based on dataset quality
3. Use random bits testing to evaluate robustness
4. Compare results across different threshold combinations

### Performance Analysis
1. Examine false positives to identify over-matching patterns
2. Review false negatives to find missed matches
3. Analyze lowest/highest scoring pairs for threshold optimization
4. Use validation results to tune production matching parameters

## Integration

The validate tool integrates with the CohortBridge ecosystem:

```bash
# Complete workflow with validation
./tokenize -config config_a.yaml -output tokens_a.csv
./tokenize -config config_b.yaml -output tokens_b.csv
./intersect -dataset1 tokens_a.csv -dataset2 tokens_b.csv -output intersection.csv
./validate config_a.yaml config_b.yaml ground_truth.csv validation_results.csv
```

## Error Handling

Common issues and solutions:

**File not found errors:**
- Verify all file paths are correct
- Check that ground truth CSV exists and is readable

**Configuration errors:**
- Ensure YAML files are valid
- Verify dataset paths in config files exist

**Memory issues with large datasets:**
- Use streaming mode in configurations
- Consider reducing batch sizes
- Monitor system resources during validation

## Advanced Usage

### Automated Testing
```bash
#!/bin/bash
# Validation script for multiple parameter combinations
for threshold in 0.90 0.95 0.98; do
    for hamming in 50 100 150; do
        ./validate config_a.yaml config_b.yaml ground_truth.csv \
            "results_${threshold}_${hamming}.csv" \
            $threshold $hamming 0.0
    done
done
```

### Performance Benchmarking
```bash
# Test with increasing noise levels
for noise in 0.0 0.01 0.05 0.10; do
    echo "Testing with ${noise} noise..."
    ./validate config_a.yaml config_b.yaml ground_truth.csv \
        "results_noise_${noise}.csv" 0.95 100 $noise
done
``` 