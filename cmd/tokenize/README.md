# Tokenization Tool

The tokenization tool converts raw PHI (Protected Health Information) data into privacy-preserving Bloom filter tokens for secure record linkage.

## 🔧 Building the Tool

```bash
# Build the tokenization tool
go build -o tokenize.exe ./cmd/tokenize/

# Or use the Makefile
make tokenize
```

## 📋 Overview

This tool takes sensitive patient data and converts it into:
- **Bloom filters**: Privacy-preserving bit arrays that represent record features
- **MinHash signatures**: Compact representations for similarity estimation
- **Base64-encoded tokens**: Secure format for transmission between parties

## 🚀 Usage

### Command Line Mode

```bash
# Basic CSV tokenization
./tokenize.exe -input data/patients.csv -output out/tokens.json

# Specify input/output formats explicitly
./tokenize.exe -input data.csv -output tokens.json -input-format csv -output-format json

# Use custom batch size for large datasets
./tokenize.exe -input large_dataset.csv -output tokens.json -batch-size 5000

# Use different field configuration
./tokenize.exe -input data.csv -output tokens.json -main-config custom_config.yaml
```

### Database Mode

```bash
# Tokenize directly from PostgreSQL database
./tokenize.exe -database -main-config postgres_a.yaml -output tokens.json

# Specify output format for database mode
./tokenize.exe -database -main-config postgres_a.yaml -output tokens.csv -output-format csv
```

### Interactive Mode

```bash
# Interactive configuration
./tokenize.exe -interactive

# Interactive with database
./tokenize.exe -interactive -database
```

### Configuration File Mode

```bash
# Use pre-configured settings
./tokenize.exe -config tokenization_config.yaml
```

## 📁 Input Formats

### CSV Format
```csv
id,FIRST,LAST,BIRTHDATE,ZIP
1,John,Smith,1990-01-01,12345
2,Jane,Doe,1985-05-15,67890
```

### JSON Format
```json
{"id": "1", "FIRST": "John", "LAST": "Smith", "BIRTHDATE": "1990-01-01", "ZIP": "12345"}
{"id": "2", "FIRST": "Jane", "LAST": "Doe", "BIRTHDATE": "1985-05-15", "ZIP": "67890"}
```

### Database
Configure PostgreSQL connection in your config file:
```yaml
database:
  type: "postgres"
  host: "localhost"
  port: 5432
  database: "patients"
  table: "patient_records"
  fields: ["FIRST", "LAST", "BIRTHDATE", "ZIP"]
```

## 📤 Output Formats

### JSON Output (Default)
```json
[
  {
    "id": "1",
    "bloom_filter": "eJwLd3EN8bB0...",
    "minhash": "AQIDBAUGBwgJ...",
    "timestamp": "2024-06-23T20:30:00Z"
  }
]
```

### CSV Output
```csv
id,bloom_filter,minhash,timestamp
1,eJwLd3EN8bB0...,AQIDBAUGBwgJ...,2024-06-23T20:30:00Z
2,fKxMd4FO9cE1...,CgsMDQ4PEBESEw...,2024-06-23T20:30:00Z
```

## ⚙️ Configuration

### Configuration File Example

Create a `tokenization_config.yaml`:

```yaml
input_file: "data/patients.csv"
output_file: "out/tokens.json"
input_format: "csv"
output_format: "json"
batch_size: 1000
fields: ["FIRST", "LAST", "BIRTHDATE", "ZIP"]
bloom_filter_size: 1000
bloom_hash_count: 5
minhash_signatures: 128
minhash_permutations: 1000
random_bits_percent: 0.0
qgram_length: 2
use_database: false
database_config: ""
```

### Main Config Integration

The tool reads field names and database settings from your main config file (default: `config.yaml`):

```yaml
database:
  fields: ["FIRST", "LAST", "BIRTHDATE", "ZIP"]
  type: "postgres"
  host: "localhost"
  port: 5432
  database: "patients"
  table: "patient_records"
```

## 🔒 Privacy Features

### Bloom Filter Parameters
- **Size**: 1000 bits (configurable)
- **Hash Functions**: 5 (configurable) 
- **Q-grams**: 2-character substrings with padding
- **Random Bits**: Optional noise injection (0% by default)

