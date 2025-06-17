# HIPAA-Compliant Decentralized Fuzzy Matching System

> **⚠️ Work in Progress Notice**  
> This system is currently under active development. While the core functionality for privacy-preserving record linkage is implemented, some features are still in progress or experimental:
> - The q-gram based matching is newly implemented and may need tuning
> - Secure multi-party computation protocols (garbled circuits, VOLE) are currently placeholders
> - Performance optimizations for large-scale matching are pending
> - Integration with external systems and APIs is not yet complete
> 
> Please use with caution in production environments and verify results carefully.

A secure, privacy-preserving record linkage system using Bloom filters, MinHash signatures, and commutative encryption for HIPAA-compliant patient matching across multiple healthcare institutions.

## 🏗️ Architecture

This system implements a complete pipeline for privacy-preserving record linkage:

1. **Privacy-Preserving Record Linkage (PPRL)** - Bloom filters with noise injection and MinHash signatures
2. **Secure Blocking** - Commutative encryption using Pohlig-Hellman over Curve25519
3. **Candidate Matching** - LSH-based blocking with encrypted bucket exchange
4. **Fuzzy Matching** - Hamming distance and Jaccard similarity with secure comparison protocols
5. **Test Harness** - Comprehensive validation with synthetic datasets

## 🔧 Components

### Core Modules

- **`internal/pprl/`** - Bloom filters, MinHash, and storage
  - `bloom.go` - Configurable Bloom filter with noise injection
  - `minhash.go` - MinHash signatures for Jaccard similarity estimation
  - `storage.go` - JSONL-based persistence layer

- **`internal/crypto/`** - Cryptographic primitives
  - `commutative.go` - Pohlig-Hellman commutative encryption over Curve25519

- **`internal/match/`** - Matching pipeline
  - `blocking.go` - Secure blocking using LSH and commutative encryption
  - `fuzzy.go` - Secure fuzzy matching with placeholder for garbled circuits/VOLE
  - `pipeline.go` - Main orchestrator for the complete matching process
  - `testharness.go` - Comprehensive testing framework

### Command Line Tools

- **`cmd/demo/`** - Demonstration application with multiple modes

## 🚀 Quick Start

### Prerequisites

- Go 1.24.3 or later
- Required dependencies (automatically managed via `go.mod`)

### Installation

```bash
git clone <repository-url>
cd cohort-bridge
go mod download
```

### Initial Setup

1. **Copy example configuration**:
   ```bash
   cp config.example.yaml config.yaml
   ```

2. **Set up your data directory**:
   - Place your CSV patient data files in the `data/` directory
   - See `data/README.md` for format requirements and examples
   - Update `config.yaml` to point to your data files

3. **Generate a private key** (for production):
   ```bash
   openssl rand -hex 32
   ```
   Replace `YOUR_PRIVATE_KEY_HERE_32_HEX_CHARACTERS` in `config.yaml` with the generated key.

### Configuration Examples

- `config.example.yaml` - Basic single-party configuration
- `config_example_party_a.yaml` - Party A in two-party matching
- `config_example_party_b.yaml` - Party B in two-party matching

### Running the Demo

```bash
# Run the test harness with default parameters
go run cmd/demo/main.go -mode=test -verbose

# Customize the test parameters
go run cmd/demo/main.go \
  -mode=test \
  -records1=200 \
  -records2=250 \
  -overlap=0.4 \
  -noise=0.15 \
  -bloom-size=2048 \
  -hash-count=10 \
  -minhash-sigs=128 \
  -hamming-threshold=150 \
  -jaccard-threshold=0.6 \
  -qgram-threshold=0.8 \
  -output=./results \
  -verbose
```

### Command Line Options

