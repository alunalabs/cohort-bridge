'use client';

import { useRouter } from 'next/navigation';
import {
    Settings,
    Terminal,
    Shield,
    Database,
    Network,
    FileText,
    ArrowRight,
    Copy,
    Check,
    Download,
    Github,
    BookOpen,
    Zap,
    Lock,
    Search
} from 'lucide-react';
import { useState } from 'react';

// CodeBlock component matching get-started page styling
const CodeBlock = ({ children, command, language }: {
    children: string;
    command?: boolean;
    language?: string;
}) => {
    const [copied, setCopied] = useState(false);

    const copyCommand = async (text: string) => {
        try {
            await navigator.clipboard.writeText(text);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch (err) {
            console.error('Failed to copy: ', err);
        }
    };

    return (
        <div className="bg-slate-50 border border-slate-200 rounded p-3 relative">
            <code className="text-slate-700 text-xs font-mono">
                {children}
            </code>
            {command && (
                <button
                    onClick={() => copyCommand(children)}
                    className="absolute top-2 right-2 p-1 hover:bg-slate-200 rounded transition-colors"
                    title="Copy to clipboard"
                >
                    {copied ? (
                        <Check className="h-3 w-3 text-green-600" />
                    ) : (
                        <Copy className="h-3 w-3 text-gray-500" />
                    )}
                </button>
            )}
        </div>
    );
};

// Parameter table component
const ParameterTable = ({ parameters }: {
    parameters: Array<{
        name: string;
        description: string;
        default: string;
        color?: string;
    }>
}) => (
    <div className="overflow-x-auto">
        <div className="bg-white rounded-lg border border-gray-200 overflow-hidden shadow-sm">
            <table className="w-full text-sm">
                <thead>
                    <tr className="bg-gray-50 border-b border-gray-200">
                        <th className="px-6 py-3 text-left font-semibold text-gray-900">Parameter</th>
                        <th className="px-6 py-3 text-left font-semibold text-gray-900">Description</th>
                        <th className="px-6 py-3 text-left font-semibold text-gray-900">Default</th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                    {parameters.map((param, index) => (
                        <tr key={index} className="hover:bg-gray-50 transition-colors">
                            <td className="px-6 py-3">
                                <code className={`${param.color || 'bg-gray-100 text-gray-800'} px-2 py-1 rounded text-xs font-mono`}>
                                    {param.name}
                                </code>
                            </td>
                            <td className="px-6 py-3 text-gray-700">{param.description}</td>
                            <td className="px-6 py-3 text-gray-600 font-mono text-xs">{param.default}</td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    </div>
);

export default function DocsPage() {
    const router = useRouter();
    const [activeTab, setActiveTab] = useState('overview');
    const [selectedOS, setSelectedOS] = useState<'windows' | 'macos' | 'linux'>('windows');

    const installCommands = {
        windows: {
            download: 'curl -L "https://github.com/alunalabs/cohort-bridge/releases/latest/download/cohort-bridge-windows.exe" -o cohort-bridge.exe',
            install: 'move cohort-bridge.exe C:\\Windows\\System32\\',
            verify: 'cohort-bridge --version'
        },
        macos: {
            download: 'curl -L "https://github.com/alunalabs/cohort-bridge/releases/latest/download/cohort-bridge-macos" -o cohort-bridge',
            install: 'chmod +x cohort-bridge && sudo mv cohort-bridge /usr/local/bin/',
            verify: 'cohort-bridge --version'
        },
        linux: {
            download: 'wget https://github.com/alunalabs/cohort-bridge/releases/latest/download/cohort-bridge-linux',
            install: 'chmod +x cohort-bridge-linux && sudo mv cohort-bridge-linux /usr/local/bin/cohort-bridge',
            verify: 'cohort-bridge --version'
        }
    };

    const tabs = [
        { id: 'overview', name: 'Overview', icon: BookOpen },
        { id: 'pprl', name: 'Peer-to-Peer Matching', icon: Network },
        { id: 'tokenize', name: 'Data Tokenization', icon: Lock },
        { id: 'validate', name: 'Result Validation', icon: Search },
        { id: 'configuration', name: 'Configuration', icon: Settings },
        { id: 'examples', name: 'Examples', icon: Zap },
    ];

    const renderTabContent = () => {
        switch (activeTab) {
            case 'overview':
                return (
                    <div className="space-y-8">
                        <div>
                            <h2 className="text-3xl font-bold text-gray-900 mb-4">Getting Started</h2>
                            <p className="text-lg text-gray-600 mb-6">
                                CohortBridge is a privacy-preserving record linkage system that enables secure data matching
                                between organizations without exposing sensitive information.
                            </p>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Installation</h3>
                            <p className="text-gray-600 mb-4">
                                Download the latest CohortBridge CLI from our releases page for your platform:
                            </p>

                            {/* OS Selection */}
                            <div className="flex space-x-2 mb-4">
                                {(['windows', 'macos', 'linux'] as const).map((os) => (
                                    <button
                                        key={os}
                                        onClick={() => setSelectedOS(os)}
                                        className={`px-3 py-1 rounded text-sm font-medium transition-colors ${selectedOS === os
                                            ? 'bg-gradient-to-r from-blue-500 to-purple-500 text-white'
                                            : 'bg-gray-100 text-gray-700 hover:bg-gray-200 border border-gray-300'
                                            }`}
                                    >
                                        {os === 'macos' ? 'macOS' : os.charAt(0).toUpperCase() + os.slice(1)}
                                    </button>
                                ))}
                            </div>

                            {/* Commands */}
                            <div className="space-y-4">
                                {/* Download */}
                                <div>
                                    <h4 className="text-sm font-medium text-gray-800 mb-2">Download</h4>
                                    <CodeBlock command>{installCommands[selectedOS].download}</CodeBlock>
                                </div>

                                {/* Install */}
                                <div>
                                    <h4 className="text-sm font-medium text-gray-800 mb-2">Install</h4>
                                    <CodeBlock command>{installCommands[selectedOS].install}</CodeBlock>
                                </div>

                                {/* Verify */}
                                <div>
                                    <h4 className="text-sm font-medium text-gray-800 mb-2">Verify Installation</h4>
                                    <CodeBlock command>{installCommands[selectedOS].verify}</CodeBlock>
                                </div>
                            </div>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Interactive Mode</h3>
                            <p className="text-gray-600 mb-4">
                                The easiest way to get started is with interactive mode. Simply run the tool without any arguments:
                            </p>
                            <CodeBlock command>./cohort-bridge</CodeBlock>
                            <p className="text-gray-600 mt-4">
                                This launches a guided interface that helps you choose the right operation and configure all parameters.
                            </p>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Command Structure</h3>
                            <p className="text-gray-600 mb-4">
                                CohortBridge uses a subcommand structure. Each operation has its own focused command:
                            </p>
                            <CodeBlock>
                                {`# General syntax
./cohort-bridge <subcommand> [options]

# Get help for any subcommand
./cohort-bridge <subcommand> -help

# Interactive mode for any subcommand
./cohort-bridge <subcommand> -interactive`}
                            </CodeBlock>
                        </div>

                        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                            <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
                                <Network className="h-8 w-8 text-blue-600 mb-3" />
                                <h4 className="font-semibold text-blue-900 mb-2">pprl</h4>
                                <p className="text-blue-800 text-sm">
                                    Main command for peer-to-peer privacy-preserving record linkage
                                </p>
                            </div>
                            <div className="bg-green-50 border border-green-200 rounded-lg p-6">
                                <Lock className="h-8 w-8 text-green-600 mb-3" />
                                <h4 className="font-semibold text-green-900 mb-2">tokenize</h4>
                                <p className="text-green-800 text-sm">
                                    Convert sensitive data into privacy-preserving tokens
                                </p>
                            </div>
                            <div className="bg-purple-50 border border-purple-200 rounded-lg p-6">
                                <Search className="h-8 w-8 text-purple-600 mb-3" />
                                <h4 className="font-semibold text-purple-900 mb-2">validate</h4>
                                <p className="text-purple-800 text-sm">
                                    Test accuracy against known ground truth data
                                </p>
                            </div>
                        </div>
                    </div>
                );

            case 'pprl':
                return (
                    <div className="space-y-8">
                        <div>
                            <h2 className="text-3xl font-bold text-gray-900 mb-4">Peer-to-Peer Privacy-Preserving Record Linkage</h2>
                            <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 mb-6">
                                <div className="flex items-start space-x-3">
                                    <Network className="h-6 w-6 text-blue-600 mt-1 flex-shrink-0" />
                                    <div>
                                        <h3 className="font-semibold text-blue-900 mb-2">Main Command</h3>
                                        <p className="text-blue-800">
                                            The <code className="bg-blue-100 px-1 rounded">pprl</code> command orchestrates a complete 7-step peer-to-peer workflow
                                            for finding intersections between two parties' databases using zero-knowledge protocols.
                                        </p>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Interactive Mode (Recommended)</h3>
                            <p className="text-gray-600 mb-3">
                                Interactive mode guides you through configuration and shows prompts for confirmation at each step.
                                This is the safest way to run PPRL operations.
                            </p>
                            <CodeBlock command>./cohort-bridge pprl</CodeBlock>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Configuration File Mode</h3>
                            <p className="text-gray-600 mb-3">
                                Use a pre-created configuration file. Still shows confirmation prompts unless combined with -force.
                            </p>
                            <CodeBlock command>./cohort-bridge pprl -config config.yaml</CodeBlock>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Automated Mode (For Scripts)</h3>
                            <p className="text-gray-600 mb-3">
                                Use -force to skip all confirmation prompts. Essential when calling from other programs or scripts.
                            </p>
                            <CodeBlock command>./cohort-bridge pprl -config config.yaml -force</CodeBlock>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Allow Multiple Matches</h3>
                            <p className="text-gray-600 mb-3">
                                By default, CohortBridge enforces 1:1 matching (each record matches at most one other record).
                                Use -allow-duplicates to enable 1:many matching.
                            </p>
                            <CodeBlock command>./cohort-bridge pprl -config config.yaml -allow-duplicates</CodeBlock>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Complete Parameter Reference</h3>
                            <ParameterTable parameters={[
                                { name: '-config', description: 'Path to YAML configuration file containing peer settings, thresholds, and database configuration', default: 'none (required)', color: 'bg-blue-100 text-blue-800' },
                                { name: '-interactive', description: 'Force interactive mode even when config file is provided. Useful for reviewing settings.', default: 'false', color: 'bg-blue-100 text-blue-800' },
                                { name: '-force', description: 'Skip all confirmation prompts and run automatically. Required when calling from scripts or other programs.', default: 'false', color: 'bg-blue-100 text-blue-800' },
                                { name: '-allow-duplicates', description: 'Allow 1:many matching where one record can match multiple peer records. Default is 1:1 matching only.', default: 'false', color: 'bg-blue-100 text-blue-800' },
                                { name: '-help', description: 'Show detailed help message with examples and configuration requirements', default: 'false', color: 'bg-blue-100 text-blue-800' },
                            ]} />
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Workflow</h3>
                            <p className="text-gray-600 mb-6">
                                The PPRL command orchestrates a complete privacy-preserving workflow:
                            </p>
                            <div className="space-y-4">
                                {[
                                    { step: 1, title: 'Configuration Loading', desc: 'Reads your config file and validates settings' },
                                    { step: 2, title: 'Data Tokenization', desc: 'Converts your PHI data into privacy-preserving tokens' },
                                    { step: 3, title: 'Peer Connection', desc: 'Establishes secure connection with the other party' },
                                    { step: 4, title: 'Token Exchange', desc: 'Securely exchanges tokenized data' },
                                    { step: 5, title: 'Intersection Computing', desc: 'Finds matches using zero-knowledge protocols' },
                                    { step: 6, title: 'Result Validation', desc: 'Compares results between parties to ensure accuracy' },
                                ].map((item) => (
                                    <div key={item.step} className="flex items-start space-x-4">
                                        <span className="bg-blue-100 text-blue-800 rounded-full w-8 h-8 flex items-center justify-center text-sm font-semibold flex-shrink-0">
                                            {item.step}
                                        </span>
                                        <div>
                                            <h4 className="font-semibold text-gray-900">{item.title}</h4>
                                            <p className="text-gray-600">{item.desc}</p>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    </div>
                );

            case 'tokenize':
                return (
                    <div className="space-y-8">
                        <div>
                            <h2 className="text-3xl font-bold text-gray-900 mb-4">Privacy-Preserving Data Tokenization</h2>
                            <div className="bg-green-50 border border-green-200 rounded-lg p-6 mb-6">
                                <div className="flex items-start space-x-3">
                                    <Shield className="h-6 w-6 text-green-600 mt-1 flex-shrink-0" />
                                    <div>
                                        <h3 className="font-semibold text-green-900 mb-2">Data Protection</h3>
                                        <p className="text-green-800">
                                            The <code className="bg-green-100 px-1 rounded">tokenize</code> command converts your sensitive PHI data into privacy-preserving tokens
                                            using Bloom filters and MinHash signatures. Files are encrypted by default for maximum security.
                                        </p>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Interactive Mode (Recommended)</h3>
                            <p className="text-gray-600 mb-3">
                                Interactive mode prompts for all settings including data source, encryption options, field configuration, and batch size.
                            </p>
                            <CodeBlock command>./cohort-bridge tokenize</CodeBlock>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">File-Based Tokenization</h3>
                            <p className="text-gray-600 mb-3">
                                Process data from a CSV or JSON file. Encryption key is auto-generated and saved to a .key file.
                            </p>
                            <CodeBlock command>./cohort-bridge tokenize -input data/patients.csv -output tokens.csv.enc</CodeBlock>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Database Tokenization</h3>
                            <p className="text-gray-600 mb-3">
                                Use database connection and field configuration from main config file instead of a CSV file.
                            </p>
                            <CodeBlock command>./cohort-bridge tokenize -database -main-config config.yaml</CodeBlock>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Custom Encryption Key</h3>
                            <p className="text-gray-600 mb-3">
                                Provide your own 32-byte (64-character) hex encryption key instead of auto-generation.
                            </p>
                            <CodeBlock command>./cohort-bridge tokenize -input data.csv -encryption-key a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456</CodeBlock>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Disable Encryption (Not Recommended)</h3>
                            <p className="text-gray-600 mb-3">
                                Store tokens in plaintext. Only use for testing - never in production with real PHI data.
                            </p>
                            <CodeBlock command>./cohort-bridge tokenize -input data.csv -output tokens.csv -no-encryption</CodeBlock>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Automated Mode</h3>
                            <p className="text-gray-600 mb-3">
                                Skip confirmation prompts for use in scripts or automated workflows.
                            </p>
                            <CodeBlock command>./cohort-bridge tokenize -input data.csv -output tokens.csv.enc -force</CodeBlock>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Complete Parameter Reference</h3>
                            <ParameterTable parameters={[
                                { name: '-input', description: 'Input file containing PHI data (CSV, JSON, or TXT format)', default: 'prompted in interactive mode', color: 'bg-emerald-100 text-emerald-800' },
                                { name: '-output', description: 'Output file for encrypted tokenized data. Extension .enc added automatically if encryption enabled.', default: 'auto-generated based on input', color: 'bg-emerald-100 text-emerald-800' },
                                { name: '-main-config', description: 'Configuration file to read field names and normalization settings from', default: 'config.yaml', color: 'bg-emerald-100 text-emerald-800' },
                                { name: '-input-format', description: 'Input file format: csv, json, postgres. Auto-detected from file extension if not specified.', default: 'csv', color: 'bg-emerald-100 text-emerald-800' },
                                { name: '-output-format', description: 'Output file format: csv, json. Usually matches input format.', default: 'csv', color: 'bg-emerald-100 text-emerald-800' },
                                { name: '-batch-size', description: 'Number of records to process in each batch. Larger values use more memory but may be faster.', default: '1000', color: 'bg-emerald-100 text-emerald-800' },
                                { name: '-database', description: 'Use database connection from main config instead of reading from a file', default: 'false', color: 'bg-emerald-100 text-emerald-800' },
                                { name: '-minhash-seed', description: 'Seed for deterministic MinHash generation. Must be same across both parties for matching.', default: '0PsRm4KNmgRSY8ynApUtpXjeO19S7OUE', color: 'bg-emerald-100 text-emerald-800' },
                                { name: '-encryption-key', description: '32-byte hex encryption key (64 characters). Auto-generated and saved to .key file if not provided.', default: 'auto-generated', color: 'bg-emerald-100 text-emerald-800' },
                                { name: '-no-encryption', description: 'Disable AES-256-GCM encryption. Output will be plaintext. NOT recommended for production.', default: 'false', color: 'bg-emerald-100 text-emerald-800' },
                                { name: '-interactive', description: 'Force interactive mode even when other parameters are provided', default: 'false', color: 'bg-emerald-100 text-emerald-800' },
                                { name: '-force', description: 'Skip all confirmation prompts. Required when calling from scripts or automated systems.', default: 'false', color: 'bg-emerald-100 text-emerald-800' },
                                { name: '-help', description: 'Show detailed help message with examples and encryption information', default: 'false', color: 'bg-emerald-100 text-emerald-800' },
                            ]} />
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Data Normalization</h3>
                            <p className="text-gray-600 mb-4">
                                CohortBridge automatically normalizes your data during tokenization to improve matching accuracy.
                                Configure normalization in your field definitions:
                            </p>
                            <CodeBlock>
                                {`# In your config.yaml file
database:
  fields:
    - name:FIRST        # Apply name normalization
    - name:LAST         # Apply name normalization
    - date:BIRTHDATE    # Apply date normalization
    - zip:ZIP           # Apply ZIP normalization
    - gender:GENDER     # Apply gender normalization
    - phone             # Basic normalization`}
                            </CodeBlock>
                            <div className="mt-6 grid grid-cols-1 md:grid-cols-2 gap-4">
                                {[
                                    { type: 'name', color: 'bg-blue-100 text-blue-800', desc: 'Standardizes names (lowercase, remove punctuation)' },
                                    { type: 'date', color: 'bg-purple-100 text-purple-800', desc: 'Standardizes dates to YYYY-MM-DD format' },
                                    { type: 'zip', color: 'bg-green-100 text-green-800', desc: 'Extracts first 5 digits from ZIP codes' },
                                    { type: 'gender', color: 'bg-pink-100 text-pink-800', desc: 'Standardizes to single characters (m/f/nb/o/u)' },
                                ].map((item) => (
                                    <div key={item.type} className="flex items-center space-x-3">
                                        <span className={`${item.color} px-2 py-1 rounded text-xs font-mono`}>
                                            {item.type}
                                        </span>
                                        <span className="text-sm text-gray-600">{item.desc}</span>
                                    </div>
                                ))}
                            </div>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Encryption & Security</h3>
                            <p className="text-gray-600 mb-4">
                                By default, tokenized files are encrypted with AES-256. An encryption key is automatically generated and saved:
                            </p>
                            <CodeBlock>
                                {`# Files created during tokenization
tokens.csv.enc     # Encrypted tokenized data
tokens.key         # Encryption key (keep this secure!)`}
                            </CodeBlock>
                            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4 mt-6">
                                <p className="text-yellow-800">
                                    <strong>Important:</strong> Keep the .key file secure and never share it. You'll need it to decrypt the tokens later.
                                </p>
                            </div>
                        </div>
                    </div>
                );

            case 'validate':
                return (
                    <div className="space-y-8">
                        <div>
                            <h2 className="text-3xl font-bold text-gray-900 mb-4">Result Validation</h2>
                            <div className="bg-purple-50 border border-purple-200 rounded-lg p-6 mb-6">
                                <div className="flex items-start space-x-3">
                                    <Database className="h-6 w-6 text-purple-600 mt-1 flex-shrink-0" />
                                    <div>
                                        <h3 className="font-semibold text-purple-900 mb-2">Quality Assurance</h3>
                                        <p className="text-purple-800">
                                            The <code className="bg-purple-100 px-1 rounded">validate</code> command tests the accuracy of matching results against known ground truth data.
                                            This helps potential users understand that the matching algorithms work correctly.
                                        </p>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Interactive Mode (Recommended)</h3>
                            <p className="text-gray-600 mb-3">
                                Interactive mode prompts for configuration files, ground truth data, thresholds, and output preferences.
                            </p>
                            <CodeBlock command>./cohort-bridge validate</CodeBlock>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Standard Validation</h3>
                            <p className="text-gray-600 mb-3">
                                Compare matching results against known ground truth with default thresholds.
                            </p>
                            <CodeBlock command>./cohort-bridge validate -config1 config_a.yaml -config2 config_b.yaml -ground-truth data/truth.csv</CodeBlock>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Custom Thresholds</h3>
                            <p className="text-gray-600 mb-3">
                                Test with stricter matching criteria to see how it affects precision and recall.
                            </p>
                            <CodeBlock command>./cohort-bridge validate -config1 config_a.yaml -config2 config_b.yaml -ground-truth data/truth.csv -match-threshold 15 -jaccard-threshold 0.8</CodeBlock>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Verbose Analysis</h3>
                            <p className="text-gray-600 mb-3">
                                Get detailed breakdown of matches, misses, and false positives for in-depth analysis.
                            </p>
                            <CodeBlock command>./cohort-bridge validate -config1 config_a.yaml -config2 config_b.yaml -ground-truth data/truth.csv -verbose</CodeBlock>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Automated Validation</h3>
                            <p className="text-gray-600 mb-3">
                                Run validation in batch mode without confirmation prompts for automated testing.
                            </p>
                            <CodeBlock command>./cohort-bridge validate -config1 config_a.yaml -config2 config_b.yaml -ground-truth data/truth.csv -force</CodeBlock>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Complete Parameter Reference</h3>
                            <ParameterTable parameters={[
                                { name: '-config1', description: 'Configuration file for dataset 1 (Party A). Contains database path, fields, and matching settings.', default: 'prompted in interactive mode', color: 'bg-purple-100 text-purple-800' },
                                { name: '-config2', description: 'Configuration file for dataset 2 (Party B). Contains database path, fields, and matching settings.', default: 'prompted in interactive mode', color: 'bg-purple-100 text-purple-800' },
                                { name: '-ground-truth', description: 'CSV file with expected matches. Format: two columns with record IDs that should match.', default: 'prompted in interactive mode', color: 'bg-purple-100 text-purple-800' },
                                { name: '-output', description: 'Output CSV file for validation report with precision, recall, F1-score, and detailed match analysis.', default: 'auto-generated from config names', color: 'bg-purple-100 text-purple-800' },
                                { name: '-match-threshold', description: 'Hamming distance threshold for considering records a match. Lower values = stricter matching.', default: '20', color: 'bg-purple-100 text-purple-800' },
                                { name: '-jaccard-threshold', description: 'Minimum Jaccard similarity (0.0-1.0) for considering records a match. Higher values = stricter matching.', default: '0.5', color: 'bg-purple-100 text-purple-800' },
                                { name: '-verbose', description: 'Include detailed analysis with lists of false positives, false negatives, and match breakdowns.', default: 'false', color: 'bg-purple-100 text-purple-800' },
                                { name: '-interactive', description: 'Force interactive mode even when parameters are provided. Useful for reviewing settings.', default: 'false', color: 'bg-purple-100 text-purple-800' },
                                { name: '-force', description: 'Skip confirmation prompts and run automatically. Required for automated testing scripts.', default: 'false', color: 'bg-purple-100 text-purple-800' },
                                { name: '-help', description: 'Show detailed help message with examples and ground truth file format requirements.', default: 'false', color: 'bg-purple-100 text-purple-800' },
                            ]} />
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Validation Metrics</h3>
                            <p className="text-gray-600 mb-6">
                                The validation report includes comprehensive accuracy metrics:
                            </p>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <div className="space-y-4">
                                    {[
                                        { name: 'Precision', color: 'bg-green-100 text-green-800', desc: 'True matches / Total matches found' },
                                        { name: 'Recall', color: 'bg-blue-100 text-blue-800', desc: 'True matches / Total actual matches' },
                                        { name: 'F1-Score', color: 'bg-purple-100 text-purple-800', desc: 'Harmonic mean of precision and recall' },
                                    ].map((metric) => (
                                        <div key={metric.name} className="flex items-center space-x-3">
                                            <span className={`${metric.color} px-3 py-1 rounded text-sm font-semibold`}>
                                                {metric.name}
                                            </span>
                                            <span className="text-sm text-gray-600">{metric.desc}</span>
                                        </div>
                                    ))}
                                </div>
                                <div className="space-y-4">
                                    {[
                                        { name: 'False Positives', color: 'bg-red-100 text-red-800', desc: 'Incorrect matches found' },
                                        { name: 'False Negatives', color: 'bg-orange-100 text-orange-800', desc: 'Actual matches missed' },
                                        { name: 'True Positives', color: 'bg-gray-100 text-gray-800', desc: 'Correct matches found' },
                                    ].map((metric) => (
                                        <div key={metric.name} className="flex items-center space-x-3">
                                            <span className={`${metric.color} px-3 py-1 rounded text-sm font-semibold`}>
                                                {metric.name}
                                            </span>
                                            <span className="text-sm text-gray-600">{metric.desc}</span>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </div>
                );

            case 'configuration':
                return (
                    <div className="space-y-8">
                        <div>
                            <h2 className="text-3xl font-bold text-gray-900 mb-4">Configuration Files</h2>
                            <p className="text-lg text-gray-600 mb-6">
                                CohortBridge uses YAML configuration files to define data sources, matching parameters, and network settings.
                                Use our <a href="/config" className="text-blue-600 hover:text-blue-800 underline">Configuration Builder</a> to create these visually.
                            </p>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Configuration Types</h3>
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                {[
                                    { name: 'Basic Configuration', file: 'config.example.yaml', desc: 'Simple two-party record linkage', color: 'from-slate-50 to-slate-100' },
                                    { name: 'Secure Configuration', file: 'config_secure.example.yaml', desc: 'Enhanced security with logging', color: 'from-emerald-50 to-emerald-100' },
                                    { name: 'Tokenized Configuration', file: 'config_tokenized.example.yaml', desc: 'For pre-tokenized data', color: 'from-purple-50 to-purple-100' },
                                    { name: 'PostgreSQL Configuration', file: 'config_postgres.example.yaml', desc: 'Database integration', color: 'from-blue-50 to-blue-100' },
                                ].map((config) => (
                                    <div key={config.name} className={`bg-gradient-to-br ${config.color} border border-gray-200 rounded-lg p-6 hover:shadow-md transition-all`}>
                                        <h4 className="font-semibold text-gray-900 mb-2">{config.name}</h4>
                                        <p className="text-sm text-gray-600 mb-3">{config.desc}</p>
                                        <CodeBlock>{config.file}</CodeBlock>
                                    </div>
                                ))}
                            </div>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Key Configuration Sections</h3>
                            <div className="space-y-6">
                                <div className="border border-gray-200 rounded-lg p-6 bg-white hover:border-gray-300 transition-all">
                                    <h4 className="font-semibold text-gray-900 mb-4 flex items-center">
                                        <Database className="h-5 w-5 mr-2 text-gray-600" />
                                        Database Section
                                    </h4>
                                    <CodeBlock>
                                        {`database:
  filename: "data/patients.csv"
  fields:
    - name:FIRST
    - name:LAST
    - date:BIRTHDATE
    - zip:ZIP`}
                                    </CodeBlock>
                                </div>
                                <div className="border border-gray-200 rounded-lg p-6 bg-white hover:border-gray-300 transition-all">
                                    <h4 className="font-semibold text-gray-900 mb-4 flex items-center">
                                        <Settings className="h-5 w-5 mr-2 text-gray-600" />
                                        Matching Thresholds
                                    </h4>
                                    <CodeBlock>
                                        {`matching:
  hamming_threshold: 20      # Lower = stricter matching
  jaccard_threshold: 0.7     # Higher = stricter matching
  qgram_threshold: 0.8       # N-gram similarity`}
                                    </CodeBlock>
                                </div>
                                <div className="border border-gray-200 rounded-lg p-6 bg-white hover:border-gray-300 transition-all">
                                    <h4 className="font-semibold text-gray-900 mb-4 flex items-center">
                                        <Network className="h-5 w-5 mr-2 text-gray-600" />
                                        Network Settings
                                    </h4>
                                    <CodeBlock>
                                        {`listen_port: 8080
peer:
  host: "peer.example.com"
  port: 8080`}
                                    </CodeBlock>
                                </div>
                            </div>
                        </div>
                    </div>
                );

            case 'examples':
                return (
                    <div className="space-y-8">
                        <div>
                            <h2 className="text-3xl font-bold text-gray-900 mb-4">Common Examples</h2>
                            <p className="text-lg text-gray-600 mb-6">
                                Real-world examples and workflows for different use cases.
                            </p>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Complete Two-Party Workflow</h3>
                            <p className="text-gray-600 mb-6">
                                Here's how to set up a complete privacy-preserving record linkage between two parties:
                            </p>
                            <div className="space-y-6">
                                <div>
                                    <h4 className="font-semibold text-gray-800 mb-3">Party A (Receiver)</h4>
                                    <CodeBlock command>
                                        {`# 1. Create configuration for Party A
cp config.example.yaml config_a.yaml
# Edit config_a.yaml to set your data file and listen port

# 2. Run PPRL in interactive mode
./cohort-bridge pprl -config config_a.yaml -interactive`}
                                    </CodeBlock>
                                </div>
                                <div>
                                    <h4 className="font-semibold text-gray-800 mb-3">Party B (Sender)</h4>
                                    <CodeBlock command>
                                        {`# 1. Create configuration for Party B
cp config.example.yaml config_b.yaml
# Edit config_b.yaml to set data file and peer connection

# 2. Run PPRL to connect to Party A
./cohort-bridge pprl -config config_b.yaml`}
                                    </CodeBlock>
                                </div>
                            </div>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Enhanced Security Workflow</h3>
                            <p className="text-gray-600 mb-4">
                                For maximum security, tokenize your data in a secure environment first:
                            </p>
                            <CodeBlock command>
                                {`# 1. Tokenize in secure environment
./cohort-bridge tokenize \\
  -input data/sensitive_patients.csv \\
  -output secure_tokens.csv.enc \\
  -main-config config.yaml

# 2. Move encrypted tokens to less secure environment
# Transfer secure_tokens.csv.enc and secure_tokens.key

# 3. Use tokenized configuration
./cohort-bridge pprl -config config_tokenized.yaml`}
                            </CodeBlock>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Local Testing & Validation</h3>
                            <p className="text-gray-600 mb-4">
                                Test your setup locally before deploying to production:
                            </p>
                            <CodeBlock command>
                                {`# 1. Create test datasets
cp data/patients.csv data/dataset_a.csv
cp data/patients.csv data/dataset_b.csv
# Modify datasets to simulate real-world scenarios

# 2. Create ground truth
# Create expected_matches.csv with known matching pairs

# 3. Validate accuracy
./cohort-bridge validate \\
  -config1 config_a.yaml \\
  -config2 config_b.yaml \\
  -ground-truth expected_matches.csv \\
  -verbose`}
                            </CodeBlock>
                        </div>

                        <div>
                            <h3 className="text-xl font-semibold text-gray-900 mb-4">Database Integration</h3>
                            <p className="text-gray-600 mb-4">
                                Working with PostgreSQL databases for large-scale deployments:
                            </p>
                            <CodeBlock command>
                                {`# 1. Configure database connection
cp config_postgres.example.yaml config_db.yaml
# Edit database connection settings

# 2. Tokenize from database
./cohort-bridge tokenize \\
  -database \\
  -main-config config_db.yaml \\
  -output tokens.csv.enc

# 3. Run PPRL with database config
./cohort-bridge pprl -config config_db.yaml`}
                            </CodeBlock>
                        </div>
                    </div>
                );

            default:
                return null;
        }
    };

    return (
        <div className="min-h-screen bg-gradient-to-br from-slate-50 via-white to-slate-100/50">
            {/* Header */}
            <header className="border-b border-slate-200/60 bg-white/80 backdrop-blur-md sticky top-0 z-50">
                <div className="max-w-6xl mx-auto px-4">
                    <div className="flex justify-between items-center py-4">
                        <div className="flex items-center space-x-3">
                            <div className="w-8 h-8 bg-gradient-to-r from-blue-500 to-purple-500 rounded flex items-center justify-center">
                                <Settings className="h-4 w-4 text-white" />
                            </div>
                            <div>
                                <h1 className="text-lg font-bold text-gray-900">CohortBridge</h1>
                                <p className="text-xs text-gray-500">Privacy-Preserving Record Linkage</p>
                            </div>
                        </div>
                        <div className="flex items-center space-x-4">
                            <button
                                onClick={() => router.push('/')}
                                className="text-gray-600 hover:text-gray-900 transition-colors text-sm font-medium"
                            >
                                Home
                            </button>
                            <button
                                onClick={() => router.push('/config')}
                                className="text-gray-600 hover:text-gray-900 transition-colors text-sm font-medium"
                            >
                                Config Builder
                            </button>
                            <a
                                href="https://github.com/alunalabs/cohort-bridge"
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-gray-600 hover:text-gray-900 transition-colors"
                            >
                                <Github className="h-5 w-5" />
                            </a>
                        </div>
                    </div>
                </div>
            </header>

            {/* Hero Section */}
            <div className="bg-gradient-to-r from-blue-500 to-purple-500 text-white">
                <div className="max-w-6xl mx-auto px-4 py-16">
                    <div className="text-center">
                        <h2 className="text-4xl font-bold mb-4">
                            CohortBridge Documentation
                        </h2>
                        <p className="text-xl opacity-90 max-w-3xl mx-auto">
                            Complete guide to using the CohortBridge CLI tool for privacy-preserving record linkage.
                            Learn how to tokenize data, run peer-to-peer matching, and validate results.
                        </p>
                    </div>
                </div>
            </div>

            {/* Main Content */}
            <div className="max-w-6xl mx-auto px-4 py-12">
                {/* Tab Navigation */}
                <div className="mb-8">
                    <div className="border-b border-gray-200">
                        <nav className="-mb-px flex space-x-8 overflow-x-auto">
                            {tabs.map((tab) => {
                                const Icon = tab.icon;
                                return (
                                    <button
                                        key={tab.id}
                                        onClick={() => setActiveTab(tab.id)}
                                        className={`flex items-center space-x-2 py-4 px-1 border-b-2 font-medium text-sm whitespace-nowrap transition-colors ${activeTab === tab.id
                                            ? 'border-blue-500 text-blue-600'
                                            : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                                            }`}
                                    >
                                        <Icon className="h-4 w-4" />
                                        <span>{tab.name}</span>
                                    </button>
                                );
                            })}
                        </nav>
                    </div>
                </div>

                {/* Tab Content */}
                <div className="bg-white/70 backdrop-blur-sm rounded-xl shadow-lg border border-gray-200 p-8">
                    {renderTabContent()}
                </div>

                {/* Footer */}
                <div className="mt-16 pt-8 border-t border-gray-200">
                    <div className="text-center">
                        <h3 className="text-lg font-semibold text-gray-900 mb-4">Need Help?</h3>
                        <div className="flex justify-center space-x-6 text-sm">
                            <a href="https://github.com/alunalabs/cohort-bridge/issues" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:text-blue-800">
                                Report Issues
                            </a>
                            <a href="https://github.com/alunalabs/cohort-bridge/discussions" target="_blank" rel="noopener noreferrer" className="text-blue-600 hover:text-blue-800">
                                Discussions
                            </a>
                            <button onClick={() => router.push('/config')} className="text-blue-600 hover:text-blue-800">
                                Config Builder
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
} 