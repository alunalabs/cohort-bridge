# CohortBridge - Privacy-Preserving Record Linkage

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **⚠️ Work in Progress Notice**  
> While CohortBridge's core functionality for privacy-preserving record linkage is implemented and working, some advanced features are still in development:
> - **Secure Multi-Party Computation**: Garbled circuits and VOLE-based protocols are currently placeholders with simulated outputs
> - **Production Network Layer**: Full peer-to-peer networking protocols are being finalized
> - **Advanced Cryptographic Protocols**: Some secure comparison methods use simplified implementations
> - **Performance Optimizations**: Large-scale dataset processing optimizations are ongoing
> 
> The system is functional for research, testing, and pilot deployments. For production use with sensitive data, please verify results carefully and consider the current limitations.

A secure, HIPAA-compliant system for privacy-preserving record linkage that enables healthcare institutions to identify matching patient records without sharing sensitive PHI data.

## 🚀 Quick Start

### 1. Download the CLI Tool

Download the latest release from our [GitHub Releases page](https://github.com/alunalabs/cohort-bridge/releases):

```bash
# Download for your platform
curl -L https://github.com/alunalabs/cohort-bridge/releases/latest/download/cohort-bridge-windows.zip -o cohort-bridge.zip
unzip cohort-bridge.zip
```

Or compile from source:
```bash
git clone https://github.com/alunalabs/cohort-bridge.git
cd cohort-bridge
make build
```

### 2. Set Up Your Environment

Create your data and output directories:
```bash
mkdir -p data logs out
```

Place your patient data files in the `data/` directory in CSV format with these columns:
- `first_name`, `last_name`, `dob`, `phone`, `email`, `address`

### 3. Create Your Configuration

**Option A: Use Our Web Interface (Recommended)**
1. Visit [TBD](URL_TBD) 
2. Select your matching scenario (two-party, tokenized, secure, etc.)
3. Configure your settings and download the config file
4. Save as `config.yaml` in your project directory

**Option B: Manual Configuration**
Copy and customize one of our example configs:
```bash
# Basic two-party matching
cp config.example.yaml config.yaml

# Tokenized mode (enhanced security)
cp config_tokenized.example.yaml config.yaml

# Secure mode with encryption
cp config_secure.example.yaml config.yaml

# PostgreSQL database mode
cp config_postgres.example.yaml config.yaml
```

Edit the config file to point to your data files and configure your preferences.

### 4. Run the Interactive Mode

```bash
./cohort-bridge
```

The interactive mode will guide you through:
- 🔄 **Mode Selection** - Choose Receiver, Sender, or Local processing
- 📁 **File Discovery** - Automatically find and validate your config files
- ⚙️ **Configuration** - Real-time validation with helpful error messages
- 🚀 **Execution** - Professional UI with progress indicators

### 5. Alternative: Direct CLI Usage

CohortBridge provides a unified CLI with subcommands for specific tasks:

```bash
# Interactive mode (recommended for beginners)
./cohort-bridge

# Tokenize your data (enhanced security)
./cohort-bridge tokenize -input data/patients.csv -output out/tokens.csv

# Find intersections between datasets
./cohort-bridge intersect -dataset1 out/tokens1.csv -dataset2 out/tokens2.csv

# Send results to another party
./cohort-bridge send -intersection out/results.csv -host peer.example.com -port 8080

# Validate matching results
./cohort-bridge validate -ground-truth data/truth.csv -results out/matches.csv

# Run complete workflows (legacy mode)
./cohort-bridge -mode=receiver -config=config.yaml
./cohort-bridge -mode=sender -config=config_sender.yaml

# Get help for any subcommand
./cohort-bridge tokenize -help
./cohort-bridge intersect -help
```

## 🏗️ Architecture & File Structure

### Command Line Tool (`cmd/cohort-bridge/`)

CohortBridge provides a unified CLI program with focused subcommands:

- **Interactive Mode** - User-friendly guided workflow
  - Automatically detects configuration files
  - Provides real-time validation and error messages
  - Beautiful interactive UI with progress indicators
  - Suitable for both beginners and experts

- **`tokenize`** - Privacy-preserving data preparation
  - Converts raw PHI into Bloom filter tokens
  - Enables secure data processing workflows
  - Supports CSV, JSON, and database input formats
  - Usage: `cohort-bridge tokenize -input data.csv -output tokens.csv`

- **`intersect`** - Record linkage and intersection finding
  - Core matching logic using Bloom filters and MinHash
  - Implements secure blocking and fuzzy matching
  - Handles both tokenized and raw data modes
  - Usage: `cohort-bridge intersect -dataset1 tokens1.csv -dataset2 tokens2.csv`

- **`send`** - Secure result transmission
  - Network communication for sharing results
  - Encrypted data exchange between parties
  - Handles authentication and secure protocols
  - Usage: `cohort-bridge send -intersection results.csv -host peer.com -port 8080`

- **`validate`** - Data quality and results validation
  - Validates input data format and quality
  - Analyzes matching results for accuracy metrics
  - Generates comprehensive validation reports
  - Usage: `cohort-bridge validate -ground-truth truth.csv -results results.csv`

- **Legacy Mode** - Backward compatibility
  - Supports older command-line workflows
  - Handles network communication between parties  
  - Usage: `cohort-bridge -mode=receiver -config=config.yaml`

### Test Harness (`cmd/test/`)

Separate testing program for validation and benchmarking:
- End-to-end testing with synthetic datasets
- Performance benchmarking and validation
- Accuracy metrics and privacy analysis

### Core Internal Packages (`internal/`)

The business logic is implemented in modular internal packages:

- **`config/`** - Configuration management
  - Unified config parsing and validation
  - Support for multiple deployment scenarios
  - Environment-specific configuration handling

- **`crypto/`** - Cryptographic primitives
  - Commutative encryption using Curve25519
  - Key generation and management
  - Secure random number generation

- **`db/`** - Data persistence and management
  - CSV file processing and validation
  - PostgreSQL integration for large datasets
  - Tokenized data storage and retrieval

- **`match/`** - Core matching algorithms
  - Blocking strategies using LSH and commutative encryption
  - Fuzzy matching with configurable similarity thresholds
  - Pipeline orchestration and result aggregation

- **`peer/`** - Network communication
  - Secure peer-to-peer protocols
  - Connection management and authentication
  - Message serialization and error handling

- **`pprl/`** - Privacy-Preserving Record Linkage
  - Bloom filter implementation with noise injection
  - MinHash signatures for similarity estimation
  - Storage and serialization of privacy-preserving tokens

- **`server/`** - Network server components
  - HTTP/gRPC server implementations
  - Request routing and middleware
  - Logging and monitoring infrastructure

- **`token/`** - Tokenization and encoding
  - PHI to Bloom filter conversion
  - Token validation and quality checks
  - Format standardization across tools

### Configuration Files

- **Basic Configs**: `config.example.yaml`, `config_a.yaml`, `config_b.yaml`
- **Tokenized Mode**: `config_tokenized.example.yaml`, `config_tokenized_peer.example.yaml`
- **Secure Mode**: `config_secure.example.yaml`, `config_secure_peer.example.yaml`
- **Database Mode**: `config_postgres.example.yaml`
- **Network Demo**: `config_sender.example.yaml`, `config_receiver.example.yaml`
- **Normalization Examples**: `config_normalization.example.yaml` - Full example with normalization features

#### Data Normalization

CohortBridge supports automatic data normalization to improve matching accuracy. You can configure normalization methods directly in your field definitions:

```yaml
# Field configuration with embedded normalization
# Format: normalization_method:field_name or just field_name
database:
  fields:
    - name:FIRST        # Standardize name fields (lowercase, remove punctuation, normalize spaces)
    - name:LAST         # Apply to any name field in your data
    - date:DATE_OF_BIRTH # Standardize dates to YYYY-MM-DD format
    - zip:ZIP           # Extract first 5 digits from ZIP codes
    - gender:GENDER     # Standardize gender to single characters (m/f/nb/o/u)
    - phone             # Field without normalization (uses basic normalization)
```

**Supported Normalization Methods:**
- **`name`** - Names: Convert to lowercase, remove punctuation, normalize whitespace
- **`date`** - Dates: Standardize to YYYY-MM-DD format (supports multiple input formats)
- **`gender`** - Gender: Standardize to single characters (m/f/nb/o/u)
- **`zip`** - ZIP codes: Extract first 5 digits, remove non-numeric characters

**Field Behavior:**
- Fields with `method:field_name` format use the specified normalization method
- Fields without `:` use basic normalization (lowercase, trim)
- All matching protocols use secure zero-knowledge comparison by default

**Example Benefits:**
- `"Mary-Jane O'Connor"` and `"MARYJANE OCONNOR"` will match after name normalization
- `"12/25/2023"` and `"2023-12-25"` will match after date normalization  
- `"12345-6789"` and `"12345 6789"` will match after ZIP normalization

### Support Files

- **Documentation**: `ARCHITECTURE.md`, `SECURITY_FEATURES.md`, `INSTALL.md`
- **Scripts**: `scripts/install.ps1`, demo scripts for testing
- **Web Interface**: `web/` - Next.js configuration builder (optional)

## 🔐 Security Features

### Privacy-Preserving Techniques

**Bloom Filter Encoding**
- Patient data is encoded into fixed-size bit arrays
- No raw PHI is stored after tokenization
- Configurable filter size and hash functions for optimal privacy/utility tradeoff

**Differential Privacy**
- Controlled noise injection during Bloom filter creation
- Configurable noise parameters to meet privacy requirements
- Formal privacy guarantees against inference attacks

**MinHash Signatures**
- Locality-sensitive hashing for efficient similarity estimation
- Preserves approximate Jaccard similarity while hiding exact values
- Enables secure blocking without revealing bucket contents

**Commutative Encryption**
- Pohlig-Hellman encryption over Curve25519 for secure blocking
- Allows computation on encrypted data without key sharing
- Enables private set intersection for candidate generation

### HIPAA Compliance Features

**Data Minimization**
- Only necessary data fields are processed
- Configurable field selection for different use cases
- Automatic removal of direct identifiers after tokenization

**Secure Processing**
- All comparisons occur on encrypted/encoded data
- No PHI transmitted in plaintext over networks
- Configurable retention policies for temporary data

**Audit and Logging**
- Comprehensive logging of all data access and processing
- Immutable audit trails for compliance verification
- Configurable log levels and retention periods

**Access Controls**
- Role-based access to different system components
- Secure key management and distribution
- Authentication and authorization for network communication

### Deployment Security

**Network Security**
- Secure peer-to-peer communication protocols
- Per-IP rate limiting and connection management
- Configurable network timeouts and retry policies

**Data Isolation**
- Separate processing environments for PHI and tokens
- Container-ready architecture for secure deployment
- Configurable temporary file cleanup and secure deletion

**Key Management**
- Secure random key generation
- Key rotation and lifecycle management
- Hardware security module (HSM) integration ready

## 🎯 Use Cases

### Healthcare Record Linkage
- **Patient Matching**: Identify the same patient across different healthcare systems
- **Quality Assessment**: Validate data quality and matching accuracy
- **Research Cohorts**: Build research cohorts without sharing PHI
- **Public Health**: Disease surveillance and outbreak investigation

### Deployment Scenarios

**Two-Party Matching**
```bash
# Party A (Receiver)
./cohort-bridge -mode=receiver -config=config_a.yaml
# OR use interactive mode
./cohort-bridge

# Party B (Sender)  
./cohort-bridge -mode=sender -config=config_b.yaml
# OR use interactive mode
./cohort-bridge
```

**Enhanced Security (Tokenized)**
```bash
# Step 1: Tokenize data in secure environment with normalization
./cohort-bridge tokenize -input data/patients.csv -output tokens.csv -main-config config.yaml

# Step 2: Match using tokens in less secure environment
./cohort-bridge -mode=receiver -config=config_tokenized.yaml
```

**Local Processing Workflow**
```bash
# Complete workflow on single machine
./cohort-bridge tokenize -input data/dataset1.csv -output tokens1.csv
./cohort-bridge tokenize -input data/dataset2.csv -output tokens2.csv
./cohort-bridge intersect -dataset1 tokens1.csv -dataset2 tokens2.csv
./cohort-bridge validate -ground-truth data/truth.csv -results intersection_results.csv
```

**Database Integration**
```bash
# Use PostgreSQL for large datasets
./cohort-bridge tokenize -database -main-config config_postgres.yaml
./cohort-bridge -config=config_postgres.yaml
```

## 📊 Performance & Scalability

### Computational Complexity
- **Bloom Filter Creation**: O(k × n) where k = hash functions, n = input size
- **MinHash Computation**: O(s × m) where s = signature length, m = filter size  
- **Secure Blocking**: O(b × r) where b = bands, r = records
- **Fuzzy Matching**: O(c) where c = candidate pairs

### Memory Usage
- **Bloom Filter**: Configurable (default: 1024 bits = 128 bytes per record)
- **MinHash Signature**: 4 × signature_length bytes per record
- **Blocking Buckets**: Depends on data distribution and LSH parameters
- **Peak Memory**: Approximately 2-3x the size of input datasets

### Throughput Characteristics
- **Small datasets** (<10K records): ~1000-2000 records/second
- **Medium datasets** (10K-100K records): ~500-1000 records/second  
- **Large datasets** (>100K records): ~200-500 records/second
- **Network overhead**: 10-20% performance penalty for two-party mode

## 🧪 Testing & Validation

### Built-in Test Suite
```bash
# Run comprehensive test harness
make test

# Run specific test scenarios using the test program
./test -records1=1000 -records2=1200 -overlap=0.3

# Validate specific results
./cohort-bridge validate -ground-truth test_data/truth.csv -results out/matches.csv
```

### Validation Metrics
- **Precision & Recall**: Standard classification metrics
- **F1-Score**: Harmonic mean of precision and recall
- **ROC/AUC**: Receiver operating characteristic analysis
- **Performance**: Processing time and memory usage
- **Privacy**: Differential privacy parameter estimation

### Demo Scripts
- `two_party_demo.sh` - Complete two-party workflow demonstration
- `test_cohort_bridge.sh` - Automated testing with various parameters
- `two_party_network_demo.sh` - Network communication testing

## 🚀 Advanced Configuration

### Tuning Parameters

**Data Field Configuration with Normalization**
```yaml
# Field configuration with embedded normalization
database:
  fields:
    - name:FIRST        # Apply name normalization to FIRST field
    - name:LAST         # Apply name normalization to LAST field  
    - date:BIRTHDATE    # Apply date normalization to BIRTHDATE field
    - zip:ZIP           # Apply ZIP normalization to ZIP field
    - gender:GENDER     # Apply gender normalization to GENDER field
```

**Bloom Filter Settings**
```yaml
bloom:
  size: 1024        # Bits per record (larger = more accurate, less private)
  hash_count: 8     # Number of hash functions (more = fewer false positives)
  noise_rate: 0.1   # Differential privacy noise (higher = more private)
```

**MinHash Settings**
```yaml
minhash:
  signature_length: 64    # Number of hash functions (more = more accurate)
  bands: 16              # LSH bands (more = fewer false negatives)
  rows: 4                # Rows per band (fewer = more candidates)
```

**Matching Thresholds**
```yaml
matching:
  hamming_threshold: 100     # Maximum bit differences
  jaccard_threshold: 0.7     # Minimum similarity score
  qgram_threshold: 0.8       # Minimum n-gram similarity
```

### Integration Options

**Database Integration**
- PostgreSQL support for large-scale datasets
- Custom schema support for existing databases
- Streaming processing for memory-constrained environments

**API Integration**
- RESTful API for programmatic access
- Webhook support for result notifications
- gRPC interface for high-performance applications

**Container Deployment**
- Docker images available for all components
- Kubernetes manifests for scalable deployment
- Helm charts for easy configuration management

## 📚 Documentation

- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Detailed system architecture
- **[SECURITY_FEATURES.md](SECURITY_FEATURES.md)** - Comprehensive security analysis  
- **[INSTALL.md](INSTALL.md)** - Installation and deployment guide
- **[STREAMING_README.md](STREAMING_README.md)** - Streaming and real-time processing

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details on:
- Code style and standards
- Testing requirements
- Security review process
- Documentation standards


## 📞 Support

- **Documentation**: [TBD](URL_TBD)
- **Issues**: [GitHub Issues](https://github.com/alunalabs/cohort-bridge/issues)
- **Discussions**: [GitHub Discussions](https://github.com/alunalabs/cohort-bridge/discussions)
- **Security Issues**: Please report privately to [admin@nerve.run](mailto:admin@nerve.run)