- `mode` - Operation mode: `test`, `single`, or `two-party`
- `records1/records2` - Number of records in each dataset
- `overlap` - Overlap rate between datasets (0.0-1.0)
- `noise` - Noise rate for data corruption (0.0-1.0)
- `bloom-size` - Bloom filter size in bits
- `hash-count` - Number of hash functions for Bloom filter
- `minhash-sigs` - Number of MinHash signatures
- `hamming-threshold` - Maximum Hamming distance for matching
- `jaccard-threshold` - Minimum Jaccard similarity for matching
- `qgram-threshold` - Minimum q-gram similarity for matching
- `output` - Output directory for results
- `verbose` - Enable detailed logging

## 📊 Example Output

```
🧪 Running Secure Fuzzy Matching Test Harness
==================================================
📊 Test Configuration:
  Dataset 1: 100 records
  Dataset 2: 120 records
  Overlap rate: 30.0%
  Noise rate: 10.0%
  Bloom filter: 1024 bits, 8 hashes
  MinHash signatures: 64
  Hamming threshold: 100
  Jaccard threshold: 0.70

🏆 Test Results Summary
=========================
📈 Matching Statistics:
  Ground truth matches: 30
  Candidate pairs generated: 892
  Total matches found: 28
  Matching buckets: 15

🎯 Evaluation Metrics:
  True Positives: 26
  False Positives: 2
  False Negatives: 4
  Precision: 0.929
  Recall: 0.867
  F1-Score: 0.897

⚡ Performance:
  Processing time: 145 ms
  Records processed: 220
  Throughput: 1517.2 records/second

🔧 Blocking Statistics:
  Total buckets: 15
  Average bucket size: 4.2
  Median bucket size: 3
  Max bucket size: 12

🔍 Quality Assessment:
  Precision: Excellent (0.929)
  Recall: Good (0.867)
  Overall: Good (F1: 0.897)
```

## 🔐 Security Features

### Privacy-Preserving Techniques

1. **Bloom Filter Encoding** - Patient data is encoded into fixed-size bit arrays
2. **Noise Injection** - Differential privacy through controlled bit flipping
3. **MinHash Signatures** - Locality-sensitive hashing for similarity estimation
4. **Commutative Encryption** - Secure blocking without revealing bucket contents

### Cryptographic Protocols

- **Pohlig-Hellman over Curve25519** - For commutative encryption in blocking
- **Placeholder for Garbled Circuits** - Future secure Hamming distance computation
- **Placeholder for VOLE-based PSI** - Future fuzzy private set intersection

### HIPAA Compliance

- No PHI is stored in plaintext after de-identification
- All comparisons occur on encrypted/encoded data
- Configurable noise parameters for differential privacy
- Modular design supports formal security analysis

## 🧪 Testing and Validation

### Test Harness Features

- **Synthetic Data Generation** - Configurable patient records with realistic noise
- **Ground Truth Tracking** - Known matches for evaluation
- **Comprehensive Metrics** - Precision, recall, F1-score, and performance statistics
- **Parameter Tuning** - Easy experimentation with different configurations

### Evaluation Metrics

- **True/False Positives and Negatives** - Standard classification metrics
- **Precision and Recall** - Quality of matching results
- **F1-Score** - Harmonic mean of precision and recall
- **Performance Metrics** - Processing time and throughput
- **Blocking Efficiency** - Bucket size distribution and candidate reduction

## 🏥 Real-World Deployment

### Preparation for Production

1. **Replace Placeholder Protocols** - Implement full garbled circuits or VOLE-based PSI
2. **Add Network Layer** - Implement gRPC or secure socket communication
3. **Key Management** - Secure key generation and distribution
4. **Audit Logging** - Comprehensive logging for compliance
5. **Performance Optimization** - GPU acceleration for cryptographic operations

### Scalability Considerations

- **Parallel Processing** - Multi-threaded Bloom filter operations
- **Distributed Blocking** - Sharded blocking buckets across nodes
- **Streaming Processing** - Handle large datasets with bounded memory
- **Load Balancing** - Distribute matching workload across multiple servers

## 📈 Performance Characteristics

