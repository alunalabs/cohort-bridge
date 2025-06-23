# Validate Tool

## Overview

The `validate` tool is an end-to-end validation script for the HIPAA-compliant fuzzy matching system. It runs the complete matching pipeline and validates the results against ground truth data to measure the accuracy and performance of the matching algorithm.

## Purpose

This tool helps you:
- **Evaluate matching accuracy** by comparing system results with known correct matches
- **Measure performance metrics** like precision, recall, and F1-score
- **Optimize parameters** by testing different thresholds and configurations
- **Analyze match quality** by examining score distributions and identifying problematic cases

## Usage

```bash
./validate <config1> <config2> <ground_truth> [output_file] [candidate_threshold] [hamming_threshold] [random_bits_percent]
```

### Required Arguments

1. **`config1`** - Configuration file for the first dataset (Party A)
2. **`config2`** - Configuration file for the second dataset (Party B) 
3. **`ground_truth`** - CSV file containing known correct matches

### Optional Arguments

4. **`output_file`** - Path to save validation results (default: `out/validation_results.csv`)
5. **`candidate_threshold`** - Minimum similarity score to consider a pair as a candidate (default: `0.95`)
6. **`hamming_threshold`** - Maximum Hamming distance for a positive match (default: `100`)
7. **`random_bits_percent`** - Percentage of random bits to add to Bloom filters, 0.0-1.0 (default: `0.0`)

## Example Usage

### Basic Validation
```bash
./validate config_a.yaml config_b.yaml data/expected_matches.csv
```

### Advanced Validation with Custom Parameters
```bash
./validate config_a.yaml config_b.yaml data/expected_matches.csv validation_results.csv 0.90 80 0.05
```

This example:
- Uses custom output file `validation_results.csv`
- Sets candidate threshold to 0.90 (90% similarity)
- Sets Hamming threshold to 80
- Adds 5% random bits to Bloom filters

## Input File Formats

### Configuration Files
The tool supports both raw data and tokenized data configurations. Your config files should specify:
- Database connection details or file paths
- Field mappings for patient records
- Whether data is tokenized (`IsTokenized: true/false`)

### Ground Truth File
A CSV file with columns:
- `patient_a_id` - Patient ID from the first dataset
- `patient_b_id` - Corresponding patient ID from the second dataset

Example:
```csv
patient_a_id,patient_b_id
P001,Q001
P002,Q002
P005,Q007
```

## Output

### Console Output
The tool provides real-time feedback including:
- Dataset loading progress
- Matching pipeline execution
- Validation metrics and quality assessment
- Examples of false positives and missed matches

### Validation Results File
A detailed CSV file containing:
- **Summary metrics**: True positives, false positives, false negatives, precision, recall, F1-score
- **Score analysis**: Lowest ground truth score, highest non-ground truth score
- **Detailed results**: All matches categorized as true positives, false positives, or false negatives with scores

## Understanding the Metrics

### Core Metrics
- **Precision**: Percentage of predicted matches that are actually correct
- **Recall**: Percentage of actual matches that were correctly identified
- **F1-Score**: Harmonic mean of precision and recall (overall quality measure)

### Quality Thresholds
- **Excellent**: ≥ 0.9
- **Good**: 0.8 - 0.89
- **Fair**: 0.7 - 0.79
- **Poor**: < 0.7

### Score Analysis
The tool analyzes score distributions to identify:
- **Score overlap**: When non-matches have higher scores than true matches (problematic)
- **Clear separation**: When all true matches score higher than non-matches (ideal)

## Parameter Tuning

### Candidate Threshold
- **Higher values (0.95-0.99)**: Fewer candidates, higher precision, potentially lower recall
- **Lower values (0.85-0.94)**: More candidates, potentially higher recall, more false positives

### Hamming Threshold
- **Lower values (50-80)**: Stricter matching, higher precision, potentially lower recall
- **Higher values (100-150)**: More lenient matching, potentially higher recall, more false positives

### Random Bits Percentage
- **0.0**: No noise, baseline performance
- **0.01-0.05**: Light noise simulation
- **0.05-0.10**: Moderate noise simulation
- **>0.10**: Heavy noise simulation

## Prerequisites

Before running the validation:

1. **Prepare your datasets** in the format specified by your configuration files
2. **Create ground truth data** with known correct matches
3. **Configure your YAML files** with proper database and field settings
4. **Ensure output directory exists** (the tool will create `out/` if needed)

## Troubleshooting

### Common Issues

**"config file does not exist"**
- Verify the config file paths are correct
- Check that files have proper YAML syntax

**"ground truth file does not exist"**
- Ensure the ground truth CSV file exists
- Verify the file has the required columns: `patient_a_id`, `patient_b_id`

**"Failed to load dataset"**
- Check that data files referenced in configs exist
- Verify CSV format and column headers match configuration
- For tokenized data, ensure tokenized files are properly formatted

**Low precision/recall**
- Try adjusting the candidate threshold
- Modify the Hamming threshold
- Check if your ground truth data is complete and accurate
- Consider whether your datasets have sufficient quality for matching

## Advanced Usage

### Batch Testing
You can create scripts to test multiple parameter combinations:

```bash
#!/bin/bash
for threshold in 0.90 0.92 0.95 0.97; do
    for hamming in 60 80 100 120; do
        echo "Testing threshold=$threshold, hamming=$hamming"
        ./validate config_a.yaml config_b.yaml ground_truth.csv \
            "results_${threshold}_${hamming}.csv" $threshold $hamming 0.0
    done
done
```

### Performance Analysis
Use the random bits parameter to simulate real-world noise:
```bash
# Test with increasing noise levels
./validate config_a.yaml config_b.yaml ground_truth.csv results_noise_0.csv 0.95 100 0.0
./validate config_a.yaml config_b.yaml ground_truth.csv results_noise_2.csv 0.95 100 0.02
./validate config_a.yaml config_b.yaml ground_truth.csv results_noise_5.csv 0.95 100 0.05
```

This helps you understand how robust your matching algorithm is to data quality issues. 