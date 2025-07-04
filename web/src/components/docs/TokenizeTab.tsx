'use client';

import { Shield, Copy, Check } from 'lucide-react';
import { useState } from 'react';

const CodeBlock = ({ children, command }: { children: string; command?: boolean }) => {
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
        <div className="bg-slate-50 border border-slate-200 rounded p-3 relative overflow-hidden">
            <div className="overflow-x-auto">
                <code className="text-slate-700 text-xs font-mono whitespace-pre block">
                    {children}
                </code>
            </div>
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

export default function TokenizeTab() {
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
} 