### Computational Complexity

- **Bloom Filter Creation**: O(k × n) where k = hash functions, n = input size
- **MinHash Computation**: O(s × m) where s = signature length, m = filter size
- **Blocking**: O(b × r) where b = bands, r = records
- **Fuzzy Matching**: O(c) where c = candidate pairs

### Memory Usage

- **Bloom Filter**: Configurable (default: 1024 bits = 128 bytes per record)
- **MinHash Signature**: 4 × signature_length bytes per record
- **Blocking Buckets**: Depends on data distribution and parameters
- **Candidate Pairs**: O(blocking efficiency × total record pairs)

## 🔮 Future Enhancements

### Cryptographic Protocols

- [ ] Full garbled circuit implementation for secure Hamming distance
- [ ] VOLE-based fuzzy PSI for large-scale matching
- [ ] Zero-knowledge proofs for result verification
- [ ] Homomorphic encryption for aggregated statistics

### System Features

- [ ] Real-time matching API
- [ ] Web-based configuration dashboard
- [ ] Integration with HL7 FHIR standards
- [ ] Multi-party matching (>2 participants)
- [ ] Federated learning for parameter optimization

### Performance Optimizations

- [ ] GPU acceleration for cryptographic operations
- [ ] Advanced blocking strategies (semantic blocking)
- [ ] Incremental matching for streaming data
- [ ] Caching and memoization for repeated computations

## 📚 Research and References

This implementation is based on current research in privacy-preserving record linkage:

- **Bloom Filter PPRL**: Schnell et al. (2009) "Privacy-preserving record linkage using Bloom filters"
- **MinHash LSH**: Broder (1997) "On the resemblance and containment of documents"  
- **Commutative Encryption**: Pohlig & Hellman (1978) "An improved algorithm for computing logarithms over GF(p)"
- **Secure Multiparty Computation**: Yao (1982) "Protocols for secure computations"

## 📄 License

[Specify your license here]

## 🤝 Contributing

Contributions are welcome! Please read our contributing guidelines and submit pull requests for any improvements.

## 📞 Support

For questions or support, please [contact information or issue tracker].

## Output Directory Structure

All generated outputs are automatically saved to the `out/` directory:

- **Match Results**: `out/matches_YYYYMMDD_HHMMSS_<connection_id>.csv`
- **Detailed Results**: `out/match_details_YYYYMMDD_HHMMSS_<connection_id>.csv`
- **Validation Results**: `out/validation_results.csv` (when using validate command)
- **Test Outputs**: `out/dataset1.jsonl`, `out/dataset2.jsonl`, etc.

All log files are saved to the `logs/` directory (when logging is enabled).

The system automatically creates these directories if they don't exist.

## Privacy-Preserving Record Linkage (PPRL) System

This system provides secure, privacy-preserving record linkage between two parties using Bloom filters and cryptographic protocols. It enables organizations to identify matching records without sharing raw personally identifiable information (PHI).

## 🔐 **NEW: Tokenization Mode**

The system now supports **two modes of operation**:

### **Raw PHI Mode** (Original)
- Processes raw patient data directly during matching
- Suitable for trusted environments
- PHI is converted to Bloom filters in real-time

### **Tokenized Mode** (Enhanced Security)
- **Pre-processes PHI** into privacy-preserving tokens using the `tokenize` tool
- **Separates PHI handling** from the matching process
- **Enhanced security** - raw PHI never leaves the tokenization environment
- **HIPAA-friendly** deployment with isolated PHI processing

```bash
# Step 1: Tokenize your data (secure environment)
./tokenize.exe -input data/patients_a.csv -output out/tokens_party_a.json

# Step 2: Run matching with tokens (can be less secure environment)
./agent.exe -mode receive -config config_tokenized.yaml
```

See [TOKENIZATION_GUIDE.md](TOKENIZATION_GUIDE.md) for complete documentation.

## Features

// ... existing content ... 