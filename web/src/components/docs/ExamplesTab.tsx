'use client';

import { Copy, Check } from 'lucide-react';
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

export default function ExamplesTab() {
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
} 