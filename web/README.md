# CohortBridge Config Builder

A beautiful, user-friendly web interface for creating and customizing YAML configuration files for CohortBridge - a privacy-preserving record linkage system.

## Features

- 🎨 **Beautiful UI** - Modern, sleek design with Tailwind CSS
- ⚡ **Type Safety** - Built with TypeScript for reliability
- 🔧 **Multiple Config Types** - Support for different configuration scenarios:
  - Basic Configuration - Simple record linkage setup
  - PostgreSQL Configuration - Database connectivity
  - Secure Configuration - Enhanced security and logging
  - Tokenized Configuration - Pre-tokenized data support
  - Network Configuration - Multi-party setups

- 📱 **Responsive Design** - Works on desktop and mobile
- 💾 **Export & Download** - Generate and download YAML files
- 📋 **Copy to Clipboard** - Easy sharing of configurations
- ✅ **Real-time Preview** - See your YAML output as you build

## Getting Started

1. **Install dependencies:**
   ```bash
   npm install
   ```

2. **Run the development server:**
   ```bash
   npm run dev
   ```

3. **Open your browser:**
   Navigate to [http://localhost:3000](http://localhost:3000)

## Configuration Types

### Basic Configuration
Simple setup for basic privacy-preserving record linkage with matching algorithm parameters.

### PostgreSQL Configuration  
Connect to PostgreSQL databases with full connection settings and table configuration.

### Secure Configuration
Enhanced security features including:
- IP address whitelisting
- Connection rate limiting
- Comprehensive logging
- Security audit trails
- Timeout configurations

### Tokenized Configuration
Work with pre-tokenized data files for enhanced privacy.

### Network Configuration
Multi-party network setups with peer connectivity and matching parameters.

## Technology Stack

- **Next.js 14** - React framework with App Router
- **TypeScript** - Type safety and better developer experience
- **Tailwind CSS** - Utility-first CSS framework
- **React Hook Form** - Form state management
- **Zod** - Schema validation
- **js-yaml** - YAML parsing and generation
- **Lucide React** - Beautiful icons

## Usage

1. Select a configuration type from the main dashboard
2. Fill out the form with your specific settings
3. See the real-time YAML preview on the right
4. Copy to clipboard or download the generated configuration file
5. Use the configuration file with your CohortBridge installation

## Contributing

This application is part of the CohortBridge project. Feel free to contribute improvements and new features.

## License

This project follows the same license as the parent CohortBridge project.
