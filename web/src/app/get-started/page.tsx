'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { ArrowLeft, Github, Download, Terminal, Copy, Check, ExternalLink, GitBranch } from 'lucide-react';

export default function GetStartedPage() {
    const router = useRouter();
    const [copiedStep, setCopiedStep] = useState<string | null>(null);
    const [selectedOS, setSelectedOS] = useState<'windows' | 'macos' | 'linux'>('windows');

    const copyToClipboard = (text: string, stepId: string) => {
        navigator.clipboard.writeText(text);
        setCopiedStep(stepId);
        setTimeout(() => setCopiedStep(null), 2000);
    };

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

    const configTemplates = [
        {
            name: 'Basic CSV',
            description: 'Simple CSV record linkage',
            command: 'cohort-bridge init --template basic'
        },
        {
            name: 'PostgreSQL',
            description: 'Database connection setup',
            command: 'cohort-bridge init --template postgres'
        },
        {
            name: 'Secure Multi-Party',
            description: 'Enhanced privacy protocols',
            command: 'cohort-bridge init --template secure'
        },
        {
            name: 'Tokenized',
            description: 'Pre-tokenized data processing',
            command: 'cohort-bridge init --template tokenized'
        }
    ];

    return (
        <div className="min-h-screen bg-gray-50">
            {/* Header */}
            <header className="border-b border-gray-200 bg-white">
                <div className="max-w-6xl mx-auto px-4">
                    <div className="flex justify-between items-center py-4">
                        <button
                            onClick={() => router.push('/')}
                            className="flex items-center space-x-2 text-gray-600 hover:text-gray-900 transition-colors"
                        >
                            <ArrowLeft className="h-4 w-4" />
                            <span className="text-sm">Back</span>
                        </button>
                        <div className="flex items-center space-x-2">
                            <div className="w-6 h-6 bg-gradient-to-r from-blue-500 to-purple-500 rounded flex items-center justify-center">
                                <Terminal className="h-3 w-3 text-white" />
                            </div>
                            <div>
                                <h1 className="text-sm font-medium text-gray-900">Installation</h1>
                                <p className="text-xs text-gray-500">CLI Setup Guide</p>
                            </div>
                        </div>
                        <a
                            href="https://github.com/alunalabs/cohort-bridge#readme"
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-gray-600 hover:text-gray-900 transition-colors text-sm"
                        >
                            Documentation
                        </a>
                    </div>
                </div>
            </header>

            <div className="max-w-4xl mx-auto px-4 py-12">
                {/* Header */}
                <div className="mb-12">
                    <h2 className="text-2xl font-bold text-gray-900 mb-3">
                        Install CohortBridge CLI
                    </h2>
                    <p className="text-gray-600 max-w-2xl">
                        Download the CLI tool and create your first configuration.
                        Supports Windows, macOS, and Linux.
                    </p>
                </div>

                {/* Steps */}
                <div className="space-y-8">
                    {/* Step 1: GitHub */}
                    <div className="gradient-card rounded p-6 border">
                        <div className="flex items-start space-x-3">
                            <div className="w-8 h-8 bg-gradient-to-r from-blue-500 to-purple-500 rounded-full flex items-center justify-center flex-shrink-0">
                                <span className="text-sm font-bold text-white">1</span>
                            </div>
                            <div className="flex-1">
                                <div className="flex items-center space-x-2 mb-2">
                                    <h3 className="text-lg font-semibold text-gray-900">
                                        GitHub Repository
                                    </h3>
                                    <div className="bg-gradient-to-r from-blue-500 to-purple-500 text-white px-2 py-1 rounded-full text-xs font-medium">
                                        OPEN SOURCE
                                    </div>
                                </div>
                                <p className="text-gray-600 mb-4 text-sm">
                                    Access releases, documentation, source code, and community discussions.
                                </p>
                                <div className="flex flex-wrap gap-3">
                                    <a
                                        href="https://github.com/alunalabs/cohort-bridge"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className="inline-flex items-center space-x-2 bg-gray-900 text-white px-4 py-2 rounded text-sm hover:bg-gray-800 transition-colors font-medium"
                                    >
                                        <Github className="h-4 w-4" />
                                        <span>View Repository</span>
                                        <ExternalLink className="h-3 w-3" />
                                    </a>
                                    <a
                                        href="https://github.com/alunalabs/cohort-bridge/releases"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className="inline-flex items-center space-x-2 border border-gray-300 text-gray-700 px-4 py-2 rounded text-sm hover:border-gray-400 hover:text-gray-900 transition-colors bg-white"
                                    >
                                        <Download className="h-4 w-4" />
                                        <span>Latest Release</span>
                                    </a>
                                    <a
                                        href="https://github.com/alunalabs/cohort-bridge/issues"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className="inline-flex items-center space-x-2 border border-gray-300 text-gray-700 px-4 py-2 rounded text-sm hover:border-gray-400 hover:text-gray-900 transition-colors bg-white"
                                    >
                                        <GitBranch className="h-4 w-4" />
                                        <span>Issues & Support</span>
                                    </a>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Step 2: Install */}
                    <div className="gradient-card rounded p-6 border">
                        <div className="flex items-start space-x-3">
                            <div className="w-8 h-8 bg-gradient-to-r from-teal-500 to-blue-500 rounded-full flex items-center justify-center flex-shrink-0">
                                <span className="text-sm font-bold text-white">2</span>
                            </div>
                            <div className="flex-1">
                                <h3 className="text-lg font-semibold text-gray-900 mb-2">
                                    Install CLI Tool
                                </h3>
                                <p className="text-gray-600 mb-4 text-sm">
                                    Download binary for your platform.
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
                                <div className="space-y-3">
                                    {/* Download */}
                                    <div>
                                        <h4 className="text-sm font-medium text-gray-800 mb-1">Download</h4>
                                        <div className="bg-slate-50 border border-slate-200 rounded p-3 relative">
                                            <code className="text-slate-700 text-xs font-mono">
                                                {installCommands[selectedOS].download}
                                            </code>
                                            <button
                                                onClick={() => copyToClipboard(installCommands[selectedOS].download, `download-${selectedOS}`)}
                                                className="absolute top-2 right-2 p-1 hover:bg-slate-200 rounded transition-colors"
                                            >
                                                {copiedStep === `download-${selectedOS}` ? (
                                                    <Check className="h-3 w-3 text-green-600" />
                                                ) : (
                                                    <Copy className="h-3 w-3 text-gray-500" />
                                                )}
                                            </button>
                                        </div>
                                    </div>

                                    {/* Install */}
                                    <div>
                                        <h4 className="text-sm font-medium text-gray-800 mb-1">Install</h4>
                                        <div className="bg-slate-50 border border-slate-200 rounded p-3 relative">
                                            <code className="text-slate-700 text-xs font-mono">
                                                {installCommands[selectedOS].install}
                                            </code>
                                            <button
                                                onClick={() => copyToClipboard(installCommands[selectedOS].install, `install-${selectedOS}`)}
                                                className="absolute top-2 right-2 p-1 hover:bg-slate-200 rounded transition-colors"
                                            >
                                                {copiedStep === `install-${selectedOS}` ? (
                                                    <Check className="h-3 w-3 text-green-600" />
                                                ) : (
                                                    <Copy className="h-3 w-3 text-gray-500" />
                                                )}
                                            </button>
                                        </div>
                                    </div>

                                    {/* Verify */}
                                    <div>
                                        <h4 className="text-sm font-medium text-gray-800 mb-1">Verify Installation</h4>
                                        <div className="bg-slate-50 border border-slate-200 rounded p-3 relative">
                                            <code className="text-slate-700 text-xs font-mono">
                                                {installCommands[selectedOS].verify}
                                            </code>
                                            <button
                                                onClick={() => copyToClipboard(installCommands[selectedOS].verify, `verify-${selectedOS}`)}
                                                className="absolute top-2 right-2 p-1 hover:bg-slate-200 rounded transition-colors"
                                            >
                                                {copiedStep === `verify-${selectedOS}` ? (
                                                    <Check className="h-3 w-3 text-green-600" />
                                                ) : (
                                                    <Copy className="h-3 w-3 text-gray-500" />
                                                )}
                                            </button>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Step 3: Initialize */}
                    <div className="gradient-card rounded p-6 border">
                        <div className="flex items-start space-x-3">
                            <div className="w-8 h-8 bg-gradient-to-r from-purple-500 to-pink-500 rounded-full flex items-center justify-center flex-shrink-0">
                                <span className="text-sm font-bold text-white">3</span>
                            </div>
                            <div className="flex-1">
                                <h3 className="text-lg font-semibold text-gray-900 mb-2">
                                    Create Configuration
                                </h3>
                                <p className="text-gray-600 mb-4 text-sm">
                                    Generate a configuration file using built-in templates.
                                </p>

                                <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                                    {configTemplates.map((template, index) => (
                                        <div key={index} className="bg-white border border-gray-200 rounded p-3">
                                            <div className="flex justify-between items-start mb-2">
                                                <div>
                                                    <h4 className="text-sm font-medium text-gray-800">{template.name}</h4>
                                                    <p className="text-xs text-gray-600">{template.description}</p>
                                                </div>
                                                <button
                                                    onClick={() => copyToClipboard(template.command, `template-${index}`)}
                                                    className="p-1 hover:bg-gray-100 rounded transition-colors"
                                                >
                                                    {copiedStep === `template-${index}` ? (
                                                        <Check className="h-3 w-3 text-green-600" />
                                                    ) : (
                                                        <Copy className="h-3 w-3 text-gray-500" />
                                                    )}
                                                </button>
                                            </div>
                                            <code className="text-xs text-blue-600 font-mono bg-blue-50 px-2 py-1 rounded block">{template.command}</code>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Step 4: Run */}
                    <div className="gradient-card rounded p-6 border">
                        <div className="flex items-start space-x-3">
                            <div className="w-8 h-8 bg-gradient-to-r from-green-500 to-teal-500 rounded-full flex items-center justify-center flex-shrink-0">
                                <span className="text-sm font-bold text-white">4</span>
                            </div>
                            <div className="flex-1">
                                <h3 className="text-lg font-semibold text-gray-900 mb-2">
                                    Execute Matching
                                </h3>
                                <p className="text-gray-600 mb-4 text-sm">
                                    Run the privacy-preserving record linkage process.
                                </p>
                                <div className="bg-slate-50 border border-slate-200 rounded p-3 relative">
                                    <code className="text-slate-700 text-xs font-mono">
                                        cohort-bridge match --config config.yaml --input data.csv
                                    </code>
                                    <button
                                        onClick={() => copyToClipboard('cohort-bridge match --config config.yaml --input data.csv', 'run-match')}
                                        className="absolute top-2 right-2 p-1 hover:bg-slate-200 rounded transition-colors"
                                    >
                                        {copiedStep === 'run-match' ? (
                                            <Check className="h-3 w-3 text-green-600" />
                                        ) : (
                                            <Copy className="h-3 w-3 text-gray-500" />
                                        )}
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Next Steps */}
                <div className="mt-12 gradient-card rounded p-6 text-center border">
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">
                        Need a visual interface?
                    </h3>
                    <p className="text-gray-600 mb-4 text-sm">
                        Use our web-based configuration builder for complex setups.
                    </p>
                    <button
                        onClick={() => router.push('/config/basic')}
                        className="inline-flex items-center px-4 py-2 bg-gradient-to-r from-blue-500 to-purple-500 text-white rounded text-sm font-medium hover:from-blue-600 hover:to-purple-600 transition-all"
                    >
                        Open Config Builder
                    </button>
                </div>
            </div>
        </div>
    );
} 