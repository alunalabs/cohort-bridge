'use client';

import { Network, Lock, Search, Copy, Check } from 'lucide-react';
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

export default function OverviewTab() {
    const [selectedOS, setSelectedOS] = useState<'windows' | 'macos' | 'linux'>('windows');

    const installCommands = {
        windows: {
            download: 'curl -L "https://github.com/alunalabs/cohort-bridge/releases/download/v0.1.0/cohort-bridge-windows-amd64.exe" -o cohort-bridge.exe',
            install: `mkdir "%USERPROFILE%\\bin" >nul 2>&1 & move /Y cohort-bridge.exe "%USERPROFILE%\\bin\\" & setx PATH "%PATH%;%USERPROFILE%\\bin"`,
            verify: 'cohort-bridge --version'
        },
        macos: {
            download: 'curl -L "https://github.com/alunalabs/cohort-bridge/releases/download/v0.1.0/cohort-bridge-darwin-amd64" -o cohort-bridge',
            install: 'chmod +x cohort-bridge && sudo mv cohort-bridge /usr/local/bin/',
            verify: 'cohort-bridge --version'
        },
        linux: {
            download: 'wget https://github.com/alunalabs/cohort-bridge/releases/download/v0.1.0/cohort-bridge-linux-amd64',
            install: 'chmod +x cohort-bridge-linux-amd64 && sudo mv cohort-bridge-linux-amd64 /usr/local/bin/cohort-bridge',
            verify: 'cohort-bridge --version'
        }
    };

    return (
        <div className="space-y-8">
            <div>
                <h2 className="text-3xl font-bold text-gray-900 mb-4">Getting Started</h2>
                <p className="text-lg text-gray-600 mb-6">
                    CohortBridge enables direct peer-to-peer record linkage without third-party intermediaries.
                    Connect directly with partners to identify matching records while maintaining complete data privacy.
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
                        {selectedOS === 'windows' && (
                            <div className="mt-2 p-3 bg-yellow-50 border border-yellow-200 rounded">
                                <p className="text-sm text-yellow-800">
                                    <strong>Note:</strong> After installation, close and reopen your Command Prompt or PowerShell
                                    for the PATH changes to take effect.
                                </p>
                            </div>
                        )}
                    </div>

                    {/* Verify */}
                    <div>
                        <h4 className="text-sm font-medium text-gray-800 mb-2">Verify Installation</h4>
                        <CodeBlock command>{installCommands[selectedOS].verify}</CodeBlock>
                        {selectedOS === 'windows' && (
                            <div className="mt-2 p-3 bg-blue-50 border border-blue-200 rounded">
                                <p className="text-sm text-blue-800">
                                    <strong>Alternative:</strong> If you prefer not to modify PATH, you can run CohortBridge directly:<br />
                                    <code className="bg-blue-100 px-1 rounded text-xs">%USERPROFILE%\bin\cohort-bridge.exe --version</code>
                                </p>
                            </div>
                        )}
                    </div>
                </div>
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Configuration Files</h3>
                <p className="text-gray-600 mb-4">
                    Configuration files define how CohortBridge processes your data and connects with peers.
                    You can create these using the web interface or manually.
                </p>

                <div className="bg-gray-50 border border-gray-200 rounded-lg p-4 mb-4">
                    <h4 className="font-semibold text-gray-900 mb-2">Basic Configuration Structure</h4>
                    <CodeBlock>{`listen_port: 8080
database:
  type: csv
  filename: data/patients.csv
  fields:
    - name:first_name
    - name:last_name
    - date:date_of_birth
    - gender:gender
    - zip:zip_code
  random_bits_percent: 0
peer:
  host: localhost
  port: 8080`}</CodeBlock>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
                    <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                        <h4 className="font-semibold text-yellow-900 mb-2">Required Fields</h4>
                        <ul className="list-disc list-inside text-yellow-800 text-sm space-y-1">
                            <li><code>database.type</code> - The type of database you are using</li>
                            <li><code>database.fields</code> - Columns to match on and their types</li>
                            <li><code>peer.host</code> - Partner's IP address</li>
                            <li><code>peer.port</code> - Partner's port number</li>
                            <li><code>listen_port</code> - Your listening port</li>
                        </ul>
                    </div>
                </div>
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Connecting to Another Party</h3>
                <p className="text-gray-600 mb-4">
                    CohortBridge establishes direct peer-to-peer connections. One party acts as the server,
                    the other as the client. The process is automatic based on who starts first.
                </p>

                <div className="space-y-4 mb-6">
                    <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                        <h4 className="font-semibold text-blue-900 mb-2">Connection Process</h4>
                        <ol className="list-decimal list-inside text-blue-800 text-sm space-y-2">
                            <li><strong>Coordinate with partner:</strong> Exchange IP addresses and choose ports</li>
                            <li><strong>Configure both sides:</strong> Each party creates a config file pointing to the other</li>
                            <li><strong>Start the process:</strong> Both parties run CohortBridge with their configuration</li>
                            <li><strong>Automatic connection:</strong> CohortBridge will automatically establish a connection</li>
                            <li><strong>Secure exchange:</strong> Encrypted token exchange and matching</li>
                            <li><strong>Result validation:</strong> Compare results between parties to ensure accuracy</li>
                            <li><strong>Result export:</strong> Export results to a CSV file</li>
                        </ol>
                    </div>
                </div>
            </div>

            <div>
                <h3 className="text-xl font-semibold text-gray-900 mb-4">Interactive Mode</h3>
                <p className="text-gray-600 mb-4">
                    For guided setup, run CohortBridge without arguments to launch interactive mode:
                </p>
                <CodeBlock command>./cohort-bridge</CodeBlock>
                <p className="text-gray-600 mt-4">
                    Interactive mode walks you through configuration creation and connection setup.
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
} 