### Example Q-gram Processing
- `"John"` → `["_J", "Jo", "oh", "hn", "n_"]`
- `"12345"` → `["_1", "12", "23", "34", "45", "5_"]`

### MinHash Signatures
- **Permutations**: 1000 (configurable)
- **Signatures**: 128 (configurable)
- Used for efficient similarity estimation

## 📊 Performance

### Batch Processing
- **Default Batch Size**: 1000 records
- **Memory Efficient**: Streams through large datasets
- **Progress Reporting**: Shows batch progress during processing

### Supported Scale
- ✅ **Small datasets**: < 1,000 records (< 1 second)
- ✅ **Medium datasets**: 1,000 - 100,000 records (< 1 minute)
- ✅ **Large datasets**: 100,000+ records (streaming processing)

## 🔧 Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `-input FILE` | Input file path | - |
| `-output FILE` | Output file path | - |
| `-input-format FORMAT` | Input format: csv, json, postgres | auto-detect |
| `-output-format FORMAT` | Output format: csv, json | auto-detect |
| `-batch-size N` | Records per batch | 1000 |
| `-main-config FILE` | Main config file | config.yaml |
| `-config FILE` | Tokenization config file | - |
| `-database` | Use database instead of file | false |
| `-interactive` | Interactive mode | false |

## 🏥 Healthcare Workflow

### Typical Usage Flow

1. **Data Preparation**
   ```bash
   # Prepare your patient data in CSV format
   # Ensure fields match your config (FIRST, LAST, BIRTHDATE, ZIP)
   ```

2. **Tokenization**
   ```bash
   # Convert PHI to privacy-preserving tokens
   ./tokenize.exe -input patients.csv -output tokens.json
   ```

3. **Secure Exchange**
   ```bash
   # Tokens can be safely shared between organizations
   # No raw PHI is exposed in the token format
   ```

4. **Record Linkage**
   ```bash
   # Use tokens with the matching system
   ./agent.exe -mode=receiver -config=config_a.yaml
   ./agent.exe -mode=sender -config=config_b.yaml
   ```

## 🛠️ Troubleshooting

### Common Issues

**"Failed to read CSV headers"**
- Ensure your CSV file has a header row
- Check file encoding (UTF-8 recommended)

**"Database connection failed"**
- Verify PostgreSQL is running
- Check connection parameters in config file
- Ensure database user has read permissions

**"Out of memory"**
- Reduce batch size: `-batch-size 500`
- Use streaming mode for very large datasets

**"Field not found"**
- Verify field names match between data and config
- Case-sensitive field matching

### Debug Tips

```bash
# Test with small dataset first
head -n 100 large_dataset.csv > test_data.csv
./tokenize.exe -input test_data.csv -output test_tokens.json

# Use smaller batch size for memory issues
./tokenize.exe -input data.csv -output tokens.json -batch-size 100

# Check field names in CSV
head -n 1 data.csv
```

## 📈 Examples

### Example 1: Hospital A Tokenizing Patient Records
```bash
# Hospital A prepares their patient data
./tokenize.exe -input hospital_a_patients.csv -output hospital_a_tokens.json

# Results in privacy-preserving tokens ready for matching
# Output: hospital_a_tokens.json (1,234 tokenized records)
```

### Example 2: Large Dataset with PostgreSQL
```bash
# Research organization with 100K+ records in database
./tokenize.exe -database -main-config research_db.yaml -output large_tokens.json -batch-size 5000

# Processes in batches of 5,000 for memory efficiency
```

### Example 3: Custom Field Configuration
```bash
# Using different fields for matching
./tokenize.exe -interactive
# When prompted:
# Fields: PATIENT_FNAME,PATIENT_LNAME,DOB,POSTAL_CODE
# Input: custom_patients.csv
# Output: custom_tokens.json
```

---

## 🔐 Security Note

This tool is designed for HIPAA-compliant record linkage. The generated tokens contain no raw PHI and use cryptographic techniques (Bloom filters + MinHash) to enable privacy-preserving matching between healthcare organizations. 